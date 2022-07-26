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
	parentPATH  = ".."
)

//buildUniqueKey tries to build a unique key for the given object, according to what's configured on the given ID param
func (thisParam *IdentificationParameter) BuildUniqueKey(ent *JsonEntity, currentPathValue string) (result string) {
	return thisParam.doBuildUniqueKey(ent, currentPathValue)
}

//nolint:gocognit,gocyclo,cyclop
func (thisParam *IdentificationParameter) doBuildUniqueKey(ent *JsonEntity, currentPathValue string) (result string) {
	// handling the particular cases specificied in the "when"
	if len(thisParam.When) > 0 {
		for _, condition := range thisParam.When {
			if condition.isVerifiedBy(ent) {
				result = concatSeparatedString(condition.Name, sepPLUS, condition.doBuildUniqueKey(ent, currentPathValue))

				goto End
			}
		}
	}

	// using the "use" if there's one
	if len(thisParam.Use) > 0 {
		for _, prop := range thisParam.Use {
			result = concatSeparatedString(result, sepPLUS, thisParam.getStringValueFromObj(ent.values, prop))
		}

		if !thisParam.isWithinWhen() && result == "" {
			panic(fmt.Sprintf("This '_use' configuration: [%s] (at path: %s), did not allow us to build a non-empty ID key",
				strings.Join(thisParam.Use, ", "), thisParam.toString()))
		}

		goto End
	}

	// else, "look"-ing for the complex case
	for _, nextIdParam := range thisParam.Look {
		if nextIdParam.At == parentPATH { // we're looking back
			// getting the origin of the current origin - we'll call it the "ancestor"
			if ancestor := ent.parent; ancestor != nil {
				result = concatSeparatedString(result, sepPLUS, nextIdParam.doBuildUniqueKey(ancestor, currentPathValue))
			} else {
				panic(fmt.Sprintf("No parent found with '%s' from '%s' (param = %s). Current obj = %v", parentPATH, currentPathValue, thisParam.toString(), ent))
			}

		} else if nextIdParam.At == currentPATH { // we're looking at our current object itself
			//
			result = concatSeparatedString(result, sepPLUS, nextIdParam.doBuildUniqueKey(ent, currentPathValue))
			//
		} else {
			// if we're not using the current object at path ".", then let's go deeper
			switch target, ok := ent.values[nextIdParam.At]; target.(type) {

			case map[string]interface{}:
				// we're "descending" into an object here
				result = concatSeparatedString(result, sepPLUS, nextIdParam.doBuildUniqueKey(entityFrom(target, ent), currentPathValue))

			case []map[string]interface{}:
				// now, we're building a key from an array of objects, hurraaay
				values := []string{}
				for _, targetItem := range target.([]map[string]interface{}) {
					key := nextIdParam.doBuildUniqueKey(entityFrom(targetItem, ent), currentPathValue)
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
		result = thisParam.incrKey(ent.parent, result)
	}

	// building an alias for this object ?
	if len(thisParam.getTpl1()) > 0 {
		thisParam.addAlias1(ent.values, currentPathValue)
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

func (thisParam *IdentificationParameter) incrKey(objOwner *JsonEntity, currentKey string) string {
	// we use a cache key that may use the param's name
	cacheKey := concatSeparatedString(thisParam.Name, "&", currentKey)

	// init if needed
	if objOwner.counts == nil {
		objOwner.counts = map[string]int{}
	}

	// increment
	objOwner.counts[cacheKey] = objOwner.counts[cacheKey] + 1

	// formating
	return fmt.Sprintf("%s#%d", currentKey, objOwner.counts[cacheKey])
}
