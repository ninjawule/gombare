package core

import (
	"fmt"
	"os"
)

//------------------------------------------------------------------------------
// Here we compare 2 files
//------------------------------------------------------------------------------

// CompareFiles : getting a diff between 2 files, JSON or XML (for now)
func CompareFiles(filepathOne, filepathTwo string, options *ComparisonOptions, doLog bool) (Comparison, error) {
	// reading the files
	if doLog {
		options.logger.Info("Reading the first file")
	}

	oneBytes, errOne := os.ReadFile(filepathOne)
	if errOne != nil {
		panic(fmt.Sprintf("Error while readling file one (%s). Cause: %s", filepathOne, errOne))
	}

	if doLog {
		options.logger.Info("Reading the second file")
	}

	twoBytes, errTwo := os.ReadFile(filepathTwo)
	if errTwo != nil {
		panic(fmt.Sprintf("Error while readling file two (%s). Cause: %s", filepathTwo, errTwo))
	}

	// doing the comparison
	if doLog {
		options.logger.Info("Done reading the two files")
	}

	return compareBytes(oneBytes, twoBytes, options, doLog)
}
