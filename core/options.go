package utils

import (
	"fmt"
	"log"
	"strings"
)

//------------------------------------------------------------------------------
// Comparison options
//------------------------------------------------------------------------------

type ComparisonOptions struct {
	fileType    FileType             // the type of the files we're comparing
	idProps     map[PropPath]*IDProp // the properties (values of the map) serving as unique IDs for given paths (keys of the map)
	autoIndex   bool                 // if true, then, in an array, an object's index is used as its IDProp, if none is specified for its path in the data tree; i.e. the IDProp `#index` is used, instead of nothing
	fast        bool                 // if true, then, in an array, an object's index is used as its IDProp, if none is specified for its path in the data tree; i.e. the IDProp `#index` is used, instead of nothing
	silent      bool                 // if true, then no info / warning message is written out
	orderBy     map[PropPath]*IDProp // the properties (values of the map) serving as sorting keys for given paths (keys of the map)
	stopAtFirst bool                 // if true, then, when comparing folders, we stop at the first couple of files that differ
}

func (thisComp *ComparisonOptions) GetFileType() FileType {
	return thisComp.fileType
}

func (thisComp *ComparisonOptions) GetIDProp(atPropPath PropPath) *IDProp {
	configuredProp := thisComp.idProps[atPropPath]

	// applying the autoIndex if required and needed
	if configuredProp == nil && thisComp.autoIndex {
		if !thisComp.silent {
			log.Println(fmt.Sprintf("WARNING: using the array index at path '%s'", atPropPath))
		}

		configuredProp = indexAsID(atPropPath)

		thisComp.idProps[atPropPath] = configuredProp
	}

	return configuredProp
}

const propALIAS_SEP = " as "

// builds a new ComparisonOptions object
func NewOptions(isXml bool, idPropsString string, autoIndex bool, orderByString string, fast bool, silent bool, stopAtFirst bool) *ComparisonOptions {
	fileType := FileTypeJSON
	if isXml {
		fileType = FileTypeXML
	}

	return &ComparisonOptions{
		fileType:    fileType,
		idProps:     parsePathsAndPropsString(idPropsString, "idprops"),
		autoIndex:   autoIndex,
		fast:        fast,
		silent:      silent,
		orderBy:     parsePathsAndPropsString(orderByString, "orderby"),
		stopAtFirst: stopAtFirst,
	}
}

func parsePathsAndPropsString(pathsAndPropsString string, optionString string) map[PropPath]*IDProp {
	// parsing the "idprops" string
	props := map[PropPath]*IDProp{}

	if pathsAndPropsString != "" {
		for _, propString := range strings.Split(pathsAndPropsString, ",") {
			propsElems := strings.Split(propString, ":::")
			//nolint:gomnd
			if len(propsElems) != 2 {
				panic(fmt.Errorf("Error in the '%s' flag: '%s' does not respect the \">prop1>prop2>...>propN:prop\n pattern, "+
					"to configure which object field should be used at a given path to uniquely identify the objects", optionString, propString))
			}

			// we're building a new ID property
			prop := &IDProp{from: PropPath(strings.TrimSpace(propsElems[0]))}

			// we're handling the potential combination of several paths used as IDs - like "contract>general>uid+contract>creationDate"
			for _, propPath := range strings.Split(propsElems[1], "+") {
				successivePaths := []PropPath{}
				alias := ""

				for _, successivePath := range strings.Split(propPath, ">") {
					switch elements := strings.Split(successivePath, propALIAS_SEP); len(elements) {
					case 1:
						successivePaths = append(successivePaths, PropPath(strings.TrimSpace(successivePath)))
					//nolint
					case 2:
						successivePaths = append(successivePaths, PropPath(strings.TrimSpace(elements[0])))
						alias = elements[1]
					default:
						panic(fmt.Errorf("Error while using an alias in this ID prop: %s", successivePath))
					}
				}

				prop.props = append(prop.props, successivePaths)
				prop.alias = alias
			}

			// mapping the ID prop to the path where it applies
			props[prop.from] = prop
		}
	}

	return props
}
