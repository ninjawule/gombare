package core

import (
	"bytes"
	"fmt"
	"html/template"
	"sort"
	"strings"
)

//------------------------------------------------------------------------------
// Handling aliases for objects shown in the diffs, by using templates
//------------------------------------------------------------------------------

const objALIAS = "__alias__"

func (thisParam *IdentificationParameter) getTpl() []string {
	if len(thisParam.Tpl) > 0 {
		return thisParam.Tpl
	}

	if p := thisParam.parent; p != nil && p.toString() == thisParam.toString() && len(p.getTpl()) > 0 {
		thisParam.Tpl = p.getTpl()
	}

	return thisParam.Tpl
}

func (thisParam *IdentificationParameter) getTplString() string {
	return strings.Join(thisParam.getTpl(), "")
}

func (thisParam *IdentificationParameter) addAlias(obj map[string]interface{}) {
	if thisParam.buildTpl == nil {
		var errParse error
		if thisParam.buildTpl, errParse = template.New(thisParam.toString()).Funcs(template.FuncMap{
			"Display": display,
		}).Parse(thisParam.getTplString()); errParse != nil {
			panic(fmt.Sprintf("Invalid template '%s' at path: %s. Cause: %s", thisParam.getTplString(), thisParam.toString(), errParse))
		}
	}

	var bytes bytes.Buffer
	if errRender := thisParam.buildTpl.Execute(&bytes, obj); errRender != nil {
		panic(fmt.Sprintf("Failed to apply template '%s' on object: %v. Cause: %s", thisParam.getTplString(), obj, errRender))
	}

	obj[objALIAS] = bytes.String()
}

func (thisParam *IdentificationParameter) getAlias(obj interface{}) string {
	// we may already have an alias
	switch obj := obj.(type) {
	case map[string]interface{}:
		aliasObj, ok := obj[objALIAS]
		if ok {
			return aliasObj.(string)
		}

		// or... we haven't had the occasion to build it yet - so we build it and return it right away
		if thisParam != nil && len(thisParam.getTpl()) > 0 {
			thisParam.addAlias(obj)

			return obj[objALIAS].(string)
		}
	}

	// no alias : we're going to use the object itself to display it
	return ""
}

func display(arg interface{}, path string, keys ...string) (result string) {
	if strings.TrimSpace(path) == "" {
		return displayObj(arg, nil, 0, keys...)
	}

	return displayObj(arg, strings.Split(path, "."), 0, keys...)
}

//nolint:cyclop,gocyclo
func displayObj(arg interface{}, paths []string, pathIndex int, keys ...string) string {
	if len(keys) == 0 {
		return fmt.Sprintf("[no keys; using default display here] %v", arg)
	}

	// we've already crossed the whole to find the targeted objects
	if pathIndex == len(paths) {
		switch arg := arg.(type) {
		case map[string]interface{}:
			var value string

			for _, key := range keys {
				switch strings.TrimSpace(key) {
				case "":
					value = value + " "
				case ":", "=", ".":
					value = value + key
				default:
					obj, ok := arg[key].(map[string]interface{})
					if val, hasText := obj["#text"]; ok && hasText {
						value = fmt.Sprintf("%s%v", value, val)
					} else {
						value = fmt.Sprintf("%s%v", value, arg[key])
					}
				}
			}

			return value

		case []map[string]interface{}:
			values := []string{}
			for _, singleObj := range arg {
				values = append(values, displayObj(singleObj, paths, pathIndex, keys...))
			}

			sort.Strings(values)

			return strings.Join(values, " | ")

		default:
			return fmt.Sprintf("[unhandled case; using default display here] %v", arg)
		}
	}

	// else, we've yet to ge deeper into the data
	switch arg := arg.(type) {
	case map[string]interface{}:
		return displayObj(arg[paths[pathIndex]], paths, pathIndex+1, keys...)

	case []map[string]interface{}:
		values := []string{}

		for _, singleObj := range arg {
			values = append(values, displayObj(singleObj[paths[pathIndex]], paths, pathIndex+1, keys...))
		}

		sort.Strings(values)

		return strings.Join(values, " | ")

	default:
		return fmt.Sprintf("[unhandled case; using default display here] %v", arg)
	}
}
