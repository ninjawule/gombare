package core

import (
	"fmt"
	"html/template"
)

//------------------------------------------------------------------------------
// Identifying paths in a data tree
//------------------------------------------------------------------------------

// IdentificationParameter allows to recursively describe how to identity the entities within arrays in a data tree
type IdentificationParameter struct {
	At       string                              `json:"at,omitempty"`   // the relative path at which to use this identification parameter
	Use      []string                            `json:"_use,omitempty"` // which simple properties to concatenate to form a key
	Tpl1     []string                            `json:"tpl1,omitempty"` // a formattable string (Go template) to build an alias for an object, instead of outputting it completely in the comparison
	TplN     []string                            `json:"tplN,omitempty"` // a formattable string (Go template) to build an alias for an object, instead of outputting it completely in the comparison
	Incr     bool                                `json:"incr,omitempty"` // if true, then any key built with this ID param is augmented with a counter of its occurrences
	When     []*ConditionalIDParameter           `json:"when,omitempty"` // when to apply this identification parameter, and what to do (_use, look, or when ?)
	Look     []*IdentificationParameter          `json:"look,omitempty"` // which relationships to look into
	For      map[string]*IdentificationParameter `json:"_for,omitempty"` // how to deal with the embedded objects from this place
	Name     string                              `json:"name,omitempty"` // a name for this ID parameter, that may be used as a prefix for the keys built here
	FullPath string                              `json:"path,omitempty"` // the relative path at which to use this identification parameter
	Keep     bool                                `json:"keep,omitempty"` // if true, then, when comparing slice elements with this ID param, we're not clearing the identical elements before comparing the diverging ones

	// technical properties
	parent             *IdentificationParameter
	isWhen             bool
	buildTpl1          *template.Template
	buildTplN          *template.Template
	withinWhen         bool
	withinWhenResolved bool
}

// ConditionalIDParameter is an IdentificationParameter that applies only if a given prop has the designated value
type ConditionalIDParameter struct {
	Prop string `json:"prop,omitempty"`
	Is   string `json:"is,omitempty"`
	IdentificationParameter
}

// buildFullPath builds this ID param's full path
func (thisParam *IdentificationParameter) buildFullPath() string {
	if thisParam.parent == nil {
		return thisParam.At
	}

	if thisParam.isWhen {
		return thisParam.parent.buildFullPath()
	}

	if thisParam.At == "." {
		return thisParam.parent.buildFullPath()
	}

	return thisParam.parent.buildFullPath() + "." + thisParam.At
}

// String returns this ID param's full path, building it once
func (thisParam *IdentificationParameter) toString() string {
	if thisParam.FullPath == "" {
		thisParam.FullPath = thisParam.buildFullPath()
	}

	return thisParam.FullPath
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
func (thisParam *IdentificationParameter) Resolve(full bool) error {
	return thisParam.doResolve(full)
}

func (thisParam *IdentificationParameter) doResolve(full bool) error {
	for path, subParam := range thisParam.For {
		subParam.parent = thisParam
		if subParam.At == "" {
			subParam.At = path
		}

		if full {
			subParam.toString()
			subParam.getTpl1()
			subParam.getTplN()
		}

		if err := subParam.doResolve(full); err != nil {
			return err
		}
	}

	for _, condition := range thisParam.When {
		condition.isWhen = true
		condition.parent = thisParam

		if condition.At == "" {
			condition.At = thisParam.At
		}

		if full {
			condition.toString()
			condition.getTpl1()
			condition.getTplN()
		}

		if err := condition.doResolve(full); err != nil {
			return err
		}
	}

	for _, looked := range thisParam.Look {
		looked.parent = thisParam

		if full {
			looked.toString()
			looked.getTpl1()
			looked.getTplN()
		}

		if err := looked.doResolve(full); err != nil {
			return err
		}
	}

	return thisParam.checkValidity()
}

// isWithinWhen tells if an identification parameter is somehow embedded in a "When" ID param
func (thisParam *IdentificationParameter) isWithinWhen() bool {
	if thisParam == nil {
		return false
	}

	if !thisParam.withinWhenResolved {
		thisParam.withinWhen = thisParam.isWhen || thisParam.parent.isWithinWhen()
		thisParam.withinWhenResolved = true
	}

	return thisParam.withinWhen
}

// isVerifiedBy returns true if the given object verifies this condition
func (thisCondition *ConditionalIDParameter) isVerifiedBy(ent *JsonEntity) bool {
	if ent == nil || len(ent.values) == 0 {
		return false
	}

	return fmt.Sprintf("%v", ent.values[thisCondition.Prop]) == thisCondition.Is
}
