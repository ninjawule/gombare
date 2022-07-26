package core

//------------------------------------------------------------------------------
// Here is the base for comparing stuff
//------------------------------------------------------------------------------

// yeah, comparisons are just big maps, in the end...
type Comparison map[string]interface{}

func (comp Comparison) hasDiffs() bool {
	return len(comp) > 0
}

func nodif() Comparison {
	return Comparison{}
}

func one(obj interface{}) Comparison {
	return Comparison{"_del_": obj}
}

func two(obj interface{}) Comparison {
	return Comparison{"_new_": obj}
}

func one_two(obj1, obj2 interface{}) Comparison {
	return Comparison{"_one_": obj1, "_two_": obj2}
}

//------------------------------------------------------------------------------
// Constants
//------------------------------------------------------------------------------

// the types for the files we're comparing
type FileType string

const (
	FileTypeJSON FileType = "JSON"
	FileTypeXML  FileType = "XML"
)
