package utils

import (
	"fmt"
	"strconv"
	"strings"
)

//------------------------------------------------------------------------------
// Here is the base for comparing stuff
//------------------------------------------------------------------------------

// yeah, let's make a recursive type... Why not ?
type Comparison map[string]interface{}

func (comp Comparison) hasDiffs() bool {
	return len(comp) > 0
}

func nodif() Comparison {
	return Comparison{}
}

func one(obj interface{}) Comparison {
	return map[string]interface{}{"_one_": obj}
}

func two(obj interface{}) Comparison {
	return map[string]interface{}{"_two_": obj}
}

func one_two(obj1, obj2 interface{}) Comparison {
	return map[string]interface{}{"_one_": obj1, "_two_": obj2}
}

//------------------------------------------------------------------------------
// Constants
//------------------------------------------------------------------------------

// the types for the files we're comparing
type FileType string

const (
	FileTypeJSON FileType = "JSON"
	FileTypeXML  FileType = "XML"
)

//------------------------------------------------------------------------------
// Identifying paths in a data tree
//------------------------------------------------------------------------------

// the path for any property in the data trees we consider
type PropPath string

func (thisPath PropPath) To(key string) PropPath {
	return PropPath(string(thisPath) + ">" + key)
}

func (thisPath PropPath) With(idProp *IDProp) PropPath {
	return PropPath(string(thisPath) + ">" + idProp.idPart())
}

//------------------------------------------------------------------------------
// Identifying objects in a data tree
//------------------------------------------------------------------------------

// the name of the property that should be used to uniquely identify an object within an array, at a given prop path;
// in some cases, the property can be composite, i.e. :
// - made to refer to a nested property (e.g. "contract>general>uid"), instead of a direct one ("contractRef")
// - a combination of several properties; e.g. "contract>general>uid+contract>creationDate"
type IDProp struct {
	from  PropPath     // the path for the objects this ID property should belong to
	props [][]PropPath // e.g. [ [contract, general, uid], [contract, creationDate] ], for "contract>general>uid+contract>creationDate"
	idStr string       // the "string" version, e.g. "contract>general>uid+contract>creationDate"
}

const (
	idPropINDEX PropPath = "#index"
)

func (thisProp *IDProp) idPart() string {
	// passing from [ [contract, general, uid], [contract, creationDate] ]
	// to "contract>general>uid+contract>creationDate"
	if thisProp.idStr == "" {
		paths := []string{}

		for _, pathElements := range thisProp.props {
			pathString := string(pathElements[0])
			for i := 1; i < len(pathElements); i++ {
				pathString = pathString + ">" + string(pathElements[i])
			}

			paths = append(paths, pathString)
		}

		thisProp.idStr = strings.Join(paths, "+")
	}

	return thisProp.idStr
}

func (thisProp *IDProp) toFullString() string {
	return fmt.Sprintf("%s:%s", thisProp.from, thisProp.idPart())
}

func (thisProp *IDProp) isIndex() bool {
	return thisProp.props[0][0] == idPropINDEX
}

func indexAsID(atPath PropPath) *IDProp {
	return &IDProp{
		from: atPath,
		props: [][]PropPath{
			{
				idPropINDEX,
			},
		},
		idStr: string(idPropINDEX),
	}
}

func (thisProp *IDProp) getValueForObj(obj map[string]interface{}) string {
	valuesForObj := []string{}

	for _, pathChain := range thisProp.props { // ranging over smth like this: [ [contract, general, uid], [contract, creationDate] ]
		currentObj := obj // starting from the "root" object

		for _, path := range pathChain { // ranging over [contract, general, uid]
			// getting the value at that path from the current object:
			switch value := currentObj[string(path)]; value.(type) {
			case float64:
				//nolint:errcheck
				floatValue := value.(float64)
				if floatValue == float64(int(floatValue)) {
					valuesForObj = append(valuesForObj, strconv.Itoa(int(floatValue)))
				} else {
					//nolint:revive, gomnd
					valuesForObj = append(valuesForObj, strconv.FormatFloat(floatValue, 'f', 6, 64))
				}
			case string:
				valuesForObj = append(valuesForObj, value.(string))
			case bool:
				if value.(bool) {
					valuesForObj = append(valuesForObj, "true")
				} else {
					valuesForObj = append(valuesForObj, "false")
				}
			case map[string]interface{}:
				// going deeper
				//nolint:errcheck
				currentObj = value.(map[string]interface{})
			default:
				panic(fmt.Errorf("Cannot handle the value (of type: %T) at path '%s' (which is part of this id property: %s)", value, path, thisProp.toFullString()))
			}
		}
	}

	return strings.Join(valuesForObj, "-")
}

//------------------------------------------------------------------------------
// Comparison options
//------------------------------------------------------------------------------

type ComparisonOptions struct {
	fileType  FileType             // the type of the files we're comparing
	idProps   map[PropPath]*IDProp // the properties (values of the map) serving as unique IDs for given paths (keys of the map)
	autoIndex bool                 // if true, then, in an array, an object's index is used as its IDProp, if none is specified for its path in the data tree; i.e. the IDProp `#index` is used, instead of nothing
	fast      bool                 // if true, then, in an array, an object's index is used as its IDProp, if none is specified for its path in the data tree; i.e. the IDProp `#index` is used, instead of nothing
}

func (thisComp *ComparisonOptions) GetFileType() FileType {
	return thisComp.fileType
}

func (thisComp *ComparisonOptions) GetIDProp(atPropPath PropPath) *IDProp {
	configuredProp := thisComp.idProps[atPropPath]

	// applying the autoIndex if required and needed
	if configuredProp == nil && thisComp.autoIndex {
		println(fmt.Sprintf("WARNING: using the array index at path '%s'", atPropPath))

		configuredProp = indexAsID(atPropPath)

		thisComp.idProps[atPropPath] = configuredProp
	}

	return configuredProp
}

// builds a new ComparisonOptions object
func NewOptions(isXml bool, idPropsString string, autoIndex bool, fast bool) *ComparisonOptions {
	fileType := FileTypeJSON
	if isXml {
		fileType = FileTypeXML
	}

	// parsing the "idprops" string
	idProps := map[PropPath]*IDProp{}

	if idPropsString != "" {
		for _, idPropString := range strings.Split(idPropsString, ",") {
			idPropsElems := strings.Split(idPropString, ":::")
			//nolint:gomnd
			if len(idPropsElems) != 2 {
				panic(fmt.Errorf("Error in the 'idprops' flag: '%s' does not respect the \">prop1>prop2>...>propN:idProp\n pattern, "+
					"to configure which object field should be used at a given path to uniquely identify the objects", idPropString))
			}

			// we're building a new ID property
			idProp := &IDProp{from: PropPath(idPropsElems[0])}

			// we're handling the potential combination of several paths used as IDs - like "contract>general>uid+contract>creationDate"
			for _, idPropPath := range strings.Split(idPropsElems[1], "+") {
				successivePaths := []PropPath{}
				for _, successivePath := range strings.Split(idPropPath, ">") {
					successivePaths = append(successivePaths, PropPath(successivePath))
				}

				idProp.props = append(idProp.props, successivePaths)
			}

			// mapping the ID prop to the path where it applies
			idProps[idProp.from] = idProp
		}
	}

	return &ComparisonOptions{
		fileType:  fileType,
		idProps:   idProps,
		autoIndex: autoIndex,
		fast:      fast,
	}
}
