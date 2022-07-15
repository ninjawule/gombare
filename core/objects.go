package core

import (
	"fmt"
	"reflect"
	"runtime/debug"
)

//------------------------------------------------------------------------------
// Here we compare objects in general
//------------------------------------------------------------------------------

// compareObjects : comparing 2 objects in general - they can be maps, slices, or simple types
//nolint:cyclop,gocyclo
// func compareObjects(currentPath PropPath, obj1, obj2 interface{}, options *ComparisonOptions, currentPathValue PropPath) (Comparison, error) {
func compareObjects(orig1, orig2 map[string]interface{}, idParam *IdentificationParameter, obj1, obj2 interface{}, options *ComparisonOptions, currentPathValue string) (Comparison, error) {
	// considering the kind for the two objects to compare
	obj1Kind := reflect.ValueOf(obj1).Kind()
	obj2Kind := reflect.ValueOf(obj2).Kind()

	// handling nils
	obj1Nil := obj1Kind == reflect.Invalid
	obj2Nil := obj2Kind == reflect.Invalid

	if obj1Nil != obj2Nil {
		if obj1Nil {
			return two(obj2), nil
		}

		return one(obj1), nil
	}

	if obj1Nil {
		// then both objects are nil - we still return a void comparison to avoid nil exceptions
		return nodif(), nil
	}

	// if the kinds are not equal, then we signal an error
	if obj1Kind != obj2Kind {
		// Go's unmarshalling process can lead to having different kinds here, when we juste have kind1 = sliceOf(kind2) or kind2 = sliceOf(kind1);
		// it just could not see that both objects are to be considered as slices, not just one of them
		if obj1Kind == reflect.Slice { // here, we assume that obj1 is a slice of objects of the same kind as the single object obj2; but this could fail!
			switch obj2Kind {
			case reflect.String:
				return compareSlicesOfStrings(idParam, obj1.([]string), []string{obj2.(string)}, options, currentPathValue)
			case reflect.Map:
				return compareSlicesOfMaps(orig1, orig2, idParam, obj1.([]map[string]interface{}), []map[string]interface{}{obj2.(map[string]interface{})}, options, currentPathValue)
			default:
				return compareSlicesOfObjects(orig1, orig2, idParam, obj1.([]interface{}), []interface{}{obj2}, options, currentPathValue)
			}
		}

		if obj2Kind == reflect.Slice { // here, we assume that obj2 is a slice of objects of the same kind as the single object obj1; but this could fail!
			switch obj1Kind {
			case reflect.String:
				return compareSlicesOfStrings(idParam, []string{obj1.(string)}, obj2.([]string), options, currentPathValue)
			case reflect.Map:
				return compareSlicesOfMaps(orig1, orig2, idParam, []map[string]interface{}{obj1.(map[string]interface{})}, obj2.([]map[string]interface{}), options, currentPathValue)
			default:
				return compareSlicesOfObjects(orig1, orig2, idParam, []interface{}{obj1}, obj2.([]interface{}), options, currentPathValue)
			}
		}

		// in any other case, we cannot go any further in the comparison (for now, maybe we'll evolve that later)
		return nil, fmt.Errorf("Issue at path '%s' (%s): type of object '%s' in the first file VS type of object '%s' in the second file\n%v\n\nVS\n\n%s\n\n%s",
			idParam, currentPathValue, obj1Kind, obj2Kind, obj1, obj2, debug.Stack())
	}

	// now, we can deal with our objects, depending on their type
	switch obj1Kind {
	case reflect.Bool:
		if obj1.(bool) != obj2.(bool) {
			return one_two(obj1, obj2), nil
		}

	case reflect.Float64:
		if obj1.(float64) != obj2.(float64) {
			return one_two(obj1, obj2), nil
		}

	case reflect.String:
		if obj1.(string) != obj2.(string) {
			return one_two(obj1, obj2), nil
		}

	case reflect.Slice:
		switch obj1.(type) {
		case []interface{}:
			return compareSlicesOfObjects(orig1, orig2, idParam, obj1.([]interface{}), obj2.([]interface{}), options, currentPathValue)
		case []map[string]interface{}:
			if idParam == nil {
				panic(fmt.Errorf("No id param at path '%s'. Currently compared slices of maps: \n\nslice 1:%v\n\nslice 2:%v", currentPathValue, obj1, obj2))
			}

			return compareSlicesOfMaps(orig1, orig2, idParam, obj1.([]map[string]interface{}), obj2.([]map[string]interface{}), options, currentPathValue)
		}

	case reflect.Map:
		return compareMaps(idParam, obj1.(map[string]interface{}), obj2.(map[string]interface{}), options, currentPathValue, false)

	default:
		// this should never happen
		return nil, fmt.Errorf("Issue at path '%s' : type '%s' is not handled", idParam, obj1Kind)
	}

	// we still return a void comparison to avoid nil exceptions
	return nodif(), nil
}
