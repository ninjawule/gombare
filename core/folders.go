package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"time"
)

//------------------------------------------------------------------------------
// Here we compare 2 folders
//------------------------------------------------------------------------------

// CompareFolders : getting a diff between 2 files, JSON or XML (for now)
func CompareFolders(pathOne, pathTwo string, options *ComparisonOptions) (Comparison, error) {
	// lesssgooooo
	start := time.Now()

	// the result from comparing the 2 folders
	thisComparison := map[string]interface{}{}

	// listing the files within the 2 folders
	filesOne, errList1 := listFilesToMap(pathOne)
	filesTwo, errList2 := listFilesToMap(pathTwo)

	if errList1 != nil {
		return nil, errList1
	}

	if errList2 != nil {
		return nil, errList2
	}

	// first, let's keep track of the files we encounter
	checked := map[string]bool{}

	// going through the files in the first folder, and comparing with the ones in the second folder
	for fileName1 := range filesOne {
		// this file is being checked
		checked[fileName1] = true

		// does this file exist in the 2nd folder ?
		if !filesTwo[fileName1] {
			// nope, fileName1 cannot be found in the 2nd folder
			thisComparison[fileName1] = one_two(pathOne, "-")

			// bit of logging
			if !options.silent {
				log.Printf("File '%s' only exists in dir one!", fileName1)
			}

			// if required, we stop here
			if options.stopAtFirst {
				goto END
			}

		} else {
			if !options.silent {
				log.Printf("Just started comparing files: %s", fileName1)
			}

			// yes, the file exists, so we can compare the 2 files
			compFile1File2, errComp := CompareFiles(path.Join(pathOne, fileName1), path.Join(pathTwo, fileName1), options)
			if errComp != nil {
				return nil, errComp
			}

			// adding only if there's at least 1 difference
			if compFile1File2.hasDiffs() {
				thisComparison[fileName1] = compFile1File2

				// if required, we stop here
				if options.stopAtFirst {
					goto END
				}
			}
		}
	}

	// now let's iterate over the second folder, because there might be stuff not found in the first folder
	for fileName2 := range filesTwo {
		// we're considering files that have not been checked yet
		if !checked[fileName2] {
			// this is a file that exists in the 2nd folder and not the first, so:
			thisComparison[fileName2] = one_two("-", pathTwo)

			// bit of logging
			if !options.silent {
				log.Printf("File '%s' only exists in dir two!", fileName2)
			}

			// if required, we stop here
			if options.stopAtFirst {
				goto END
			}
		}
	}

END:
	if !options.silent {
		log.Printf("Finished comparing the two folders in %s", time.Since(start))
	}

	return thisComparison, nil
}

// returns a directory's list of files as a map
func listFilesToMap(path string) (map[string]bool, error) {
	// we'll use the filenames as keys
	result := map[string]bool{}

	// reading the current path
	fileInfos, errRead := ioutil.ReadDir(path)
	if errRead != nil {
		return nil, fmt.Errorf("Error while listing file at path '%s'. Cause: %s", path, errRead)
	}

	// let's list the files
	for _, fileInfo := range fileInfos {
		result[fileInfo.Name()] = true
	}

	return result, nil
}
