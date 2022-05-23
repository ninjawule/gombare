package utils

import (
	"fmt"
	"reflect"
)

//------------------------------------------------------------------------------
// Here we compare objects in general
//------------------------------------------------------------------------------

// compareObjects : comparing 2 objects in general - they can be maps, slices, or simple types
func compareObjects(currentPath string, obj1, obj2 interface{}, idProps map[string]string) (Comparison, error) {
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
		return nil, fmt.Errorf("Issue at path '%s' : type '%s' in the first file VS type '%s' in the second file", currentPath, obj1Kind, obj2Kind)
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
		return compareSlices(currentPath, obj1.([]interface{}), obj2.([]interface{}), idProps)

	case reflect.Map:
		return compareMaps(currentPath, obj1.(map[string]interface{}), obj2.(map[string]interface{}), idProps)
	default:
		// this should never happen
		return nil, fmt.Errorf("Issue at path '%s' : type '%s' is not handled", currentPath, obj1Kind)
	}

	// we still return a void comparison to avoid nil exceptions
	return nodif(), nil
}
