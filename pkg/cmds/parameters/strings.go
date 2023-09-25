package parameters

import (
	"fmt"
	"strings"
)

// GatherFlagsFromStringList is a function that parses command line flags according to a list of parameter definitions.
// It returns a map of parameter names to parsed values.
//
// ### Usage
//
// The function takes two arguments:
//
// - `args`: a slice of strings representing the command line arguments.
// - `params`: a slice of `*ParameterDefinition` representing the parameter definitions.
//
// The function returns a map where the keys are the parameter names and the values are the parsed values.
// If a flag is not recognized or its value cannot be parsed, an error is returned.
//
// ### Internals
//
// The function first creates a map of possible flag names from the list of parameter definitions.
// It then iterates over the command line arguments. If an argument starts with `--` or `-`,
// it is considered a flag. The function then checks if the flag is recognized by looking it up in
// the map of flag names. If the flag is recognized, its value is collected.
//
// The system sets the flag to "true" automatically if it's a boolean flag. If a flag repeats,
// the system collects all its values in a slice. Once the system has collected all flags and their raw values,
// it parses the raw values based on the parameter definitions.
//
// The `ParseParameter` method of the corresponding `ParameterDefinition` performs the parsing.
// This method receives a slice of strings and returns the parsed value and an error.
// The system stores the parsed values in the result map.
//
// ### Example
//
// Here is an example of how to use the function:
//
// ```go
//
//	params := []*ParameterDefinition{
//	   {Name: "verbose", ShortFlag: "v", Type: ParameterTypeBool},
//	   {Name: "output", ShortFlag: "o", Type: ParameterTypeString},
//	}
//
// args := []string{"--verbose", "-o", "file.txt"}
// result, err := GatherFlagsFromStringList(args, params)
//
//	if err != nil {
//	   log.Fatal(err)
//	}
//
// fmt.Println(result) // prints: map[verbose:true output:file.txt]
// ```
//
// In this example, the function parses the `--verbose` and `-o` flags according to the provided parameter definitions. The `--verbose` flag is a boolean flag and is set to "true". The `-o` flag is a string flag and its value is "file.txt".
func GatherFlagsFromStringList(
	args []string,
	params []*ParameterDefinition,
) (map[string]interface{}, error) {
	flagMap := make(map[string]*ParameterDefinition)
	for _, param := range params {
		flagMap[param.Name] = param
		if param.ShortFlag != "" {
			flagMap[param.ShortFlag] = param
		}
	}

	rawValues := make(map[string][]string)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		var flagName string
		var param *ParameterDefinition
		var ok bool
		if strings.HasPrefix(arg, "--") {
			flagName = arg[2:]
			param, ok = flagMap[flagName]
			if !ok {
				return nil, fmt.Errorf("unknown flag: --%s", flagName)
			}
		} else if strings.HasPrefix(arg, "-") {
			flagName = arg[1:]
			param, ok = flagMap[flagName]
			if !ok {
				return nil, fmt.Errorf("unknown flag: -%s", flagName)
			}
		} else {
			continue
		}

		if param.Type == ParameterTypeBool {
			rawValues[param.Name] = append(rawValues[param.Name], "true")
		} else {
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing value for flag: -%s", flagName)
			}
			value := args[i+1]
			i++ // skip next arg
			if IsListParameter(param.Type) {
				value = strings.Trim(value, "[]")
				values := strings.Split(value, ",")
				rawValues[param.Name] = append(rawValues[param.Name], values...)
				continue
			}
			rawValues[param.Name] = append(rawValues[param.Name], value)
		}
	}

	result := make(map[string]interface{})
	for paramName, values := range rawValues {
		param := flagMap[paramName]
		parsedValue, err := param.ParseParameter(values)
		if err != nil {
			return nil, fmt.Errorf("invalid value for flag --%s: %v", paramName, err)
		}
		result[param.Name] = parsedValue
	}
	return result, nil
}
