package anonymizer

import (
	"bufio"
	"encoding/json"
	_ "embed"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fufild/pgtdm/internal/file"
	"github.com/fufild/pgtdm/internal/logging"
	"github.com/fufild/pgtdm/internal/random"
)

// postgresNull is the PostgreSQL COPY marker for a NULL value. Strategies that
// remove a value rather than substitute it (e.g. ssn) write this so there is no
// ambiguity about whether the field was anonymized or left intact.
const postgresNull = "\\N"

type Anonymizer struct {
	ColumnsToAnonymize   map[string][]string
	File                 *os.File
	OutputPath           string
	AnonymizedValueCount int
	// substitutions caches the anonymized value chosen for each original value,
	// keyed by "<strategy>\x00<original>". This guarantees referential
	// integrity: the same source value (e.g. a company id shared by every user
	// in that company) always maps to the same anonymized value, so joins on
	// anonymized columns still line up.
	substitutions map[string]string
}

func NewAnonymizer(file *os.File, outputPath string, rules map[string][]string) *Anonymizer {
	return &Anonymizer{
		ColumnsToAnonymize:   rules,
		File:                 file,
		OutputPath:           outputPath,
		AnonymizedValueCount: 0,
		substitutions:        map[string]string{},
	}
}

// LoadSubstitutions seeds the substitution cache from a previously saved map file so a
// later run reuses the same anonymized value for the same original. This is what keeps
// cross-database joins intact (e.g. auth.identifiers.value <-> users.email/phone) when the
// app and auth dumps are anonymized in separate processes. A missing file is a no-op.
func (a *Anonymizer) LoadSubstitutions(path string) error {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read substitution map: %v", err)
	}
	if err := json.Unmarshal(data, &a.substitutions); err != nil {
		return fmt.Errorf("failed to parse substitution map: %v", err)
	}
	return nil
}

// SaveSubstitutions persists the substitution cache so subsequent runs stay consistent.
// A run that loads then saves the same file accumulates mappings across runs.
func (a *Anonymizer) SaveSubstitutions(path string) error {
	if path == "" {
		return nil
	}
	data, err := json.MarshalIndent(a.substitutions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode substitution map: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write substitution map: %v", err)
	}
	return nil
}

func (a *Anonymizer) Anonymize() error {
	defer file.CloseHandle(a.File) // Close the input when we're done

	outputFile, outputFileErr := file.CreateOutputFile(a.OutputPath)
	if outputFileErr != nil {
		return outputFileErr
	}
	defer file.CloseHandle(outputFile) // Close the output when we're done

	writer := bufio.NewWriter(outputFile)
	defer func() {
		if err := writer.Flush(); err != nil {
			logging.FatalError("Failed to flush writer: %v", err)
		}
	}()

	reader := bufio.NewReader(a.File)
	var currentTable string

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read string: %v", err)
		}
		if strings.HasPrefix(line, "CREATE TABLE") {
			currentTable = parseTableName(currentTable)
		}

		if strings.HasPrefix(line, "COPY") {
			table, columns := parseCopyHeader(line)
			if strings.HasPrefix(table, "_") {
				// Skip relationship tables
				continue
			}
			if rules, ok := a.ColumnsToAnonymize[table]; ok { // Check we have this table in our rules
				writer.WriteString(line) // Write the original COPY block
				currentTable = table     // Overwrite the current table
				err := a.processCopyBlock(reader, writer, columns, rules)
				if err != nil {
					return fmt.Errorf("failed to anonymize copy block: %v", err)
				}
				continue
			}
		}

		_, writerErr := writer.WriteString(line)
		if writerErr != nil {
			return fmt.Errorf("failed to write string: %v", writerErr)
		}

		if err == io.EOF {
			break
		}
	}
	return nil
}

func (a *Anonymizer) processCopyBlock(reader *bufio.Reader, writer *bufio.Writer, columns, rules []string) error {
	columnMap := map[string]int{}
	for i, col := range columns {
		normalizedColumnName := normalizeColumnName(col)
		columnMap[normalizedColumnName] = i
	}

	for {
		line, err := reader.ReadString('\n')
		if strings.TrimSpace(line) == "\\." { // End of the insert
			writer.WriteString(line)
			break
		}

		values := strings.Split(strings.TrimRight(line, "\n"), "\t") // Postgres separates by tab
		for _, rule := range rules {
			normalizedRule := normalizeColumnName(rule) // Incase we put some funky characters in the rule for some reason.
			normalizedRuleName := strings.Split(normalizedRule, ":")[0]
			if index, ok := columnMap[normalizedRuleName]; ok && index < len(values) {
				if values[index] != postgresNull { // The value is NULL, ignore it
					values[index], err = a.anonymizeValue(normalizedRule, values[index])
					if err != nil {
						return err
					}
					a.AnonymizedValueCount++
				}
			}
		}

		writer.WriteString(strings.Join(values, "\t"))
		writer.WriteString("\n")
		if err != nil {
			return err
		}
	}
	return nil
}

func parseTableName(line string) string {
	parts := strings.Fields(line)
	if len(parts) >= 3 {
		return strings.Trim(parts[2], `"`)
	}
	return ""
}

func parseCopyHeader(line string) (string, []string) {
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return "", nil
	}

	table := normalizeColumnName(parts[1])

	start := strings.Index(line, "(")
	end := strings.Index(line, ")")

	if start == -1 || end == -1 || end <= start {
		return table, nil // Table with no columns
	}

	columnSection := line[start+1 : end]
	columnSection = strings.ReplaceAll(columnSection, `"`, "") // Remove quotes
	columnList := strings.Split(columnSection, ",")

	for i, col := range columnList {
		columnList[i] = normalizeColumnName(col)
	}
	return table, columnList
}

func normalizeColumnName(column string) string {
	column = strings.TrimSpace(column)
	column = strings.ReplaceAll(column, `"`, "")
	column = strings.ReplaceAll(column, `\`, "")
	return column
}

// anonymizeValue returns the anonymized replacement for original, reusing a
// previously generated value when the same (strategy, original) pair has been
// seen before so that repeated source values map consistently.
func (a *Anonymizer) anonymizeValue(column, original string) (string, error) {
	replacementType := replacementTypeFor(column)
	// Polymorphic identifier columns (e.g. the auth DB's identifiers.value holds both
	// emails and phones) route by content to the email or phone strategy. This reuses the
	// exact same (strategy, original) cache key as users.email / users.phone in the app DB,
	// so with a shared substitution map the anonymized auth<->app join stays intact.
	if replacementType == "emailorphone" {
		if strings.Contains(original, "@") {
			replacementType = "email"
		} else {
			replacementType = "phone"
		}
	}
	cacheKey := replacementType + "\x00" + original
	if cached, ok := a.substitutions[cacheKey]; ok {
		return cached, nil
	}

	anonymizedValue, err := generateValue(replacementType)
	if err != nil {
		return "", err
	}

	a.substitutions[cacheKey] = anonymizedValue
	return anonymizedValue, nil
}

// replacementTypeFor extracts the masking strategy from a rule string,
// e.g. "work_phone:phone" -> "phone", "ssn" -> "ssn".
func replacementTypeFor(column string) string {
	replacementType := strings.ToLower(strings.ReplaceAll(column, "_", ""))
	if strings.Contains(column, ":") {
		replacementType = strings.Split(replacementType, ":")[1]
	}
	return replacementType
}

func generateValue(replacementType string) (string, error) {
	anonymizedValue := ""
	switch replacementType {
	default:
		return "", fmt.Errorf("unknown replacement type: %s", replacementType)
	case "address":
		randomAddress := random.Address()
		anonymizedValue = fmt.Sprintf("%s, %s, %s %s, USA", randomAddress[0], randomAddress[1], randomAddress[2], randomAddress[3])
	case "username":
		anonymizedValue = random.Username(8)
	case "firstname", "fname":
		anonymizedValue = random.FirstName()
	case "lastname", "lname":
		anonymizedValue = random.LastName()
	case "phone", "phonenumber":
		anonymizedValue = random.Phone()
	case "street":
		anonymizedValue = random.Address()[0]
	case "street2":
		anonymizedValue = random.Address()[4]
	case "city":
		anonymizedValue = random.Address()[1]
	case "state":
		anonymizedValue = random.Address()[2]
	case "zip":
		anonymizedValue = random.Address()[3]
	case "password", "pwd":
		anonymizedValue = random.Password()
	case "pin":
		anonymizedValue = random.Pin()
	case "ssn", "socialsecurity", "socialsecuritynumber":
		anonymizedValue = postgresNull // Remove SSNs entirely rather than substitute a fake one
	case "companyname":
		anonymizedValue = random.CompanyName()
	case "ein":
		anonymizedValue = random.EIN()
	case "routingnumber", "routing":
		anonymizedValue = random.RoutingNumber()
	case "bankname", "bank":
		anonymizedValue = random.BankName()
	case "uuid":
		anonymizedValue = random.UUID()
	case "ip", "ipaddr", "ipaddress":
		anonymizedValue = random.IPAddress()
	case "email":
		anonymizedValue = random.Email()
	}
	return anonymizedValue, nil
}
