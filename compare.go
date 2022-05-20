package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/ninjawule/json-compare/utils"
)

func main() {

	// reading the arguments
	var one string
	var two string
	var xml bool
	flag.StringVar(&one, "one", "", "required: the path to the first file to compare; must be a JSON file, or XML with the -xml option")
	flag.StringVar(&two, "two", "", "required: the path to the second file to compare; must be of the same first file's type")
	flag.BoolVar(&xml, "xml", false, "use this option if the files are XML files")
	flag.Parse()

	// controlling their presence
	if one == "" || two == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// reading the files
	oneBytes, errOne := os.ReadFile(one)
	if errOne != nil {
		panic(errOne)
	}

	twoBytes, errTwo := os.ReadFile(two)
	if errTwo != nil {
		panic(errTwo)
	}

	// doing the comparison
	comparison, errComp := utils.CompareBytes(oneBytes, twoBytes, xml)
	if errComp != nil {
		panic(fmt.Errorf("Could not perform the comparison. Cause: %s", errComp))
	}

	// JSON-marshalling it
	comparisonBytes, errMarsh := json.MarshalIndent(comparison, "", "	")
	if errMarsh != nil {
		panic(fmt.Errorf("Error while JSON-marshalling the comparison. Cause: %s", errMarsh))
	}

	// outputting it
	if _, errWrite := os.Stdout.Write(comparisonBytes); errWrite != nil {
		panic(fmt.Errorf("Error while writing out the comparison. Cause: %s", errWrite))
	}
}
