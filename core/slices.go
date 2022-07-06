package core

import (
	"fmt"
	"reflect"
	"strconv"
)

//------------------------------------------------------------------------------
// Here we compare slices, which is the more problematic case,
// since we need ordering - general slice case here
//------------------------------------------------------------------------------

func compareSlicesOfObjects(idParam *IdentificationParameter, slice1, slice2 []interface{}, options *ComparisonOptions, currentPathValue string) (Comparison, error) {
	// handling empty
	slice1Empty := len(slice1) == 0
	slice2Empty := len(slice2) == 0

	// one is empty and not this other ?
	if slice1Empty != slice2Empty {
		if slice1Empty {
			return two(slice2), nil
		}

		return one(slice1), nil
	}

	// both are empty ?
	if slice1Empty {
		return nodif(), nil
	}

	// both are non empty, we can consider their 1st element
	slice1Kind := reflect.ValueOf(slice1[0]).Kind()
	slice2Kind := reflect.ValueOf(slice2[0]).Kind()

	// for now, we reject slices with heterogenous kinds
	if slice1Kind != slice2Kind {
		return nil, fmt.Errorf("Issue at path '%s' : type '[]%s' in the first file VS type '[]%s' in the second file.\n\n%v\n\nVS\n\n%v",
			idParam, slice1Kind, slice2Kind, slice1, slice2)
	}

	// transforming the slices to maps, to allow for map comparison
	var map1, map2 map[string]interface{}

	var errTransfo error

	if map1, errTransfo = sliceToMapOfObjects(idParam, slice1Kind, slice1, options, currentPathValue); errTransfo != nil {
		return nil, errTransfo
	}

	if map2, errTransfo = sliceToMapOfObjects(idParam, slice2Kind, slice2, options, currentPathValue); errTransfo != nil {
		return nil, errTransfo
	}

	// we know how to deal with maps
	return compareMaps(idParam, map1, map2, options, currentPathValue, true)
}

//nolint:cyclop,gocyclo,gocognit
func sliceToMapOfObjects(idParam *IdentificationParameter, sliceKind reflect.Kind, slice []interface{}, options *ComparisonOptions, currentPathValue string) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	switch sliceKind {
	case reflect.Bool:
		return nil, fmt.Errorf("Slices of %ss (like at path '%s') are not handled yet! Who use them anyway ??", sliceKind, idParam)

	case reflect.Float64: // building a map of floats (or integers), using their values as keys
		for _, number := range slice {
			//nolint:errcheck
			floatID := number.(float64)
			if floatID == float64(int(floatID)) {
				result[strconv.Itoa(int(floatID))] = number
			} else {
				//nolint:revive, gomnd
				result[strconv.FormatFloat(floatID, 'f', 6, 64)] = number
			}
		}

	case reflect.String: // building a map of strings, using their values as keys
		for _, word := range slice {
			result[word.(string)] = word
		}

	case reflect.Map: // building a map of objects, using their id prop as keys
		// // do we need to sort here ?
		// if sortProp := options.orderBy[currentPath]; sortProp != nil {
		// 	sort.Slice(slice, func(i, j int) bool {
		// 		return sortProp.getValueForObj(slice[i].(map[string]interface{})) < sortProp.getValueForObj(slice[j].(map[string]interface{}))
		// 	})
		// }
		//
		// special case where we use the slice elements' indexes as keys for the map we're building
		if idParam.isIndex() {
			for i, objInSlice := range slice {
				result[fmt.Sprintf("#%d", i+1)] = objInSlice
			}

			return result, nil
		}

		// or... using the value targeted by the ID property for each object as its ID
		for index, object := range slice {
			key := idParam.BuildUniqueKey(object.(map[string]interface{}), index)
			if key == "" {
				return nil, fmt.Errorf("Comparison cannot be done: there is 1 object with an empty key at path '%s' (%s)", idParam, currentPathValue)
			}

			if options.fast {
				result[key] = object
			} else {
				if result[key] != nil && !options.ignoredDups[key] {
					return nil, fmt.Errorf("Comparison has failed: there is more than 1 object with key '%s' at path '%s' (%s)", key, idParam, currentPathValue)
				}
				result[key] = object
			}
		}

	case reflect.Slice: // we have a freaking MATRIX here !
		// we ASSUME that the elements inside each "cell" of this multi-dimensional array are of the same nature; (if not, for now, we're screwed)
		// that being said, we don't see any interest, for the purpose of comparing stuff, of maintaining such a complex structure;
		// so, we'll put every single element that is not a slice, into a single array, before treating it like a normal slice
		matrixAsSlice := matrixToSlice(slice)
		matrixCellKind := reflect.ValueOf(matrixAsSlice[0]).Kind()

		return sliceToMapOfObjects(idParam, matrixCellKind, matrixAsSlice, options, currentPathValue)

	default:
		// this should never happen
		return nil, fmt.Errorf("Cannot sort a slice of %ss yet!", sliceKind)
	}

	return result, nil
}

//------------------------------------------------------------------------------
// Here we specifically compare slices of maps
//------------------------------------------------------------------------------

func compareSlicesOfMaps(idParam *IdentificationParameter, slice1, slice2 []map[string]interface{}, options *ComparisonOptions, currentPathValue string) (Comparison, error) {
	// handling empty
	slice1Empty := len(slice1) == 0
	slice2Empty := len(slice2) == 0

	// one is empty and not this other ?
	if slice1Empty != slice2Empty {
		if slice1Empty {
			return two(slice2), nil
		}

		return one(slice1), nil
	}

	// both are empty ?
	if slice1Empty {
		return nodif(), nil
	}

	// mapping all the maps
	var map1, map2 map[string]interface{}

	var errTransfo error

	if map1, errTransfo = sliceToMapOfMaps(idParam, slice1, options, currentPathValue); errTransfo != nil {
		return nil, errTransfo
	}

	if map2, errTransfo = sliceToMapOfMaps(idParam, slice2, options, currentPathValue); errTransfo != nil {
		return nil, errTransfo
	}

	// we know how to deal with maps
	return compareMaps(idParam, map1, map2, options, currentPathValue, true)
}

func sliceToMapOfMaps(idParam *IdentificationParameter, slice []map[string]interface{}, options *ComparisonOptions, currentPathValue string) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	// // do we need to sort here ?
	// if sortProp := options.orderBy[currentPath]; sortProp != nil {
	// 	sort.Slice(slice, func(i, j int) bool {
	// 		return sortProp.getValueForObj(slice[i]) < sortProp.getValueForObj(slice[j])
	// 	})
	// }

	// using the slice elements' indexes as keys ?
	if idParam.isIndex() {
		for i, mapInSlice := range slice {
			result[fmt.Sprintf("#%d", i+1)] = mapInSlice
		}

		return result, nil
	}

	// or... using the value targeted by the ID property for each object as its ID
	for index, mapInSlice := range slice {
		key := idParam.BuildUniqueKey(mapInSlice, index)
		if key == "" {
			return nil, fmt.Errorf("Comparison cannot be done: there is 1 object with an empty key at path '%s' (%s)", idParam, currentPathValue)
		}

		if options.fast {
			result[key] = mapInSlice
		} else {
			if result[key] != nil && !options.ignoredDups[key] {
				return nil, fmt.Errorf("Comparison has failed: there is more than 1 object with key '%s' at path '%s' (%s)", key, idParam, currentPathValue)
			}
			result[key] = mapInSlice
		}
	}

	return result, nil
}

//------------------------------------------------------------------------------
// Here we specifically compare slices of strings
//------------------------------------------------------------------------------

func compareSlicesOfStrings(idParam *IdentificationParameter, slice1, slice2 []string, options *ComparisonOptions, currentPathValue string) (Comparison, error) {
	return compareMaps(idParam, sliceOfStringsToMap(slice1), sliceOfStringsToMap(slice2), options, currentPathValue, false)
}

func sliceOfStringsToMap(slice []string) map[string]interface{} {
	result := map[string]interface{}{}

	for _, value := range slice {
		result[value] = value
	}

	return result
}
