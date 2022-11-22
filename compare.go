package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	c "github.com/ninjawule/gombare/core"
)

func main() {
	// reading the arguments
	var one, two string

	// gathering the desired options
	opt := &c.ComparisonOptions{}

	flag.StringVar(&one, "one", "",
		"required: the path to the first file to compare; must be a JSON file, or XML with the -xml option")
	flag.StringVar(&two, "two", "",
		"required: the path to the second file to compare; must be of the same first file's type")
	flag.BoolVar(&opt.IsXml, "xml", false,
		"use this option if the files are XML files")
	flag.StringVar(&opt.IdParamsString, "idparams", "",
		"a JSON representation of a IdentificationParameter parameter; see the docs for an example; can be the path to an existing JSON file")
	// flag.StringVar(&opt.Outdir, "outdir", "",
	// 	"when specified, the result is written out as a JSON into this specified output directory")
	flag.BoolVar(&opt.Fast, "fast", false,
		"if true, then some verifications are not performed, like the uniqueness of IDs coming from the id props specified by the user; WARNING: this can lead to missing some differences!")
	flag.BoolVar(&opt.Silent, "silent", false,
		"if true, then no info / warning message is written out")
	flag.BoolVar(&opt.StopAtFirst, "stopAtFirst", false,
		"if true, then, when comparing folders, we stop at the first couple of files that differ")
	flag.BoolVar(&opt.Check, "check", false,
		"if true, then the ID params are output to allow for some checks")
	flag.StringVar(&opt.IgnoredString, "ignore", "",
		"the files to ignore, separated by a comma")
	flag.BoolVar(&opt.AllowRaw, "allowRaw", false,
		"if true, then it's allowed to display the raw JSON entities as difference, when added or removed; else, a display template is required")
	//nolint:revive,gomnd
	flag.IntVar(&opt.NParallel, "nparallel", 10,
		"the number of routines used at the same time when comparing several files at once (i.e. comparing folders)")

	flag.Parse()

	// controlling the presence of 2 things to compare
	if one == "" || two == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// let's set a logger, and "finalize" the options
	opt.SetDefaultLogger().Resolve()

	// are we just performing a check ?
	if opt.Check {
		doJsonOutput(opt.GetIdParams(), "the ID params")

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
		comparison, errComp = c.CompareFiles(one, two, opt, true)
	} else {
		comparison, errComp = c.CompareFolders(one, two, opt)
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
