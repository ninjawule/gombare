package core

import (
	"fmt"
	"io/ioutil"
	"path"
	"sort"
	"sync"
	"time"
)

//------------------------------------------------------------------------------
// Here we compare 2 folders
//------------------------------------------------------------------------------

// CompareFolders : getting a diff between 2 files, JSON or XML (for now)
//nolint:gocognit,gocyclo,cyclop
func CompareFolders(pathOne, pathTwo string, options *ComparisonOptions) (Comparison, error) {
	// lesssgooooo
	start := time.Now()

	// the result from comparing the 2 folders
	thisComparison := Comparison{}

	// listing the files within the 2 folders
	_, filesSliceOne, errList1 := getFiles(pathOne, options)
	filesMapTwo, filesSliceTwo, errList2 := getFiles(pathTwo, options)

	if errList1 != nil {
		return nil, errList1
	}

	if errList2 != nil {
		return nil, errList2
	}

	// first, let's keep track of the files we encounter
	checked := map[string]bool{}

	// let's count the total number of different files in the union of the two folders
	nbFilesInitial := len(filesSliceOne)

	// we going to handle chunks in parallel
	chunkSize := nbFilesInitial / options.NParallel

	// we're going to have a little bit of synchronization here
	mx := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	wg.Add(options.NParallel)

	// we keep a count of the files we're handling here
	//nolint
	nbFilesCounted := 0

	// let's gather the errors in here
	var errors []error

	// let's create as many Go routines as desired
	for chunkID := 1; chunkID <= options.NParallel; chunkID++ {
		go func(chunkID int) {
			defer wg.Done()

			// we're dealing with files with index from chunkID*chunkSize to (chunkID+1)*chunkSize - 1 (which size equals: chunkSize)
			limit := chunkID * chunkSize

			// but the last chunk will contain a few more elements
			if chunkID == options.NParallel {
				limit = limit + (nbFilesInitial - options.NParallel*chunkSize)
			}

			// the number of files handled in this routine
			nbFilesCountedLocal := 0

			// going through the files in the first folder, and comparing with the ones in the second folder
			for fileNum := (chunkID - 1) * chunkSize; fileNum < limit; fileNum++ {
				// this is one more file
				nbFilesCountedLocal++

				// we just need the name of the file we're handling here
				fileName1 := filesSliceOne[fileNum]

				// the comparison object potentially showing some diffs here
				var compFile1File2 Comparison

				// does this file exist in the 2nd folder ?
				if !filesMapTwo[fileName1] {
					// nope, fileName1 cannot be found in the 2nd folder
					compFile1File2 = one_two(pathOne, "-")

					// bit of logging
					// if !options.Silent {
					// 	options.Logger.Info("File '%s' only exists in dir one!", fileName1)
					// }

				} else {
					// if !options.Silent {
					// 	options.Logger.Info("Just started comparing files: %s", fileName1)
					// }

					// yes, the file exists, so we can compare the 2 files
					var errComp error
					compFile1File2, errComp = CompareFiles(path.Join(pathOne, fileName1), path.Join(pathTwo, fileName1), options, false)

					// we've found an error
					if errComp != nil {
						mx.Lock()
						errors = append(errors, errComp)
						mx.Unlock()

						break
					}
				}

				// we're making a block here for the synchronization
				if true {
					// making sure we're not getting race conditions
					mx.Lock()

					// this file is being checked
					checked[fileName1] = true

					// adding the diffs, if any
					if compFile1File2.hasDiffs() {
						if !options.StopAtFirst || len(thisComparison) == 0 {
							thisComparison[fileName1] = compFile1File2
						}
					}

					// let's release the lock now
					mx.Unlock()

					// if required, we stop here
					if compFile1File2.hasDiffs() && options.StopAtFirst {
						break
					}
				}
			}

			mx.Lock()
			nbFilesCounted = nbFilesCounted + nbFilesCountedLocal
			options.Logger.Info("Just finished comparing %d files", nbFilesCountedLocal)
			mx.Unlock()
			//
		}(chunkID)
	}

	// we're waiting here for all the routines to be done
	wg.Wait()

	// have we forgotten a file ?
	if nbFilesCounted != nbFilesInitial { // this should never happen
		panic(fmt.Sprintf("Had %d files, but handled %d files", nbFilesInitial, nbFilesCounted))
	}

	// now let's iterate over the second folder, because there might be stuff not found in the first folder
	for _, fileName2 := range filesSliceTwo {
		// we're considering files that have not been checked yet
		if !checked[fileName2] {
			// that's one more file
			nbFilesInitial++
			nbFilesCounted++

			// applying the 'StopAtFirst' option if needed
			if !options.StopAtFirst || len(thisComparison) == 0 {
				// this is a file that exists in the 2nd folder and not the first, so:
				thisComparison[fileName2] = one_two("-", pathTwo)

				// bit of logging
				if !options.Silent {
					options.Logger.Info("File '%s' only exists in dir two!", fileName2)
				}
			}

			if options.StopAtFirst {
				break
			}
		}
	}

	if !options.Silent {
		options.Logger.Info("Finished comparing the two folders in %s; %d diffs over %d files", time.Since(start), len(thisComparison), nbFilesInitial)
	}

	return thisComparison, nil
}

// returns a directory's list of files as a map
func getFiles(path string, options *ComparisonOptions) (map[string]bool, []string, error) {
	// we'll use the filenames as keys
	filesMap := map[string]bool{}
	filesSlice := []string{}

	// reading the current path
	fileInfos, errRead := ioutil.ReadDir(path)
	if errRead != nil {
		return nil, nil, fmt.Errorf("Error while listing files at path '%s'. Cause: %s", path, errRead)
	}

	// let's list the files
	for _, fileInfo := range fileInfos {
		if filename := fileInfo.Name(); !options.Ignored[filename] {
			filesMap[filename] = true

			filesSlice = append(filesSlice, filename)
		}
	}

	// let's sort the file names
	sort.Strings(filesSlice)

	return filesMap, filesSlice, nil
}
