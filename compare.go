package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ninjawule/json-compare/utils"
)

func main() {
	// reading the arguments
	var one, two, idPropsString string

	var xml bool

	flag.StringVar(&one, "one", "", "required: the path to the first file to compare; must be a JSON file, or XML with the -xml option")
	flag.StringVar(&two, "two", "", "required: the path to the second file to compare; must be of the same first file's type")
	flag.BoolVar(&xml, "xml", false, "use this option if the files are XML files")
	flag.StringVar(&idPropsString, "idprops", "", "for array of objects, we need an identifying property for the objects, for sorting purposes amongst other things")
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

	// parsing the "order by" string
	idProps := map[string]string{}

	if idPropsString != "" {
		for _, idPropString := range strings.Split(idPropsString, ",") {
			idPropsElems := strings.Split(idPropString, ":")
			//nolint:gomnd
			if len(idPropsElems) != 2 {
				panic(fmt.Errorf("Error in the 'idprops' flag: '%s' does not respect the \">prop1>prop2>...>propN:idProp\n pattern, "+
					"to configure which object field should be used at a given path to uniquely identify the objects", idPropString))
			}

			idProps[idPropsElems[0]] = idPropsElems[1]
		}
	}

	// doing the comparison
	comparison, errComp := utils.CompareBytes(oneBytes, twoBytes, xml, idProps)
	if errComp != nil {
		panic(fmt.Errorf("Could not perform the comparison. Cause: %s", errComp))
	}

	// JSON-marshaling it
	comparisonBytes, errMarsh := json.MarshalIndent(comparison, "", "	")
	if errMarsh != nil {
		panic(fmt.Errorf("Error while JSON-marshaling the comparison. Cause: %s", errMarsh))
	}

	// outputting it
	if _, errWrite := os.Stdout.Write(comparisonBytes); errWrite != nil {
		panic(fmt.Errorf("Error while writing out the comparison. Cause: %s", errWrite))
	}
}
