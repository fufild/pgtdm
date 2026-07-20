package anonymizer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/fufild/pgtdm/internal/file"
)

func newTestAnonymizer() *Anonymizer {
	return &Anonymizer{substitutions: map[string]string{}}
}

func TestAnonymizeValueIsConsistentForSameOriginal(t *testing.T) {
	// Arrange
	a := newTestAnonymizer()

	// Act
	first, firstErr := a.anonymizeValue("company_id:uuid", "company-original-123")
	second, secondErr := a.anonymizeValue("company_id:uuid", "company-original-123")

	// Assert
	assert.NoError(t, firstErr)
	assert.NoError(t, secondErr)
	assert.Equal(t, first, second, "the same original value must always map to the same anonymized value")
}

func TestAnonymizeValueDiffersByOriginal(t *testing.T) {
	// Arrange
	a := newTestAnonymizer()

	// Act
	companyA, _ := a.anonymizeValue("company_id:uuid", "company-A")
	companyB, _ := a.anonymizeValue("company_id:uuid", "company-B")

	// Assert
	assert.NotEqual(t, companyA, companyB, "different original values should map to different anonymized values")
}

func TestAnonymizeValueRoutingStrategy(t *testing.T) {
	// Arrange
	a := newTestAnonymizer()

	// Act
	routing, err := a.anonymizeValue("routing_number:routing", "021000021")

	// Assert
	assert.NoError(t, err)
	assert.Regexp(t, `^\d{9}$`, routing, "routing strategy should produce a 9-digit number")
}

func TestAnonymizeValueRemovesSSN(t *testing.T) {
	// Arrange
	a := newTestAnonymizer()

	// Act
	val, err := a.anonymizeValue("ssn", "123-45-6789")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, postgresNull, val, "ssn should be removed (set to NULL), not replaced with a fake")
}

func TestAnonymizeValueUnknownStrategy(t *testing.T) {
	// Arrange
	a := newTestAnonymizer()

	// Act
	_, err := a.anonymizeValue("mystery:notarealstrategy", "value")

	// Assert
	assert.Error(t, err, "an unknown strategy should return an error")
}

// Finch OAuth access tokens are stored in the shared public.tokens table
// (discriminated by an "application" column), not a Finch-specific table, so
// this drives the full Anonymize() COPY-block pipeline to prove the existing
// public.tokens rule actually catches them end to end.
func TestAnonymizeFullPipelineAnonymizesFinchAccessToken(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "dump.sql")
	outputPath := filepath.Join(dir, "output.sql")

	const originalAccessToken = "secret-finch-access-token"
	dump := "COPY public.tokens (id, user_id, application, scope, token_type, access_token, created_at, expires_in, refresh_token) FROM stdin;\n" +
		"tok-1\tuser-1\tfinch\tcompany_info\taccess_token\t" + originalAccessToken + "\t2026-01-01 00:00:00\t0\t\n" +
		"\\.\n"
	assert.NoError(t, os.WriteFile(inputPath, []byte(dump), 0o644))

	inputFile, err := file.NewFile(inputPath)
	assert.NoError(t, err)

	rules := map[string][]string{
		"public.tokens": {"access_token:uuid", "refresh_token:uuid"},
	}
	a := NewAnonymizer(inputFile, outputPath, rules)

	// Act
	anonymizeErr := a.Anonymize()

	// Assert
	assert.NoError(t, anonymizeErr)
	output, err := os.ReadFile(outputPath)
	assert.NoError(t, err)

	lines := strings.Split(string(output), "\n")
	row := strings.Split(lines[1], "\t")
	assert.Equal(t, "tok-1", row[0], "id should pass through unchanged")
	assert.Equal(t, "finch", row[2], "application discriminator should pass through unchanged")
	assert.NotEqual(t, originalAccessToken, row[5], "the finch access_token must be anonymized")
	assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, row[5], "the anonymized access_token should be a UUID, not empty or NULL")
}
