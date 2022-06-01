package utils

//------------------------------------------------------------------------------
// Here we deal with slices of slices
//------------------------------------------------------------------------------

// matrixToSlice : flattens the given "matrix" (2D or more) into a simple slice
func matrixToSlice(matrix []interface{}) []interface{} {
	return copyMatrixElementsToSlice(matrix, []interface{}{})
}

// recursive function that helps flattening a matrix into a slice
func copyMatrixElementsToSlice(matrix []interface{}, resultSlice []interface{}) []interface{} {
	// no elements ? get out!
	if len(matrix) == 0 {
		return resultSlice
	}

	// now, let's handle the matrix elements, depending on their type
	switch matrix[0].(type) {

	// here we need recursion to handle 3D and more
	case []interface{}:
		for _, elem := range matrix {
			resultSlice = copyMatrixElementsToSlice(elem.([]interface{}), resultSlice)
		}

		// we have a 2D matrix here of maps (i.e. of JSON objects)
	case []map[string]interface{}:
		// ranging over 1D
		for _, elem := range matrix {
			// ranging over the other 1D
			for _, innerMap := range elem.([]map[string]interface{}) {
				resultSlice = append(resultSlice, innerMap)
			}
		}

		// for any other case, for now, we consider that it's flat enough :)
	default:
		resultSlice = append(resultSlice, matrix...)
	}

	return resultSlice
}
