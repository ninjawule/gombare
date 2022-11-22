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
	FileType       FileType                 // the type of the files we're comparing
	IdParamsString string                   // a JSON representation of a IdentificationParameter parameter; can be the path to an existing JSON file
	IdParams       *IdentificationParameter // the properties (values of the map) serving as unique IDs for given paths (keys of the map)
	Check          bool                     // if true, then the ID params are output to allow for some checks
	Fast           bool                     // if true, then, in an array, an object's index is used as its IDProp, if none is specified for its path in the data tree; i.e. the IDProp `#index` is used, instead of nothing
	Silent         bool                     // if true, then no info / warning message is written out
	StopAtFirst    bool                     // if true, then, when comparing folders, we stop at the first couple of files that differ
	Logger         Logger                   // a logger
	IgnoredString  string                   // the files to ignore, separated by a comma
	Ignored        map[string]bool          // the ignored files
	AllowRaw       bool                     // if true, then it's allowed to display the raw JSON entities as difference, when added or removed; else, a display template is required
	IsXml          bool                     // if true, then the compared files are XML files
	NParallel      int                      // the number of routines used at the same time when comparing several files at once (i.e. comparing folders)
}

func (thisComp *ComparisonOptions) GetFileType() FileType {
	return thisComp.FileType
}

func (thisComp *ComparisonOptions) SetLogger(logger Logger) *ComparisonOptions {
	// not allowing the logger to be changed
	if thisComp.Logger == nil {
		thisComp.Logger = logger
	}

	return thisComp
}

func (thisComp *ComparisonOptions) GetIdParams() *IdentificationParameter {
	return thisComp.IdParams
}

// Resolve allows to transform some of the options, so as to make them usable by the comparison functions
func (thisComp *ComparisonOptions) Resolve() {
	thisComp.IdParams = thisComp.getIdParamsFromString()
	thisComp.Ignored = thisComp.getIgnoredFiles()
	thisComp.FileType = FileTypeJSON

	if thisComp.IsXml {
		thisComp.FileType = FileTypeXML
	}
}

func (thisComp *ComparisonOptions) getIgnoredFiles() map[string]bool {
	result := map[string]bool{}

	for _, ignored := range strings.Split(thisComp.IgnoredString, ",") {
		result[ignored] = true
	}

	return result
}

func (thisComp *ComparisonOptions) getIdParamsFromString() *IdentificationParameter {
	if thisComp.IdParamsString == "" {
		panic("no ID params!")
	}

	// at first, we suppose the whole JSON string has been provided
	idParamsJsonString := thisComp.IdParamsString

	// but what if it's the path to an existing file ?
	if _, errExist := os.Stat(thisComp.IdParamsString); errExist == nil {
		fileBytes, errRead := os.ReadFile(thisComp.IdParamsString)
		if errRead != nil {
			panic(fmt.Sprintf("error while readling config file (%s). Cause: %s", thisComp.IdParamsString, errRead))
		}

		idParamsJsonString = string(fileBytes)
	}

	param := &IdentificationParameter{}

	if err := json.Unmarshal([]byte(idParamsJsonString), param); err != nil {
		panic(fmt.Errorf("not a valid JSON (%s)", err))
	}

	if err := param.Resolve(thisComp.Check); err != nil {
		panic(fmt.Errorf("not a valid ID parameter: %s", err))
	}

	return param
}
