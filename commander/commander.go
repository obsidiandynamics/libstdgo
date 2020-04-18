// Package commander contains a simple command-line argument parser that supports flags, switches and free-form
// trailing (and mid-stream) arguments, and provides methods for their retrieval.
//
// Commander provides similar
// functionality to the flags package, but also supports 'schemaless' argument parsing, where the structure of
// arguments might not be known up front. It supports user-defined, key-value style arguments — for example,
// as 'myapp -key1=value1 --key2 value2', where key1 and key2 could be anything.
package commander

import (
	"fmt"
	"strings"
)

// Part is a tuple of parsed arguments.
//
// For named arguments — for example, the argument pair '-key=value' or '-key value', the
// Name field is 'key' (i.e. without the leading dash or pair of dashes) and the Value field is
// simply 'value'.
// For free-form (a.k.a. trailing) arguments — for example, 'file1.txt', the Name field is a blank string and the Value
// field contains the argument.
type Part struct {
	Name, Value string
}

// String returns a textual representation of the part.
func (p Part) String() string {
	return "Part[Name=" + p.Name + ", Value=" + p.Value + "]"
}

// IsFreeForm return true if this is a free-form part.
func (p Part) IsFreeForm() bool {
	return p.Name == ""
}

// Parts is a slice of Part tuples.
type Parts []Part

// PartsMap keys unique part names to the values. For free-form values, the key is FreeForm.
type PartsMap map[string][]string

// FreeForm is a key in PartsMap for free-form (trailing) values.
const FreeForm = "__free_form"

// Value obtains a single value for the given name, returning the default value if none exist.
// If the map contains two or more values for the given name, the first value is returned as
// well as an error.
func (pm PartsMap) Value(name string, def string) (string, error) {
	values := pm[name]
	switch {
	case len(values) == 1:
		return values[0], nil
	case len(values) == 0:
		return def, nil
	default:
		return values[0], fmt.Errorf("too many arguments: expected one or none, got %d", len(values))
	}
}

// Mappify transforms the parsed Parts into a PartsMap for convenient retrieval of argument values.
func (parts Parts) Mappify() PartsMap {
	partsMap := PartsMap{}
	for _, p := range parts {
		var key string
		if p.IsFreeForm() {
			key = FreeForm
		} else {
			key = p.Name
		}
		partsMap[key] = append(partsMap[key], p.Value)
	}
	return partsMap
}

// Parse processes the given cmdArgs into a Parts slice. No error is returned as parsing is schema-less; the parser
// extracts all flags, switches and free-form values that may be present.
func Parse(cmdArgs []string) Parts {
	len := len(cmdArgs)
	args := make([]Part, 0, len/2)
	for i := 0; i < len; i++ {
		currArg := cmdArgs[i]
		if currDashes := dashes(currArg); currDashes > 0 {
			split := strings.IndexByte(currArg, '=')
			if split != -1 {
				// In the form '-arg=value'
				args = append(args, Part{currArg[currDashes:split], currArg[split+1:]})
			} else if i < len-1 {
				peekArg := cmdArgs[i+1]
				if peekDashes := dashes(peekArg); peekDashes > 0 {
					// In the form '-arg -arg'
					args = append(args, Part{currArg[currDashes:], "true"})
				} else {
					// In the form '-arg value'
					args = append(args, Part{currArg[currDashes:], peekArg})
					i++
				}
			} else {
				// In the form '-arg'
				args = append(args, Part{currArg[currDashes:], "true"})
			}
		} else {
			// Standalone token
			args = append(args, Part{"", currArg})
		}
	}
	return args
}

// Returns the number of leading dashes contained in a given argument, up to a maximum of two. If the argument
// has three or more leading dashes, it is reported as containing no dashes, thereby being treated as something
// other than a switch or a flag. If the argument is just a dash (or double-dash) on its own, it is also reported
// as having no dashes.
func dashes(cmdArg string) int {
	length := len(cmdArg)
	switch {
	case length >= 3 && cmdArg[0] == '-' && cmdArg[1] == '-' && cmdArg[2] != '-':
		return 2
	case length >= 2 && cmdArg[0] == '-' && cmdArg[1] != '-':
		return 1
	default:
		return 0
	}
}
