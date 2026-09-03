package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"example.com/ai-lang/spike004/experiment"
)

func main() {
	fixture := flag.String("fixture", "fixtures/cases.json", "typed-core fixture")
	pretty := flag.Bool("pretty", true, "pretty-print JSON")
	flag.Parse()
	artifact, err := experiment.LoadFixture(*fixture)
	if err != nil {
		fail(err)
	}
	report, err := experiment.Run(artifact)
	if err != nil {
		fail(err)
	}
	var data []byte
	if *pretty {
		data, err = json.MarshalIndent(report, "", "  ")
	} else {
		data, err = json.Marshal(report)
	}
	if err != nil {
		fail(err)
	}
	fmt.Println(string(data))
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
