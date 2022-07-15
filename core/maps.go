package core

import "sort"

//------------------------------------------------------------------------------
// Here we compare maps
//------------------------------------------------------------------------------

// compareMaps : getting a diff between 2 maps
func compareMaps(idParam *IdentificationParameter, map1, map2 map[string]interface{}, options *ComparisonOptions, currentPathValue string, fromSlice bool) (Comparison, error) {
	// the result from comparing the 2 maps
	thisComparison := map[string]interface{}{}

	// first, let's keep track of the keys we encounter
	checked := map[string]bool{}

	// getting all the keys in map 1, and sorting them
	keys1 := []string{}
	for key1 := range map1 {
		keys1 = append(keys1, key1)
	}

	sort.Strings(keys1)

	// let's iterate over the first map, and see what we have in the second
	for _, key1 := range keys1 {
		// we're excluding some technical properties
		if key1 != objINCREMENTS {
			//getting the corresponding object
			obj1 := map1[key1]

			// this key is being checked
			checked[key1] = true

			// this is the full path of the particular object we'll compare to another
			nextPathValue := currentPathValue + ">" + key1

			// what's in the 2nd map ?
			obj2 := map2[key1]

			// what's the next ID parameter associated with the current object ?
			nextIdParam := idParam

			// if the maps here come from 2 arrays, then the key is not a natural map key, but an object ID
			if !fromSlice && idParam != nil {
				nextIdParam = idParam.For[key1]
			}

			// obj1 and obj2 should be compared
			compObj1Obj2, errComp := compareObjects(map1, map2, nextIdParam, obj1, obj2, options, nextPathValue)
			if errComp != nil {
				return nil, errComp
			}

			// adding only if there's at least 1 difference
			if compObj1Obj2.hasDiffs() {
				thisComparison[key1] = compObj1Obj2
			}
		}
	}

	// getting all the keys in map 1, and sorting them
	keys2 := []string{}
	for key2 := range map2 {
		keys2 = append(keys2, key2)
	}

	sort.Strings(keys2)

	// now let's iterate over the second map, because there might be stuff not found in the first map
	for _, key2 := range keys2 {
		// we're considering keys that have not been checked yet - still excluding some technical properties
		if !checked[key2] && key2 != objINCREMENTS {
			// this is the full path of the particular object we'll compare to another
			nextPathValue := currentPathValue + ">" + key2

			// what's the next ID parameter associated with the current object ?
			nextIdParam := idParam

			// if the maps here come from 2 arrays, then the key is not a natural map key, but an object ID
			if !fromSlice && idParam != nil {
				nextIdParam = idParam.For[key2]
			}

			// at this point, obj1 does not exist for this key...
			compObj1Obj2, errComp := compareObjects(map1, map2, nextIdParam, map1[key2], map2[key2], options, nextPathValue)
			if errComp != nil {
				return nil, errComp
			}

			// ... so we're sure to have a difference here
			thisComparison[key2] = compObj1Obj2
		}
	}

	// returning
	return thisComparison, nil
}
