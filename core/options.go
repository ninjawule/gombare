package core

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
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
	ignored     map[string]bool          // the ignored files
	allowRaw    bool                     // if true, then it's allowed to display the raw JSON entities as difference, when added or removed; else, a display template is required
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
func NewOptions(isXml bool, idParamsString string, fast bool, silent bool, ignoreString string, stopAtFirst bool, check bool, allowRaw bool) *ComparisonOptions {
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
		ignored:     getIgnoredFiles(ignoreString),
		allowRaw:    allowRaw,
	}
}

func getIgnoredFiles(ignoreString string) map[string]bool {
	result := map[string]bool{}

	for _, ignored := range strings.Split(ignoreString, ",") {
		result[ignored] = true
	}

	return result
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
