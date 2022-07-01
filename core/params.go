package core

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

//------------------------------------------------------------------------------
// Identifying paths in a data tree
//------------------------------------------------------------------------------

// IdentificationParameter allows to recursively describe how to identity the entities within arrays in a data tree
type IdentificationParameter struct {
	At   string                              `json:"at,omitempty"`
	Use  []string                            `json:"_use,omitempty"`
	When []*ConditionalIDParameter           `json:"when,omitempty"`
	Look []*IdentificationParameter          `json:"look,omitempty"`
	For  map[string]*IdentificationParameter `json:"_for,omitempty"`
	Name string                              `json:"name,omitempty"`

	// technical properties
	parent      *IdentificationParameter
	conditional bool
	fullPath    string
}

// ConditionalIDParameter is an IdentificationParameter that applies only if a given prop has the designated value
type ConditionalIDParameter struct {
	Prop string `json:"prop,omitempty"`
	Is   string `json:"is,omitempty"`
	IdentificationParameter
}

const (
	idParamINDEX = "#index"
)

func (thisParam *IdentificationParameter) isIndex() bool {
	return len(thisParam.Use) == 1 && thisParam.Use[0] == idParamINDEX
}

var _ fmt.Stringer = (*IdentificationParameter)(nil)

// buildFullPath builds this ID param's full path
func (thisParam *IdentificationParameter) buildFullPath() string {
	if thisParam.parent == nil {
		return thisParam.At
	}

	return thisParam.parent.buildFullPath() + ">" + thisParam.At
}

// String returns this ID param's full path, building it once
func (thisParam *IdentificationParameter) String() string {
	if thisParam.fullPath == "" {
		thisParam.fullPath = thisParam.buildFullPath()
	}

	return thisParam.fullPath
}

// isValid checks that this ID parameter does point to identification properties
func (thisParam *IdentificationParameter) checkValidity() error {
	// if len(thisParam.For) == 0 && len(thisParam.Use) == 0 && len(thisParam.Look) == 0 && len(thisParam.When) == 0 {
	// 	return fmt.Errorf("ID param '%s' does not specify which properties to '_use' to build an ID, nor which inner objects to 'look' into, "+
	// 		"nor does it serve as a path '_for' entities deeper in the data tree, nor 'when' to apply!", thisParam)
	// }

	return nil
}

// Resolve makes sure any identification parameter can be properly located within the root identification parameter;
// we take the opportunity here for checking this object's validity
func (thisParam *IdentificationParameter) Resolve() error {
	return thisParam.doResolve(false)
}

func (thisParam *IdentificationParameter) doResolve(conditional bool) error {
	thisParam.conditional = conditional

	for path, subParam := range thisParam.For {
		subParam.parent = thisParam
		if subParam.At == "" {
			subParam.At = path
		}

		if err := subParam.doResolve(conditional); err != nil {
			return err
		}
	}

	for _, condition := range thisParam.When {
		condition.parent = thisParam
		if condition.At == "" {
			condition.At = thisParam.At
		}

		if err := condition.doResolve(true); err != nil {
			return err
		}
	}

	for _, looked := range thisParam.Look {
		looked.parent = thisParam

		if err := looked.doResolve(conditional); err != nil {
			return err
		}
	}

	return thisParam.checkValidity()
}

// isVerifiedBy returns true if the given object verifies this condition
func (thisCondition *ConditionalIDParameter) isVerifiedBy(obj map[string]interface{}) bool {
	if obj == nil {
		return false
	}

	return fmt.Sprintf("%v", obj[thisCondition.Prop]) == thisCondition.Is
}

//------------------------------------------------------------------------------
// Using an idenfication parameter to build a unique ID key
//------------------------------------------------------------------------------

const (
	sepPLUS     = "+"
	sepPIPE     = "|"
	currentPATH = "."
	removeNOPE  = "nope"
)

//buildUniqueKey tries to build a unique key for the given object, according to what's configured on the given ID param
func (thisParam *IdentificationParameter) BuildUniqueKey(obj map[string]interface{}) (result string) {
	return thisParam.doBuildUniqueKey(obj, 0)
}

//nolint:gocognit,gocyclo,cyclop
func (thisParam *IdentificationParameter) doBuildUniqueKey(obj map[string]interface{}, depth int) (result string) {
	// handling the particular cases specificied in the "when"
	if len(thisParam.When) > 0 {
		for _, condition := range thisParam.When {
			if condition.isVerifiedBy(obj) {
				result = concatSeparatedString(condition.Name, sepPLUS, condition.doBuildUniqueKey(obj, depth+1))

				goto End
			}
		}
	}

	// using the "use" if there's one
	if len(thisParam.Use) > 0 {
		for _, prop := range thisParam.Use {
			switch value, ok := obj[prop]; value.(type) {
			case float64:
				//nolint:errcheck
				floatValue := value.(float64)
				if floatValue == float64(int(floatValue)) {
					result = concatSeparatedString(result, sepPLUS, strconv.Itoa(int(floatValue)))
				} else {
					//nolint:revive, gomnd
					result = concatSeparatedString(result, sepPLUS, strconv.FormatFloat(floatValue, 'f', 6, 64))
				}

			case string:
				result = concatSeparatedString(result, sepPLUS, value.(string))

			case bool:
				if value.(bool) {
					result = concatSeparatedString(result, sepPLUS, "true")
				} else {
					result = concatSeparatedString(result, sepPLUS, "false")
				}

			default:
				// if we have a nil value at the intended path, we still use it
				if value == nil {
					if ok { // the value was present
						result = concatSeparatedString(result, sepPLUS, thisParam.At)
					} else { // the value was missing
						result = concatSeparatedString(result, sepPLUS, "("+thisParam.At+")")
					}
				} else {
					panic(fmt.Errorf("Cannot handle the value (of type: %T) at path '%s' (which is part of this id property: %s:::%s - or: %v). Value = %v",
						value, thisParam.At, removeNOPE, removeNOPE, removeNOPE, value))
				}
			}
		}

		if !thisParam.conditional && result == "" {
			panic(fmt.Sprintf("This '_use' configuration: [%s] (at path: %s), did not allow us to build a non-empty ID key",
				strings.Join(thisParam.Use, ", "), thisParam.String()))
		}

		goto End
	}

	// else, "look"-ing for the complex case
	for _, idParam := range thisParam.Look {
		// we're looking at our current object itself
		if idParam.At == currentPATH {
			//
			result = concatSeparatedString(result, sepPLUS, idParam.doBuildUniqueKey(obj, depth+1))
			//
		} else {
			// if we're not using the current object at path ".", then let's go deeper
			switch target, ok := obj[idParam.At]; target.(type) {

			case map[string]interface{}:
				// we're "descending" into an object here
				result = concatSeparatedString(result, sepPLUS, idParam.doBuildUniqueKey(target.(map[string]interface{}), depth+1))

			case []map[string]interface{}:
				// now, we're building a key from an array of objects, hurraaay
				values := []string{}
				for _, targetItem := range target.([]map[string]interface{}) {
					key := idParam.doBuildUniqueKey(targetItem, depth+1)
					if key != "" || !idParam.conditional {
						values = append(values, key)
					}
				}

				// making sure we'll build consistent keys
				sort.Strings(values)

				// let's not forget we might be looking at several objects here
				result = concatSeparatedString(result, sepPLUS, strings.Join(values, sepPIPE))

			default:
				// if we have a nil value at the intended path, we still use it
				if target == nil {
					if ok { // the value was present
						result = concatSeparatedString(result, sepPLUS, idParam.At)
					} else { // the value was missing
						result = concatSeparatedString(result, sepPLUS, "("+idParam.At+")")
					}
				} else {
					panic(fmt.Errorf("Cannot handle the object (of type: %T) at path '%s' (which is part of this id property: %s:::%s - or: %v). Value = %v",
						target, thisParam.At, removeNOPE, removeNOPE, removeNOPE, target))
				}
			}
		}
	}

	if !thisParam.conditional && result == "" {
		panic(fmt.Sprintf("The 'look' configuration at path: '%s' did not allow us to build a non-empty ID key", thisParam))
	}

End:

	return
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
