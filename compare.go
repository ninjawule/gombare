package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	c "github.com/ninjawule/json-compare/core"
)

func main() {
	// reading the arguments
	// var one, two, idParamsString, outdir, ignoreString string
	var one, two, idParamsString, outdir string

	// var xml, split, autoIndex, fast, silent, stopAtFirst bool
	var xml, fast, silent, stopAtFirst, check bool

	flag.StringVar(&one, "one", "",
		"required: the path to the first file to compare; must be a JSON file, or XML with the -xml option")
	flag.StringVar(&two, "two", "",
		"required: the path to the second file to compare; must be of the same first file's type")
	flag.BoolVar(&xml, "xml", false,
		"use this option if the files are XML files")
	flag.StringVar(&idParamsString, "idparams", "",
		"a JSON representation of a IdentificationParameter parameter; see the docs for an example; can be the path to an existing JSON file")
	flag.StringVar(&outdir, "outdir", "",
		"when specified, the result is written out as a JSON into this specified output directory")
	// flag.BoolVar(&split, "split", false,
	// 	"if 2 folders are compared, and if -outpir is used, then there's 1 comparison JSON produced for each pair of compared files")
	// flag.BoolVar(&autoIndex, "autoIndex", false,
	// 	"if true, then for array of objects with no id prop (cf. idprops option), the objects' indexes in the arrays are used as IDs")
	flag.BoolVar(&fast, "fast", false,
		"if true, then some verifications are not performed, like the uniqueness of IDs coming from the id props specified by the user; WARNING: this can lead to missing some differences!")
	flag.BoolVar(&silent, "silent", false,
		"if true, then no info / warning message is written out")
	flag.BoolVar(&stopAtFirst, "stopAtFirst", false,
		"if true, then, when comparing folders, we stop at the first couple of files that differ")
	flag.BoolVar(&check, "check", false,
		"if true, then the ID params are output to allow for some checks")
	// flag.StringVar(&ignoreString, "ignore", "", "the ID keys that should be ignored when they're duplicates: key1:valA~valB~valC;key2:valD~valE")

	flag.Parse()

	// controlling their presence
	if one == "" || two == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// the comparison options
	// options := c.NewOptions(xml, idParamsString, autoIndex, fast, ignoreString, silent, stopAtFirst).SetDefaultLogger()
	options := c.NewOptions(xml, idParamsString, fast, silent, stopAtFirst, check).SetDefaultLogger()

	// are we just performing a check ?
	if check {
		doJsonOutput(options.GetIdParams(), "the ID params")

		return // we're out
	}

	// checking the nature of the inputs
	//nolint:ifshort
	oneDir := isDirectory(one)
	//nolint:ifshort
	twoDir := isDirectory(two)

	if oneDir != twoDir {
		panic(fmt.Errorf("Cannot compare a file to a directory (one is directory: %t; two is a directory: %t)", oneDir, twoDir))
	}

	// the comparison result
	var comparison c.Comparison

	var errComp error

	// comparing 2 files, or 2 folders
	if !oneDir {
		comparison, errComp = c.CompareFiles(one, two, options, true)
	} else {
		comparison, errComp = c.CompareFolders(one, two, options)
	}

	if errComp != nil {
		panic(fmt.Errorf("Could not perform the comparison. Cause: %s", errComp))
	}

	// outputting the comparison
	doJsonOutput(comparison, "the comparison")
}

// isDirectory determines if a file represented by `path` is a directory or not
func isDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		panic(fmt.Errorf("Could not check wether '%s' is a directory or not. Cause: %s", path, err))
	}

	return fileInfo.IsDir()
}

// outputting an object
func doJsonOutput(object interface{}, what string) {

	// JSON-marshaling it
	objectBytes, errMarsh := json.MarshalIndent(object, "", "	")
	if errMarsh != nil {
		panic(fmt.Errorf("Error while JSON-marshaling %s. Cause: %s", what, errMarsh))
	}

	// outputting it
	if _, errWrite := os.Stdout.Write(objectBytes); errWrite != nil {
		panic(fmt.Errorf("Error while writing out %s. Cause: %s", what, errWrite))
	}
}
