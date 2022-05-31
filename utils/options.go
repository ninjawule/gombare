package utils

import (
	"fmt"
	"strings"
)

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

const idPropAS = " as "

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
				alias := ""

				for _, successivePath := range strings.Split(idPropPath, ">") {
					switch elements := strings.Split(successivePath, idPropAS); len(elements) {
					case 1:
						successivePaths = append(successivePaths, PropPath(successivePath))
					case 2:
						successivePaths = append(successivePaths, PropPath(elements[0]))
						alias = elements[1]
					default:
						panic(fmt.Errorf("Error while using an alias in this ID prop: %s", successivePath))
					}
				}

				idProp.props = append(idProp.props, successivePaths)
				idProp.alias = alias
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
