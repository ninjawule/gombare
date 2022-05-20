package utils

//------------------------------------------------------------------------------
// Here we compare slices, which is the more problematic case,
// since we need ordering
//------------------------------------------------------------------------------

func compareSlices(currentPath string, obj1, obj2 []interface{}) (Comparison, error) {
	// return nil, errors.New("slices are not handled yet")
	return Comparison{}, nil
}
