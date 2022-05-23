package utils

//------------------------------------------------------------------------------
// Here is the base for comparing stuff
//------------------------------------------------------------------------------

// yeah, let's make a recursive type... Why not ?
type Comparison map[string]interface{}

func (comp Comparison) hasDiffs() bool {
	return len(comp) > 0
}

func nodif() Comparison {
	return Comparison{}
}

func one(obj interface{}) Comparison {
	return map[string]interface{}{"one": obj}
}

func two(obj interface{}) Comparison {
	return map[string]interface{}{"two": obj}
}

func one_two(obj1, obj2 interface{}) Comparison {
	return map[string]interface{}{"one": obj1, "two": obj2}
}
