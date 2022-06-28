package core

import (
	"fmt"
	"os"
)

//------------------------------------------------------------------------------
// Here we compare 2 files
//------------------------------------------------------------------------------

// CompareFiles : getting a diff between 2 files, JSON or XML (for now)
func CompareFiles(filepathOne, filepathTwo string, options *ComparisonOptions) (Comparison, error) {
	// reading the files
	oneBytes, errOne := os.ReadFile(filepathOne)
	if errOne != nil {
		panic(fmt.Sprintf("Error while readling file one (%s). Cause: %s", filepathOne, errOne))
	}

	twoBytes, errTwo := os.ReadFile(filepathTwo)
	if errTwo != nil {
		panic(fmt.Sprintf("Error while readling file two (%s). Cause: %s", filepathTwo, errTwo))
	}

	// doing the comparison
	return compareBytes(oneBytes, twoBytes, options)
}
