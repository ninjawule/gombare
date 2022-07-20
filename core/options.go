package core

import (
	"encoding/json"
	"fmt"
	"os"
)

//------------------------------------------------------------------------------
// Comparison options
//------------------------------------------------------------------------------

type ComparisonOptions struct {
	fileType    FileType                 // the type of the files we're comparing
	idParams    *IdentificationParameter // the properties (values of the map) serving as unique IDs for given paths (keys of the map)
	fast        bool                     // if true, then, in an array, an object's index is used as its IDProp, if none is specified for its path in the data tree; i.e. the IDProp `#index` is used, instead of nothing
	silent      bool                     // if true, then no info / warning message is written out
	stopAtFirst bool                     // if true, then, when comparing folders, we stop at the first couple of files that differ
	logger      Logger                   // a logger
	// ignoredDups map[string]map[string]bool // the duplicate keys we won't report
	// autoIndex   bool                       // if true, then, in an array, an object's index is used as its IDProp, if none is specified for its path in the data tree; i.e. the IDProp `#index` is used, instead of nothing
}

func (thisComp *ComparisonOptions) GetFileType() FileType {
	return thisComp.fileType
}

func (thisComp *ComparisonOptions) SetLogger(logger Logger) {
	// not allowing the logger to be changed
	if thisComp.logger == nil {
		thisComp.logger = logger
	}
}

func (thisComp *ComparisonOptions) GetIdParams() *IdentificationParameter {
	return thisComp.idParams
}

// builds a new ComparisonOptions object
// func NewOptions(isXml bool, idParamsString string, autoIndex bool, fast bool, ignoreString string, silent bool, stopAtFirst bool) *ComparisonOptions {
func NewOptions(isXml bool, idParamsString string, fast bool, silent bool, stopAtFirst bool, check bool) *ComparisonOptions {
	fileType := FileTypeJSON
	if isXml {
		fileType = FileTypeXML
	}

	return &ComparisonOptions{
		fileType:    fileType,
		idParams:    getIdParamsFromString(idParamsString, check),
		fast:        fast,
		silent:      silent,
		stopAtFirst: stopAtFirst,
		// autoIndex:   autoIndex,
		// ignoredDups: getIgnoredDupsMap(ignoreString),
	}
}

func getIdParamsFromString(idParamsString string, check bool) *IdentificationParameter {
	// at first, we suppose the whole JSON string has been provided
	idParamsJsonString := idParamsString

	// but what if it's the path to an existing file ?
	if _, errExist := os.Stat(idParamsString); errExist == nil {
		fileBytes, errRead := os.ReadFile(idParamsString)
		if errRead != nil {
			panic(fmt.Sprintf("Error while readling config file (%s). Cause: %s", idParamsString, errRead))
		}

		idParamsJsonString = string(fileBytes)
	}

	param := &IdentificationParameter{}

	if err := json.Unmarshal([]byte(idParamsJsonString), param); err != nil {
		panic(fmt.Errorf("-idparams 2: Not a valid JSON (%s)", err))
	}

	if err := param.Resolve(check); err != nil {
		panic(fmt.Errorf("Not a valid ID parameter: %s", err))
	}

	return param
}

// func getIgnoredDupsMap(ignoreString string) map[string]map[string]bool {
// 	result := map[string]map[string]bool{}

// 	if ignoreString != "" {
// 		for _, ignoredPath := range strings.Split(ignoreString, ";") {
// 			keyAndValues := strings.Split(strings.TrimSpace(ignoredPath), ":")

// 			//nolint:gomnd
// 			if len(keyAndValues) != 2 {
// 				panic(fmt.Sprintf("this ignored path '%s' (in '%s') does not have the right format (key:val1~val2~etc)", ignoredPath, ignoreString))
// 			}

// 			values := map[string]bool{}
// 			for _, value := range strings.Split(keyAndValues[1], "~") {
// 				values[value] = true
// 			}

// 			result[keyAndValues[0]] = values
// 		}
// 	}

// 	return result
// }

// func (thisComp *ComparisonOptions) isIgnoredDuplicate(realPath, value string) bool {
// 	if ignoredPath := thisComp.ignoredDups[realPath]; ignoredPath != nil {
// 		return ignoredPath[value]
// 	}

// 	return false
// }
