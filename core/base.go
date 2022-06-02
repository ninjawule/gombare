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
	return thisPath + ">" + idProp.getIdPartAsPath()
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
	idStr PropPath     // the "string" version, e.g. "contract>general>uid+contract>creationDate"
	alias string       // an alias, if the ID prop is too long
}

const (
	idPropINDEX PropPath = "#index"
)

func (thisProp *IDProp) getIdPartAsPath() PropPath {
	if thisProp.alias != "" {
		return PropPath(thisProp.alias)
	}

	return thisProp.getFullIdPart()
}

func (thisProp *IDProp) getFullIdPart() PropPath {
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

		thisProp.idStr = PropPath(strings.Join(paths, "+"))
	}

	return thisProp.idStr
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
		idStr: idPropINDEX,
	}
}

func (thisProp *IDProp) getValueForObj(obj map[string]interface{}) string {
	valuesForObj := []string{}

	for _, pathChain := range thisProp.props { // ranging over smth like this: [ [contract, general, uid], [contract, creationDate] ]
		currentObj := obj // starting from the "root" object

		for _, path := range pathChain { // ranging over [contract, general, uid]
			// getting the value at that path from the current object:
			switch value, ok := currentObj[string(path)]; value.(type) {
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
				// if we have a nil value at the intended path, we still use it
				if value == nil {
					if ok {
						valuesForObj = append(valuesForObj, string(path))
					} else {
						valuesForObj = append(valuesForObj, "("+string(path)+")")
					}
				} else {
					panic(fmt.Errorf("Cannot handle the value (of type: %T) at path '%s' (which is part of this id property: %s:::%s). Value = %v",
						value, path, thisProp.from, thisProp.getFullIdPart(), value))
				}
			}
		}
	}

	return strings.Join(valuesForObj, "-")
}
