package cfg

import (
	"encoding/json"
	"fmt"
	"os"
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
