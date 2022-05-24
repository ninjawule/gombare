package utils

import (
	"fmt"
	"os"
)

//------------------------------------------------------------------------------
// Here we compare 2 files
//------------------------------------------------------------------------------

// CompareFiles : getting a diff between 2 files, JSON or XML (for now)
func CompareFiles(pathOne, pathTwo string, xml bool, idProps map[string]string) (Comparison, error) {
	// reading the files
	oneBytes, errOne := os.ReadFile(pathOne)
	if errOne != nil {
		panic(fmt.Sprintf("Error while readling file one (%s). Cause: %s", pathOne, errOne))
	}

	twoBytes, errTwo := os.ReadFile(pathTwo)
	if errTwo != nil {
		panic(fmt.Sprintf("Error while readling file two (%s). Cause: %s", pathTwo, errTwo))
	}

	// doing the comparison
	return compareBytes(oneBytes, twoBytes, xml, idProps)
}
