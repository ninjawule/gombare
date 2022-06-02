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
	var one, two, idPropsString, outdir, orderByString string

	var xml, split, autoIndex, fast, silent, stopAtFirst bool

	flag.StringVar(&one, "one", "",
		"required: the path to the first file to compare; must be a JSON file, or XML with the -xml option")
	flag.StringVar(&two, "two", "",
		"required: the path to the second file to compare; must be of the same first file's type")
	flag.BoolVar(&xml, "xml", false,
		"use this option if the files are XML files")
	flag.StringVar(&idPropsString, "idprops", "",
		"for an array of objects, we need an identifying property for the objects, for sorting purposes amongst other things; "+
			"if '#index' is used as an ID, then that means that an object's index in the surrounding array is used as its ID; "+
			"example: \">path1>path2>path3:::propA+path4>propB as id3,>path1>path2>path3>id3>path5:::propC\"")
	flag.StringVar(&outdir, "outdir", "",
		"when specified, the result is written out as a JSON into this specified output directory")
	flag.BoolVar(&split, "split", false,
		"if 2 folders are compared, and if -outpir is used, then there's 1 comparison JSON produced for each pair of compared files")
	flag.BoolVar(&autoIndex, "autoIndex", false,
		"if true, then for array of objects with no id prop (cf. idprops option), the objects' indexes in the arrays are used as IDs")
	flag.BoolVar(&fast, "fast", false,
		"if true, then some verifications are not performed, like the uniqueness of IDs coming from the id props specified by the user; WARNING: this can lead to missing some differences!")
	flag.BoolVar(&silent, "silent", false,
		"if true, then no info / warning message is written out")
	flag.StringVar(&orderByString, "orderby", "",
		"for an array of objects that we cannot really define an ID property for, we want to sort the objects before comparing them with their index. The syntax is the same as for the -idprops option")
	flag.BoolVar(&stopAtFirst, "stopAtFirst", false,
		"if true, then, when comparing folders, we stop at the first couple of files that differ")

	flag.Parse()

	// controlling their presence
	if one == "" || two == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// the comparison options
	options := c.NewOptions(xml, idPropsString, autoIndex, orderByString, fast, silent, stopAtFirst)

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
		comparison, errComp = c.CompareFiles(one, two, options)
	} else {
		comparison, errComp = c.CompareFolders(one, two, options)
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
		panic(fmt.Errorf("Could not check wether '%s' is a directory or not. Cause: %s", path, err))
	}

	return fileInfo.IsDir()
}
