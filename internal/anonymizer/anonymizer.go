package anonymizer

import (
	"bufio"
	_ "embed"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fufild/pgtdm/internal/file"
	"github.com/fufild/pgtdm/internal/logging"
	"github.com/fufild/pgtdm/internal/random"
)

type Anonymizer struct {
	ColumnsToAnonymize   map[string][]string
	File                 *os.File
	OutputPath           string
	AnonymizedValueCount int
}

func NewAnonymizer(file *os.File, outputPath string, rules map[string][]string) *Anonymizer {
	return &Anonymizer{
		ColumnsToAnonymize:   rules,
		File:                 file,
		OutputPath:           outputPath,
		AnonymizedValueCount: 0,
	}
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
				if values[index] != "\\N" { // The value is NULL, ignore it
					values[index], err = anonymizeValue(normalizedRule)
					if err != nil {
						return err
					}
					a.AnonymizedValueCount++
				}
			}
		}

		writer.WriteString(strings.Join(values, "\t") + "\n")
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

func anonymizeValue(column string) (string, error) {
	anonymizedValue := ""
	replacementType := strings.ToLower(strings.ReplaceAll(column, "_", ""))
	if strings.Contains(column, ":") {
		replacementType = strings.Split(replacementType, ":")[1]
	}
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
		anonymizedValue = random.SSN()
	case "companyname":
		anonymizedValue = random.CompanyName()
	case "ein":
		anonymizedValue = random.EIN()
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
