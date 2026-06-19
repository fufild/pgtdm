package anonymizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
