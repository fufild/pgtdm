package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/fufild/pgtdm/internal/anonymizer"
	"github.com/fufild/pgtdm/internal/file"
	"github.com/fufild/pgtdm/internal/logging"
)

func main() {
	inputFile := flag.String("input", "", "Input file")
	outputFile := flag.String("output", "./output.txt", "Output file")
	rulesFile := flag.String("rules", "./rules.json", "Path to JSON rules file")
	mapFile := flag.String("map", "", "Path to a persistent substitution map (JSON). Loaded before and saved after the run, so repeated runs map the same originals to the same anonymized values (preserves cross-database joins).")
	flag.Parse()

	if len(*inputFile) == 0 {
		logging.FatalError("Invalid input file. Specify with the --file flag")
	}

	// Load rules
	rules, err := loadRules(*rulesFile)
	if err != nil {
		logging.FatalError("Failed to load rules from JSON file: %v", err)
	}

	file, fileErr := file.NewFile(*inputFile)
	if fileErr != nil {
		logging.FatalError("Failed to open input file: %v", fileErr)
	}
	defer file.Close()

	anonymizer := anonymizer.NewAnonymizer(file, *outputFile, rules)
	if err := anonymizer.LoadSubstitutions(*mapFile); err != nil {
		logging.FatalError("Failed to load substitution map: %v", err)
	}
	if err := anonymizer.Anonymize(); err != nil {
		logging.FatalError("Anonymization failed: %v", err)
	}
	if err := anonymizer.SaveSubstitutions(*mapFile); err != nil {
		logging.FatalError("Failed to save substitution map: %v", err)
	}
	fmt.Printf("Data anonymized successfully.\nAnonymized %d values to %s\n", anonymizer.AnonymizedValueCount, anonymizer.OutputPath)
}

func loadRules(rulesFile string) (map[string][]string, error) {
	file, err := os.Open(rulesFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var rules map[string][]string
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&rules); err != nil {
		return nil, err
	}
	return rules, nil
}
