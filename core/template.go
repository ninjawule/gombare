package core

import (
	"bytes"
	"fmt"
	"html/template"
	"runtime/debug"
	"sort"
	"strings"
)

//------------------------------------------------------------------------------
// Handling aliases for objects shown in the diffs, by using templates
//------------------------------------------------------------------------------

const objALIAS = "__alias__"

func (thisParam *IdentificationParameter) getTpl1() []string {
	if len(thisParam.Tpl1) > 0 {
		return thisParam.Tpl1
	}

	if p := thisParam.parent; p != nil && p.toString() == thisParam.toString() && len(p.getTpl1()) > 0 {
		thisParam.Tpl1 = p.getTpl1()
	}

	return thisParam.Tpl1
}

func (thisParam *IdentificationParameter) getTplN() []string {
	if len(thisParam.TplN) > 0 {
		return thisParam.TplN
	}

	if p := thisParam.parent; p != nil && p.toString() == thisParam.toString() && len(p.getTplN()) > 0 {
		thisParam.TplN = p.getTplN()
	}

	if thisParam.TplN == nil {
		thisParam.TplN = thisParam.getTpl1()
	}

	return thisParam.TplN
}

func (thisParam *IdentificationParameter) getTpl1String() string {
	return strings.Join(thisParam.getTpl1(), "")
}

func (thisParam *IdentificationParameter) getTplNString() string {
	return strings.Join(thisParam.getTplN(), "")
}

func (thisParam *IdentificationParameter) addAlias1(obj map[string]interface{}, currentPathValue string) {
	// not adding an alias twice
	if obj[objALIAS] != nil {
		return
	}

	// building the template, if not already done
	if thisParam.buildTpl1 == nil {
		var errParse error
		if thisParam.buildTpl1, errParse = template.New(thisParam.toString()).Funcs(template.FuncMap{
			"Display": display,
			"Slice":   slice,
		}).Parse(thisParam.getTpl1String()); errParse != nil {
			panic(fmt.Sprintf("Invalid template '%s' at path: %s. Cause: %s", thisParam.getTpl1String(), thisParam.toString(), errParse))
		}
	}

	var bytes bytes.Buffer
	if errRender := thisParam.buildTpl1.Execute(&bytes, obj); errRender != nil {
		panic(fmt.Sprintf("\n\nFailed to apply template '%s'\n\non object (at path: %s): %v.\n\nCause: %s", thisParam.getTpl1String(), currentPathValue, obj, errRender))
	}

	obj[objALIAS] = bytes.String()
}

func (thisParam *IdentificationParameter) addAliasN(objects []map[string]interface{}, currentPathValue string) {
	// not adding an alias twice
	if objects[0][objALIAS] != nil {
		return
	}

	// building the template, if not already done
	if thisParam.buildTplN == nil {
		var errParse error
		if thisParam.buildTplN, errParse = template.New(thisParam.toString()).Funcs(template.FuncMap{
			"Display": display,
			"Slice":   slice,
		}).Parse(thisParam.getTplNString()); errParse != nil {
			panic(fmt.Sprintf("Invalid template '%s' at path: %s. Cause: %s", thisParam.getTplNString(), thisParam.toString(), errParse))
		}
	}

	for _, obj := range objects {
		var bytes bytes.Buffer
		if errRender := thisParam.buildTplN.Execute(&bytes, obj); errRender != nil {
			panic(fmt.Sprintf("\n\nFailed to apply template '%s'\n\non object (at path: %s): %v.\n\nCause: %s", thisParam.getTplNString(), currentPathValue, obj, errRender))
		}

		obj[objALIAS] = bytes.String()
	}
}

func (thisParam *IdentificationParameter) getAlias(obj interface{}, currentPathValue string, options *ComparisonOptions) string {
	// we may already have an alias
	switch obj := obj.(type) {
	case map[string]interface{}:
		aliasObj, ok := obj[objALIAS]
		if ok {
			return aliasObj.(string)
		}

		// or... we haven't had the occasion to build it yet - so we build it and return it right away
		if thisParam != nil {
			if len(thisParam.getTpl1()) > 0 {
				thisParam.addAlias1(obj, currentPathValue)

				return obj[objALIAS].(string)
			} else if !options.AllowRaw {
				panic(fmt.Sprintf("No template configure to display an object at path: '%s'. This object: %v", currentPathValue, obj))
			}
		}

	case []map[string]interface{}:
		_, ok := obj[0][objALIAS]
		if ok {
			aliases := []string{}
			for _, elem := range obj {
				aliases = append(aliases, elem[objALIAS].(string))
			}

			return strings.Join(aliases, " ### ")
		}

		// or... we haven't had the occasion to build it yet - so we build it and return it right away
		if thisParam != nil {
			if len(thisParam.getTplN()) > 0 {
				thisParam.addAliasN(obj, currentPathValue)

				return thisParam.getAlias(obj, currentPathValue, options)
			} else if !options.AllowRaw {
				panic(fmt.Sprintf("No template configure to display an object at path: '%s'. This object: %v", currentPathValue, obj))
			}
		}

	case []interface{}:
		// we do something here if somehow there's an issue to detect the '[]map[string]interface{}' type
		if newMap, isMap := toMap(obj); isMap {
			return thisParam.getAlias(newMap, currentPathValue, options)
		}

		return ""
	}

	// no alias : we're going to use the object itself to display it
	return ""
}

func toMap(obj []interface{}) ([]map[string]interface{}, bool) {
	if _, ok := obj[0].(map[string]interface{}); ok {
		objMap := make([]map[string]interface{}, len(obj))
		for i, item := range obj {
			//nolint:errcheck
			objMap[i] = item.(map[string]interface{})
		}

		return objMap, true
	}

	return nil, false
}

func display(arg interface{}, path string, keys ...string) (result string) {
	if strings.TrimSpace(path) == "" {
		return displayObj(arg, nil, 0, keys...)
	}

	return displayObj(arg, strings.Split(path, "."), 0, keys...)
}

//nolint:cyclop,gocyclo,gocognit
func displayObj(arg interface{}, paths []string, pathIndex int, keys ...string) string {
	if len(keys) == 0 {
		return fmt.Sprintf("[no keys; using default display here] %v", arg)
	}

	// we've already crossed the whole to find the targeted objects => we're dealing with the keys now
	if pathIndex == len(paths) {
		switch arg := arg.(type) {
		case map[string]interface{}:
			var value string

			for _, key := range keys {
				switch strings.TrimSpace(key) {
				case "":
					value = value + " "
				case ":", "=", ".", "[", "]", "(", ")", "{", "}", "-":
					value = value + key
				default:
					// the key is just a label here
					if key[0] == '\'' {
						value = fmt.Sprintf("%s%s", value, key[1:len(key)-1])
					} else {
						// a key can have a "path1.path2.etc.property" form, to refer to an indirect property
						currentObj := arg
						localKeys := strings.Split(key, ".")
						last := len(localKeys) - 1
						for i, subKey := range localKeys {
							if i < last {
								if newObj, ok := currentObj[subKey].(map[string]interface{}); ok {
									currentObj = newObj
								} else {
									panic(fmt.Sprintf("sub-key '%s' does not refer to a JSON entity (map[string]interface{}) on object: %v", subKey, currentObj))
								}
							} else {
								obj, ok := currentObj[subKey].(map[string]interface{})
								if val, hasText := obj["#text"]; ok && hasText {
									value = fmt.Sprintf("%s%v", value, val)
								} else {
									if target, ok := currentObj[subKey]; ok {
										value = fmt.Sprintf("%s%v", value, target)
									} else {
										value = fmt.Sprintf("%s(%s)", value, key)
									}
								}
							}
						}
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

		case []interface{}:
			// we have to force the type "[]interface{}" into "map[string]interface{}" here
			newMap, _ := toMap(arg)

			return displayObj(newMap, paths, pathIndex, keys...)

		default:
			panic(fmt.Sprintf("[unhandled case (keys); using default display here] %v (%T).\n\nPaths: %v, index: %d.\n\nStack: %s",
				arg, arg, paths, pathIndex, string(debug.Stack())))
		}
	}

	// else, we've yet to ge deeper into the data ==> we're going down the paths here
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

	case []interface{}:
		newMap, _ := toMap(arg)

		return displayObj(newMap, paths, pathIndex, keys...)

	default:
		panic(fmt.Sprintf("[unhandled case (path); using default display here] %v (%T).\n\nPaths: %v, index: %d.\n\nStack: %s",
			arg, arg, paths, pathIndex, string(debug.Stack())))
	}
}

func slice(obj interface{}) interface{} {
	switch obj := obj.(type) {
	case []interface{}, []map[string]interface{}:
		return obj
	}

	return []interface{}{obj}
}
