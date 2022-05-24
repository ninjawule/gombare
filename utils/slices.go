package utils

import (
	"fmt"
	"reflect"
	"strconv"
)

//------------------------------------------------------------------------------
// Here we compare slices, which is the more problematic case,
// since we need ordering - general slice case here
//------------------------------------------------------------------------------

func compareSlicesOfObjects(currentPath string, slice1, slice2 []interface{}, idProps map[string]string) (Comparison, error) {
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
	map1 := sliceToMapOfObjects(currentPath, slice1Kind, slice1, idProps)
	map2 := sliceToMapOfObjects(currentPath, slice2Kind, slice2, idProps)

	// we know how to deal with maps
	return compareMaps(currentPath, map1, map2, idProps)
}

func sliceToMapOfObjects(currentPath string, sliceKind reflect.Kind, slice []interface{}, idProps map[string]string) map[string]interface{} {
	result := map[string]interface{}{}

	switch sliceKind {
	case reflect.Bool, reflect.Slice:
		panic(fmt.Errorf("Slices of %ss (like at path '%s') are not handled yet! Who use them anyway ??", sliceKind, currentPath))

	case reflect.Float64:
		for _, number := range slice {
			//nolint:revive, gomnd
			result[strconv.FormatFloat(number.(float64), 'f', 6, 64)] = number
		}

	case reflect.String:
		for _, word := range slice {
			result[word.(string)] = word
		}

	case reflect.Map:
		// controlling that we have an ID to identify the objects in the map
		idProp := idProps[currentPath]
		if idProp == "" {
			panic(fmt.Sprintf("Cannot compare the arrays at path '%s' since no ID property has been provided to uniquely identify the objects within (cf. -idprops option)", currentPath))
		}

		// we build the map depending on the objects' IDs' type
		switch idKind := reflect.ValueOf(slice[0].(map[string]interface{})[idProp]).Kind(); idKind {
		case reflect.Float64:
			for _, object := range slice {
				//nolint:errcheck
				floatID := (object.(map[string]interface{}))[idProp].(float64)
				if floatID == float64(int(floatID)) {
					result[strconv.Itoa(int(floatID))] = object
				} else {
					//nolint:revive, gomnd
					result[strconv.FormatFloat(floatID, 'f', 6, 64)] = object
				}
			}

		case reflect.String:
			for _, object := range slice {
				result[(object.(map[string]interface{}))[idProp].(string)] = object
			}

		default:
			panic(fmt.Sprintf("The property '%s', which is of type '%s', cannot serve as an ID for the objects at path '%s'", idProp, idKind, currentPath))
		}

	default:
		// this should never happen
		panic(fmt.Errorf("Cannot sort a slice of %ss yet!", sliceKind))
	}

	return result
}

//------------------------------------------------------------------------------
// Here we specifically compare slices of maps
//------------------------------------------------------------------------------

func compareSlicesOfMaps(currentPath string, slice1, slice2 []map[string]interface{}, idProps map[string]string) (Comparison, error) {
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
	map1 := sliceToMapOfMaps(currentPath, slice1, idProps)
	map2 := sliceToMapOfMaps(currentPath, slice2, idProps)

	// we know how to deal with maps
	return compareMaps(currentPath, map1, map2, idProps)
}

func sliceToMapOfMaps(currentPath string, slice []map[string]interface{}, idProps map[string]string) map[string]interface{} {
	result := map[string]interface{}{}

	// controlling that we have an ID to identify the objects in the map
	idProp := idProps[currentPath]
	if idProp == "" {
		panic(fmt.Sprintf("Cannot compare the arrays at path '%s' since no ID property has been provided to uniquely identify the objects within (cf. -idprops option)", currentPath))
	}

	// we build the map depending on the objects' IDs' type
	switch idKind := reflect.ValueOf(slice[0][idProp]).Kind(); idKind {
	case reflect.Float64:
		for _, mapInSlice := range slice {
			//nolint:revive, gomnd
			result[strconv.FormatFloat(mapInSlice[idProp].(float64), 'f', 6, 64)] = mapInSlice
		}

	case reflect.String:
		for _, mapInSlice := range slice {
			result[mapInSlice[idProp].(string)] = mapInSlice
		}

	default:
		panic(fmt.Sprintf("The property '%s', which is of type '%s', cannot serve as an ID for the objects at path '%s'", idProp, idKind, currentPath))
	}

	return result
}
