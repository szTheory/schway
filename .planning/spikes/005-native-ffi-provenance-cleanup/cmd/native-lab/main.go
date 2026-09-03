package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"example.com/ai-lang/spike005/lab"
)

func main() {
	root := flag.String("root", ".", "spike root")
	pretty := flag.Bool("pretty", true, "pretty-print JSON")
	flag.Parse()
	absolute, err := filepath.Abs(*root)
	if err != nil {
		fail(err)
	}
	report, err := lab.Run(absolute)
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
