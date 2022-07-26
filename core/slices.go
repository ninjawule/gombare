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

func compareSlicesOfObjects(root1, root2 *JsonEntity, idParam *IdentificationParameter, slice1, slice2 []interface{},
	options *ComparisonOptions, currentPathValue string) (Comparison, error) {
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

	// should we clear the identical elements before performing a comparison on the diverging elements only ?
	if !idParam.Keep {
		slice1, slice2 = clearObjectsSiblings(slice1, slice2, options, currentPathValue)
	}

	// both are non empty, we can consider their 1st element
	slice1Kind := reflect.ValueOf(slice1[0]).Kind()
	slice2Kind := reflect.ValueOf(slice2[0]).Kind()

	// for now, we reject slices with heterogenous kinds
	if slice1Kind != slice2Kind {
		return nil, fmt.Errorf("Issue at path '%s' : type '[]%s' in the first file VS type '[]%s' in the second file.\n\n%v\n\nVS\n\n%v",
			idParam.toString(), slice1Kind, slice2Kind, slice1, slice2)
	}

	// transforming the slices to maps, to allow for map comparison
	var ent1, ent2 *JsonEntity

	var errTransfo error

	if ent1, errTransfo = sliceToMapOfObjects(1, root1, idParam, slice1Kind, slice1, options, currentPathValue); errTransfo != nil {
		return nil, errTransfo
	}

	//nolint:gomnd
	if ent2, errTransfo = sliceToMapOfObjects(2, root2, idParam, slice2Kind, slice2, options, currentPathValue); errTransfo != nil {
		return nil, errTransfo
	}

	// we know how to deal with maps
	return compareJsonEntities(idParam, ent1, ent2, options, currentPathValue, true)
}

//nolint:cyclop,gocyclo,gocognit
func sliceToMapOfObjects(file int, root *JsonEntity, idParam *IdentificationParameter, sliceKind reflect.Kind, slice []interface{}, options *ComparisonOptions, currentPathValue string) (*JsonEntity, error) {
	ent := &JsonEntity{parent: root, values: map[string]interface{}{}}

	switch sliceKind {
	case reflect.Bool:
		return nil, fmt.Errorf("Slices of %ss (like at path '%s') are not handled yet! Who use them anyway ??", sliceKind, idParam.toString())

	case reflect.Float64: // building a map of floats (or integers), using their values as keys
		for _, number := range slice {
			//nolint:errcheck
			floatID := number.(float64)
			if floatID == float64(int(floatID)) {
				ent.values[strconv.Itoa(int(floatID))] = number
			} else {
				//nolint:revive, gomnd
				ent.values[strconv.FormatFloat(floatID, 'f', 6, 64)] = number
			}
		}

	case reflect.String: // building a map of strings, using their values as keys
		for _, word := range slice {
			ent.values[word.(string)] = word
		}

	case reflect.Map: // building a map of objects, using their id prop as keys
		// using the value targeted by the ID property for each object as its ID
		for _, object := range slice {
			key := idParam.BuildUniqueKey(entityFrom(object, root), currentPathValue)

			// we should never up with an empty key
			if key == "" {
				return nil, fmt.Errorf("Comparison of the 2 slices of OBJECTSs cannot be done: there is 1 object with an empty key at path '%s' in file %d (%s)",
					idParam.toString(), file, currentPathValue)
			}

			if options.fast {
				ent.values[key] = object
			} else {
				// if ent[key] != nil && !options.isIgnoredDuplicate(currentPathValue, key) {
				if ent.values[key] != nil {
					// if it's a real duplicate, then we have to warn about it
					if fmt.Sprintf("%v", ent.values[key]) == fmt.Sprintf("%v", object) {
						// if reflect.DeepEqual(ent[key], object) {
						options.logger.Warn("There are 2 identical objects at path '%s' in file %d (with key '%s'): %v\n", currentPathValue, file, key, object)
					} else {
						return nil, fmt.Errorf("Comparison of the 2 slices of OBJECTs has failed: there is more than 1 object with key '%s' at path '%s'"+
							" in file %d (%s)\n\nmap1: %v\n\nmap2: %v", key, idParam.toString(), file, currentPathValue, ent.values[key], object)
					}
				}
				ent.values[key] = object
			}
		}

	case reflect.Slice: // we have a freaking MATRIX here !
		// we ASSUME that the elements inside each "cell" of this multi-dimensional array are of the same nature; (if not, for now, we're screwed)
		// that being said, we don't see any interest, for the purpose of comparing stuff, of maintaining such a complex structure;
		// so, we'll put every single element that is not a slice, into a single array, before treating it like a normal slice
		matrixAsSlice := matrixToSlice(slice)
		matrixCellKind := reflect.ValueOf(matrixAsSlice[0]).Kind()

		return sliceToMapOfObjects(file, root, idParam, matrixCellKind, matrixAsSlice, options, currentPathValue)

	default:
		// this should never happen
		return nil, fmt.Errorf("Cannot compare a slice of %ss yet!", sliceKind)
	}

	return ent, nil
}

type intObj struct {
	val int
}

func toIntObj(val int) *intObj {
	return &intObj{val}
}

func clearObjectsSiblings(slice1In, slice2In []interface{}, options *ComparisonOptions, currentPathValue string) ([]interface{}, []interface{}) {
	// finding the shortest slice, to use it...
	littleSlice := slice1In
	secondSlice := slice2In
	usedSlice := 1

	if len(slice2In) < len(slice1In) {
		littleSlice = slice2In
		secondSlice = slice1In
		usedSlice = 2
	}

	//... to build a map of all the values therein
	littleSliceValues := map[string]*intObj{} // keeping an "image" for each element of the little slice
	littleSliceKeptIndexes := map[int]bool{}  // the indexes of these elements in their original slice

	for index, littleSliceElement := range littleSlice {
		key := fmt.Sprintf("%v", littleSliceElement)
		if littleSliceValues[key] == nil {
			littleSliceValues[key] = toIntObj(index) // the "image" is just a brute printout of the element
			littleSliceKeptIndexes[index] = true     // at first, we suppose we'll keep every element of the little slice
		} else { // we've detected a duplicate!
			options.logger.Warn("There are 2 identical maps at path '%s' in file %d: %s)\n", currentPathValue, usedSlice, key)
			littleSliceKeptIndexes[index] = false // we won't keep it
		}
	}

	// we'll build the 2nd slice along the way
	var secondSliceOut []interface{}

	for _, secondSliceElement := range secondSlice {
		key := fmt.Sprintf("%v", secondSliceElement)
		if index := littleSliceValues[key]; index != nil { // there!! we've found a common value between the 2 slices
			// we won't be keep the element of the little slice at this index
			littleSliceKeptIndexes[index.val] = false
		} else { // the second slice element here does not have a perfect sibling in the little slice here
			// we know we're keeping this element in the second slice
			secondSliceOut = append(secondSliceOut, secondSliceElement)
		}
	}

	// now, let's find out what we can keep in the little slice
	var littleSliceOut []interface{}

	for index, littleSliceElement := range littleSlice {
		if littleSliceKeptIndexes[index] {
			littleSliceOut = append(littleSliceOut, littleSliceElement)
		}
	}

	if usedSlice == 1 {
		return littleSliceOut, secondSliceOut
	}

	return secondSliceOut, littleSliceOut
}

//------------------------------------------------------------------------------
// Here we specifically compare slices of maps
//------------------------------------------------------------------------------

func compareSlicesOfMaps(root1, root2 *JsonEntity, idParam *IdentificationParameter, slice1, slice2 []map[string]interface{},
	options *ComparisonOptions, currentPathValue string) (Comparison, error) {
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

	// should we clear the identical elements before performing a comparison on the diverging elements only ?
	if !idParam.Keep {
		slice1, slice2 = clearMapsSiblings(slice1, slice2, options, currentPathValue)
	}

	// mapping all the maps
	var map1, map2 *JsonEntity

	var errTransfo error

	if map1, errTransfo = sliceToMapOfMaps(1, root1, idParam, slice1, options, currentPathValue); errTransfo != nil {
		return nil, errTransfo
	}

	//nolint:gomnd
	if map2, errTransfo = sliceToMapOfMaps(2, root2, idParam, slice2, options, currentPathValue); errTransfo != nil {
		return nil, errTransfo
	}

	// we know how to deal with maps
	return compareJsonEntities(idParam, map1, map2, options, currentPathValue, true)
}

func sliceToMapOfMaps(file int, root *JsonEntity, idParam *IdentificationParameter, slice []map[string]interface{}, options *ComparisonOptions, currentPathValue string) (*JsonEntity, error) {
	ent := &JsonEntity{parent: root, values: map[string]interface{}{}}

	// using the value targeted by the ID property for each object as its ID
	for _, mapInSlice := range slice {
		key := idParam.BuildUniqueKey(entity(mapInSlice).from(root), currentPathValue)

		// we should never up with an empty key
		if key == "" {
			return nil, fmt.Errorf("Comparison of the 2 slices of MAPs cannot be done: there is 1 object with an empty key at path '%s' in file %d (%s)",
				idParam.toString(), file, currentPathValue)
		}

		if options.fast {
			ent.values[key] = mapInSlice
		} else {
			// if ent[key] != nil && !options.isIgnoredDuplicate(currentPathValue, key) {
			if ent.values[key] != nil {
				// if it's a real duplicate, then we have to warn about it
				//nolint:revive
				if fmt.Sprintf("%v", ent.values[key]) == fmt.Sprintf("%v", mapInSlice) {
					// if reflect.DeepEqual(ent[key], mapInSlice) {
					options.logger.Warn("There are 2 identical maps at path '%s' in file %d (with key '%s'): %v\n", currentPathValue, file, key, mapInSlice)
				} else {
					return nil, fmt.Errorf("Comparison of the 2 slices of MAPs has failed: there is more than 1 MAP with key '%s' at path '%s'"+
						" in file %d (%s)\n\nmap1: %v\n\nmap2: %v", key, idParam.toString(), file, currentPathValue, ent.values[key], mapInSlice)
				}
			}
			ent.values[key] = mapInSlice
		}
	}

	return ent, nil
}

func clearMapsSiblings(slice1In, slice2In []map[string]interface{}, options *ComparisonOptions, currentPathValue string) ([]map[string]interface{}, []map[string]interface{}) {
	// finding the shortest slice, to use it...
	littleSlice := slice1In
	secondSlice := slice2In
	usedSlice := 1

	if len(slice2In) < len(slice1In) {
		littleSlice = slice2In
		secondSlice = slice1In
		usedSlice = 2
	}

	//... to build a map of all the values therein
	littleSliceValues := map[string]*intObj{} // keeping an "image" for each element of the little slice
	littleSliceKeptIndexes := map[int]bool{}  // the indexes of these elements in their original slice

	for index, littleSliceElement := range littleSlice {
		key := fmt.Sprintf("%v", littleSliceElement)
		if littleSliceValues[key] == nil {
			littleSliceValues[key] = toIntObj(index) // the "image" is just a brute printout of the element
			littleSliceKeptIndexes[index] = true     // at first, we suppose we'll keep every element of the little slice
		} else { // we've detected a duplicate!
			options.logger.Warn("There are 2 identical maps at path '%s' in file %d: %s)\n", currentPathValue, usedSlice, key)
			littleSliceKeptIndexes[index] = false // we won't keep it
		}
	}

	// we'll build the 2nd slice along the way
	var secondSliceOut []map[string]interface{}

	for _, secondSliceElement := range secondSlice {
		key := fmt.Sprintf("%v", secondSliceElement)
		if index := littleSliceValues[key]; index != nil { // there!! we've found a common value between the 2 slices
			// we won't be keep the element of the little slice at this index
			littleSliceKeptIndexes[index.val] = false
		} else { // the second slice element here does not have a perfect sibling in the little slice here
			// we know we're keeping this element in the second slice
			secondSliceOut = append(secondSliceOut, secondSliceElement)
		}
	}

	// now, let's find out what we can keep in the little slice
	var littleSliceOut []map[string]interface{}

	for index, littleSliceElement := range littleSlice {
		if littleSliceKeptIndexes[index] {
			littleSliceOut = append(littleSliceOut, littleSliceElement)
		}
	}

	if usedSlice == 1 {
		return littleSliceOut, secondSliceOut
	}

	return secondSliceOut, littleSliceOut
}

//------------------------------------------------------------------------------
// Here we specifically compare slices of strings
//------------------------------------------------------------------------------

func compareSlicesOfStrings(idParam *IdentificationParameter, slice1, slice2 []string, options *ComparisonOptions, currentPathValue string) (Comparison, error) {
	return compareJsonEntities(idParam, sliceOfStringsToEntity(slice1), sliceOfStringsToEntity(slice2), options, currentPathValue, false)
}

func sliceOfStringsToEntity(slice []string) *JsonEntity {
	ent := &JsonEntity{values: map[string]interface{}{}}

	for _, value := range slice {
		ent.values[value] = value
	}

	return ent
}
