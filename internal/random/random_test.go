package random

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsername(t *testing.T) {
	username := Username(10)
	assert.Len(t, username, 10)
}

func TestFirstName(t *testing.T) {
	firstName := FirstName()
	assert.NotEmpty(t, firstName)
	assert.Regexp(t, `^[A-Z][a-z]+$`, firstName, "FirstName should start with an uppercase letter and contain lowercase letters.")
}

func TestLastName(t *testing.T) {
	lastName := LastName()
	assert.NotEmpty(t, lastName)
	assert.Regexp(t, `^[A-Z][a-z]+$`, lastName, "LastName should start with an uppercase letter and contain lowercase letters.")
}

func TestPhone(t *testing.T) {
	phone := Phone()
	assert.Regexp(t, `^\+1-\d{3}-\d{3}-\d{4}$`, phone, "Phone should match the expected format +1-XXX-XXX-XXXX.")
}

func TestAddress(t *testing.T) {
	address := Address()
	assert.Len(t, address, 5)
	assert.NotEmpty(t, address[0], "Street address should not be empty.")
	assert.NotEmpty(t, address[1], "City should not be empty.")
	assert.NotEmpty(t, address[2], "State should not be empty.")
	assert.Regexp(t, `^\d{5}$`, address[3], "ZIP code should be a 5-digit number.")
}

func TestPassword(t *testing.T) {
	password := Password()
	assert.Regexp(t, `^[a-zA-Z]+@\d+$`, password, "Password should have the format string@digits.")
}

func TestString(t *testing.T) {
	str := String(10)
	assert.Len(t, str, 10)
	assert.Regexp(t, `^[a-zA-Z]+$`, str, "String should only contain alphabetic characters.")
}

func TestNumber(t *testing.T) {
	number := Number(4)
	assert.GreaterOrEqual(t, number, 1000)
	assert.LessOrEqual(t, number, 9999)
}

func TestPin(t *testing.T) {
	pin := Pin()
	assert.Regexp(t, `^\d{4}$`, pin, "PIN should be a 4-digit number.")
}

func TestCompanyName(t *testing.T) {
	company := CompanyName()
	assert.NotEmpty(t, company)
	assert.Contains(t, company, " Corp", "Company name should end with 'Corp'.")
}

func TestEIN(t *testing.T) {
	ein := EIN()
	assert.Regexp(t, `^\d{2}-\d{7}$`, ein, "EIN should match the format XX-XXXXXXX.")
}

func TestBankName(t *testing.T) {
	bankName := BankName()
	assert.NotEmpty(t, bankName)
}

func TestUUID(t *testing.T) {
	uuid := UUID()
	assert.Regexp(t, `^[a-f0-9\-]+$`, uuid, "UUID should match the expected format.")
}

func TestRoutingNumber(t *testing.T) {
	routing := RoutingNumber()
	assert.Regexp(t, `^\d{9}$`, routing, "Routing number should be a 9-digit number.")

	d := make([]int, 9)
	for i, c := range routing {
		d[i] = int(c - '0')
	}
	weightedSum := 3*(d[0]+d[3]+d[6]) + 7*(d[1]+d[4]+d[7]) + 1*(d[2]+d[5]+d[8])
	assert.Equal(t, 0, weightedSum%10, "Routing number should have a valid ABA checksum.")
}

func TestIPAddress(t *testing.T) {
	ip := IPAddress()
	assert.Regexp(t, `^\d{1,3}(\.\d{1,3}){3}$`, ip, "IP address should match the format XXX.XXX.XXX.XXX.")
}

func TestEmail(t *testing.T) {
	email := Email()
	assert.Regexp(t, `^tdm\+\w+@gmail\.com$`, email, "Email should match the format tdm+<random_string>@gmail.com.")
}
