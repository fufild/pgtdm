package random

import (
	_ "embed"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/fufild/pgtdm/internal/logging"
	"github.com/gofrs/uuid"
)

//go:embed data/first_names.txt
var firstNames []byte

//go:embed data/last_names.txt
var lastNames []byte

//go:embed data/states.txt
var states []byte

var seedRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

const usernameCharset string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._"
const alphaCharset string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const numericCharset string = "0123456789"

func Username(length int) string {
	finalStr := make([]byte, length)
	for i := range finalStr {
		finalStr[i] = usernameCharset[seedRand.Intn(len(usernameCharset))]
	}
	return string(finalStr)
}

func FirstName() string {
	names := strings.Split(string(firstNames), "\n")
	name := names[seedRand.Intn(len(names))]
	return strings.ToUpper(name[:1]) + name[1:]
}

func LastName() string {
	names := strings.Split(string(lastNames), "\n")
	name := names[seedRand.Intn(len(names))]
	return strings.ToUpper(name[:1]) + name[1:]
}

func Phone() string {
	return fmt.Sprintf("+1-%03d-%03d-%04d", seedRand.Intn(900)+100, seedRand.Intn(900)+100, seedRand.Intn(10000))
}

func Address() []string {
	streetNumber := seedRand.Intn(9999) + 1
	streetName := fmt.Sprintf("%s Street", FirstName())
	city := fmt.Sprintf("%s City", LastName())

	// Random state
	allStates := strings.Split(string(states), "\n")
	state := allStates[seedRand.Intn(len(allStates))]
	state = strings.ToUpper(state)

	zip := fmt.Sprintf("%05d", seedRand.Intn(100000))

	return []string{
		fmt.Sprintf("%d %s", streetNumber, streetName),
		city,
		state,
		zip,
		"", // Address line 2
	}
}

func Password() string {
	return fmt.Sprintf("%s@%d", String(8), seedRand.Intn(9999))
}

func String(length int) string {
	finalStr := make([]byte, length)
	for i := range finalStr {
		finalStr[i] = alphaCharset[seedRand.Intn(len(alphaCharset))]
	}
	return string(finalStr)
}

func Number(length int) int {
	if length <= 0 {
		return 0
	}
	min := 1
	for i := 1; i < length; i++ {
		min *= 10
	}
	max := min*10 - 1
	return seedRand.Intn(max-min+1) + min
}

func Pin() string {
	return fmt.Sprintf("%04d", seedRand.Intn(9999)+1000)
}

func SSN() string {
	return fmt.Sprintf("%03d-%02d-%04d", seedRand.Intn(900)+100, seedRand.Intn(100), seedRand.Intn(10000))
}

func CompanyName() string {
	return fmt.Sprintf("%s Corp", LastName())
}

func EIN() string {
	return fmt.Sprintf("%02d-%07d", seedRand.Intn(90)+10, seedRand.Intn(10000000))
}

func BankName() string {
	banks := []string{
		"Bank of America",
		"Chase Bank",
		"Wells Fargo",
		"Citibank",
		"Capital One",
		"PNC Bank",
		"US Bank",
		"TD Bank",
		"Truist",
		"Legend",
	}
	return banks[seedRand.Intn(len(banks))]
}

func UUID() string {
	newId, err := uuid.NewV4()
	if err != nil {
		logging.FatalError("Failed to generate UUID: %v", err)
	}
	return newId.String()
}

func IPAddress() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		seedRand.Intn(256),
		seedRand.Intn(256),
		seedRand.Intn(256),
		seedRand.Intn(256),
	)
}

func Email() string {
	return fmt.Sprintf("tdm+%s@gmail.com", String(5))
}
