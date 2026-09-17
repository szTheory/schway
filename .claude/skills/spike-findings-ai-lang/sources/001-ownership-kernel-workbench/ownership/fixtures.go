package ownership

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

func LoadFixtures(path string) (FixtureFile, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return FixtureFile{}, err
	}
	var file FixtureFile
	if err := json.Unmarshal(bytes, &file); err != nil {
		return FixtureFile{}, err
	}
	if file.Schema != 1 {
		return FixtureFile{}, fmt.Errorf("unsupported fixture schema %d", file.Schema)
	}
	return file, nil
}

func LoadFlowFixtures(path string) (FlowFixtureFile, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return FlowFixtureFile{}, err
	}
	var file FlowFixtureFile
	if err := json.Unmarshal(bytes, &file); err != nil {
		return FlowFixtureFile{}, err
	}
	if file.Schema != 1 {
		return FlowFixtureFile{}, fmt.Errorf("unsupported flow fixture schema %d", file.Schema)
	}
	return file, nil
}

func Compare(program Program, options CheckerOptions) Comparison {
	oracle := RunOracle(program)
	checker := Check(program, options)
	agreement := oracle.Valid == checker.Valid && reflect.DeepEqual(oracle.Events, checker.Events)
	if !oracle.Valid && !checker.Valid {
		agreement = agreement && oracle.Diagnostic.Code == checker.Diagnostic.Code && oracle.Diagnostic.SourceIndex == checker.Diagnostic.SourceIndex
	}
	return Comparison{Program: program.ID, Agreement: agreement, Oracle: oracle, Checker: checker}
}
