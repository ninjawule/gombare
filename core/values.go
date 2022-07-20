package core

import (
	"fmt"
	"strconv"
	"strings"
)

//------------------------------------------------------------------------------
// Using an idenfication parameter to build a unique ID key
//------------------------------------------------------------------------------

const (
	sepPLUS     = "~"
	sepPIPE     = "|"
	currentPATH = "."
)

//buildUniqueKey tries to build a unique key for the given object, according to what's configured on the given ID param
func (thisParam *IdentificationParameter) BuildUniqueKey(orig, obj map[string]interface{}) (result string) {
	return thisParam.doBuildUniqueKey(orig, obj)
}

//nolint:gocognit,gocyclo,cyclop
func (thisParam *IdentificationParameter) doBuildUniqueKey(orig, obj map[string]interface{}) (result string) {
	// handling the particular cases specificied in the "when"
	if len(thisParam.When) > 0 {
		for _, condition := range thisParam.When {
			if condition.isVerifiedBy(obj) {
				result = concatSeparatedString(condition.Name, sepPLUS, condition.doBuildUniqueKey(orig, obj))

				goto End
			}
		}
	}

	// using the "use" if there's one
	if len(thisParam.Use) > 0 {
		for _, prop := range thisParam.Use {
			result = concatSeparatedString(result, sepPLUS, thisParam.getStringValueFromObj(obj, prop))
		}

		if !thisParam.isWithinWhen() && result == "" {
			panic(fmt.Sprintf("This '_use' configuration: [%s] (at path: %s), did not allow us to build a non-empty ID key",
				strings.Join(thisParam.Use, ", "), thisParam.toString()))
		}

		goto End
	}

	// else, "look"-ing for the complex case
	for _, nextIdParam := range thisParam.Look {
		// we're looking at our current object itself
		if nextIdParam.At == currentPATH {
			//
			result = concatSeparatedString(result, sepPLUS, nextIdParam.doBuildUniqueKey(orig, obj))
			//
		} else {
			// if we're not using the current object at path ".", then let's go deeper
			switch target, ok := obj[nextIdParam.At]; target.(type) {

			case map[string]interface{}:
				// we're "descending" into an object here
				result = concatSeparatedString(result, sepPLUS, nextIdParam.doBuildUniqueKey(obj, target.(map[string]interface{})))

			case []map[string]interface{}:
				// now, we're building a key from an array of objects, hurraaay
				values := []string{}
				for _, targetItem := range target.([]map[string]interface{}) {
					key := nextIdParam.doBuildUniqueKey(obj, targetItem)
					if key != "" || !nextIdParam.isWithinWhen() {
						values = append(values, key)
					}
				}

				// let's not forget we might be looking at several objects here
				result = concatSeparatedString(result, sepPLUS, strings.Join(values, sepPIPE))

			default:
				// if we have a nil value at the intended path, we still use it
				if target == nil {
					if ok { // the value was present
						result = concatSeparatedString(result, sepPLUS, nextIdParam.At+"empty ??")
					} else { // the value was missing
						result = concatSeparatedString(result, sepPLUS, "("+nextIdParam.At+")")
					}
				} else {
					panic(fmt.Errorf("Cannot handle the OBJECT (of type: %T) at path '%s' (which is part of this id param: %v). Value = %v",
						target, thisParam.At, thisParam.toString(), target))
				}
			}
		}
	}

	if !thisParam.isWithinWhen() && result == "" {
		panic(fmt.Sprintf("The 'look' configuration at path: '%s' did not allow us to build a non-empty ID key", thisParam.toString()))
	}

End:

	// handling the increment
	if thisParam.Incr {
		result = thisParam.incrKey(orig, obj, result)
	}

	// building an alias for this object ?
	if len(thisParam.getTpl()) > 0 {
		thisParam.addAlias(obj)
	}

	return
}

// getStringValueFromObj: for a given id param, builds a string value for the given object's property (given by its name)
func (thisParam *IdentificationParameter) getStringValueFromObj(obj map[string]interface{}, prop string) string {
	switch value, ok := obj[prop]; value.(type) {
	case float64:
		//nolint:errcheck
		floatValue := value.(float64)
		if floatValue == float64(int(floatValue)) {
			return strconv.Itoa(int(floatValue))
		}
		//nolint:revive, gomnd
		return strconv.FormatFloat(floatValue, 'f', 6, 64)

	case string:
		return value.(string)

	case bool:
		if value.(bool) {
			return "true"
		}

		return "false"

	case map[string]interface{}:
		// a f*cked up case: we expect to get a tag's value, but if this tag unexpectedly contains attributes,
		// then go creates a map for it, and stores the value with the "#text" key
		return thisParam.getStringValueFromObj(value.(map[string]interface{}), "#text")

	default:
		// if we have a nil value at the intended path, we still use it
		if value == nil {
			if ok { // the value was present
				return prop
			}
			// the value was missing
			return "(" + prop + ")"
		}

		panic(fmt.Errorf("Cannot handle the VALUE (of type: %T) at path '%s', for prop '%s' (which is part of this id param: %s). Value = %v",
			value, thisParam.At, prop, thisParam.toString(), value))
	}
}

//------------------------------------------------------------------------------
// Utils
//------------------------------------------------------------------------------

// utility function to gracefully concatenate 2 strings
func concatSeparatedString(val1, sep, val2 string) string {
	if val1 == "" {
		return val2
	}

	if val2 == "" {
		return val1
	}

	return val1 + sep + val2
}

const objINCREMENTS = "__increments__"

func (thisParam *IdentificationParameter) incrKey(orig, obj map[string]interface{}, currentKey string) string {
	// we use a cache key that may use the param's name
	cacheKey := concatSeparatedString(thisParam.Name, "&", currentKey)

	// init if needed
	if orig[objINCREMENTS] == nil {
		orig[objINCREMENTS] = map[string]int{}
	}

	// increment
	orig[objINCREMENTS].(map[string]int)[cacheKey] = orig[objINCREMENTS].(map[string]int)[cacheKey] + 1

	// formating
	return fmt.Sprintf("%s#%d", currentKey, orig[objINCREMENTS].(map[string]int)[cacheKey])
}
