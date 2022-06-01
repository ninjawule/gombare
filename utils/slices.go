package utils

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
)

//------------------------------------------------------------------------------
// Here we compare slices, which is the more problematic case,
// since we need ordering - general slice case here
//------------------------------------------------------------------------------

func compareSlicesOfObjects(currentPath PropPath, slice1, slice2 []interface{}, options *ComparisonOptions) (Comparison, error) {
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
		return nil, fmt.Errorf("Issue at path '%s' : type '%s' in the first file VS type '%s' in the second file", currentPath, slice1Kind, slice2Kind)
	}

	// transforming the slices to maps, to allow for map comparison
	map1 := sliceToMapOfObjects(currentPath, slice1Kind, slice1, options)
	map2 := sliceToMapOfObjects(currentPath, slice2Kind, slice2, options)

	// we know how to deal with maps
	return compareMaps(currentPath, map1, map2, options, true)
}

//nolint:cyclop,gocyclo,gocognit
func sliceToMapOfObjects(currentPath PropPath, sliceKind reflect.Kind, slice []interface{}, options *ComparisonOptions) map[string]interface{} {
	result := map[string]interface{}{}

	switch sliceKind {
	case reflect.Bool:
		panic(fmt.Errorf("Slices of %ss (like at path '%s') are not handled yet! Who use them anyway ??", sliceKind, currentPath))

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
		// controlling that we have an ID to identify the objects in the map
		idProp := options.GetIDProp(currentPath)
		if idProp == nil {
			panic(fmt.Sprintf("Cannot compare the arrays at path '%s' since no ID property has been provided to uniquely identify the objects within (cf. -idprops option)", currentPath))
		}

		// do we need to sort here ?
		if sortProp := options.orderBy[currentPath]; sortProp != nil {
			sort.Slice(slice, func(i, j int) bool {
				return sortProp.getValueForObj(slice[i].(map[string]interface{})) < sortProp.getValueForObj(slice[j].(map[string]interface{}))
			})
		}

		// special case where we use the slice elements' indexes as keys for the map we're building
		if idProp.isIndex() {
			for i, objInSlice := range slice {
				result[fmt.Sprintf("#%d", i+1)] = objInSlice
			}

			return result
		}

		// or... using the value targeted by the ID property for each object as its ID
		for _, object := range slice {
			if options.fast {
				result[idProp.getValueForObj(object.(map[string]interface{}))] = object
			} else {
				key := idProp.getValueForObj(object.(map[string]interface{}))
				if result[key] != nil {
					panic(fmt.Errorf("Comparison has failed: there is more than 1 object with key '%s' at path '%s'", key, currentPath))
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

		return sliceToMapOfObjects(currentPath, matrixCellKind, matrixAsSlice, options)

	default:
		// this should never happen
		panic(fmt.Errorf("Cannot sort a slice of %ss yet!", sliceKind))
	}

	return result
}

//------------------------------------------------------------------------------
// Here we specifically compare slices of maps
//------------------------------------------------------------------------------

func compareSlicesOfMaps(currentPath PropPath, slice1, slice2 []map[string]interface{}, options *ComparisonOptions) (Comparison, error) {
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
	map1 := sliceToMapOfMaps(currentPath, slice1, options)
	map2 := sliceToMapOfMaps(currentPath, slice2, options)

	// we know how to deal with maps
	return compareMaps(currentPath, map1, map2, options, true)
}

func sliceToMapOfMaps(currentPath PropPath, slice []map[string]interface{}, options *ComparisonOptions) map[string]interface{} {
	result := map[string]interface{}{}

	// controlling that we have an ID to identify the objects in the map
	idProp := options.GetIDProp(currentPath)
	if idProp == nil {
		panic(fmt.Sprintf("Cannot compare the arrays at path '%s' since no ID property has been provided to uniquely identify the objects within (cf. -idprops option)", currentPath))
	}

	// using the slice elements' indexes as keys ?
	if idProp.isIndex() {
		for i, mapInSlice := range slice {
			result[fmt.Sprintf("#%d", i+1)] = mapInSlice
		}

		return result
	}

	// or... using the value targeted by the ID property for each object as its ID
	for _, mapInSlice := range slice {
		result[idProp.getValueForObj(mapInSlice)] = mapInSlice
	}

	return result
}
