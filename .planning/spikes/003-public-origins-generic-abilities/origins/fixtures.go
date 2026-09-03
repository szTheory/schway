package origins

import (
	"encoding/json"
	"fmt"
	"os"
)

func LoadFixtures(path string) (FixtureFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return FixtureFile{}, err
	}
	var file FixtureFile
	if err := json.Unmarshal(data, &file); err != nil {
		return FixtureFile{}, err
	}
	if file.Schema != 1 {
		return FixtureFile{}, fmt.Errorf("fixture schema %d is unsupported", file.Schema)
	}
	return file, nil
}

func FunctionMap(functions []Function) (map[string]Function, error) {
	result := make(map[string]Function, len(functions))
	for _, function := range functions {
		if function.ID == "" {
			return nil, fmt.Errorf("function id is empty")
		}
		if _, exists := result[function.ID]; exists {
			return nil, fmt.Errorf("duplicate function %q", function.ID)
		}
		result[function.ID] = function
	}
	return result, nil
}
