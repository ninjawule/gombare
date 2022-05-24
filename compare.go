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
	var one, two, idPropsString, outdir string

	var xml, split bool

	flag.StringVar(&one, "one", "", "required: the path to the first file to compare; must be a JSON file, or XML with the -xml option")
	flag.StringVar(&two, "two", "", "required: the path to the second file to compare; must be of the same first file's type")
	flag.BoolVar(&xml, "xml", false, "use this option if the files are XML files")
	flag.StringVar(&idPropsString, "idprops", "", "for array of objects, we need an identifying property for the objects, for sorting purposes amongst other things")
	flag.StringVar(&outdir, "outdir", "", "when specified, the result is written out as a JSON into this specified output directory")
	flag.BoolVar(&split, "split", false, "if 2 folders are compared, and if -outpir is used, then there's 1 comparison JSON produced for each pair of compared files")
	flag.Parse()

	// controlling their presence
	if one == "" || two == "" {
		flag.PrintDefaults()
		os.Exit(1)
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

	// checking the nature of the inputs
	oneDir := isDirectory(one)
	twoDir := isDirectory(two)

	if oneDir != twoDir {
		panic(fmt.Errorf("Cannot compare a file to a directory (one is directory: %t; two is a directory: %t)", oneDir, twoDir))
	}

	// the comparison result
	var comparison utils.Comparison

	var errComp error

	// comparing 2 files, or 2 folders
	if !oneDir {
		comparison, errComp = utils.CompareFiles(one, two, xml, idProps)
	} else {
		comparison, errComp = utils.CompareFolders(one, two, xml, idProps)
	}

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

// isDirectory determines if a file represented by `path` is a directory or not
func isDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		panic(fmt.Sprintf("Could not check wether '%s' is a directory or not. Cause: %s", path, err))
	}

	return fileInfo.IsDir()
}
