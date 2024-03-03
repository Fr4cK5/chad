package parse

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// A datastructure representing the result
// from parsing input.
type ParseResult struct {
	Flags map[string]string
	Positionals []string
}

// Get a string from the parsed positional arguments according to some index `idx`
func (slf *ParseResult) GetStringIndex(idx int) (*string, error) {
	length := len(slf.Positionals)
	if checkBounds(idx, length) {
		return &slf.Positionals[idx], nil
	}

	return nil, fmt.Errorf("GetStringIndex() index %v is out of bounds for lenght %v", idx, length)
}

// Get an int from the parsed positional arguments according to some index `idx`
func (slf *ParseResult) GetIntIndex(idx int) (*int, error) {
	length := len(slf.Positionals)
	if checkBounds(idx, length) {
		value, err := strconv.Atoi(slf.Positionals[idx])
		if err != nil {
			return nil, fmt.Errorf("unable to parse value '%v' at index '%v' to integer type", value, idx)
		}

		return &value, nil
	}

	return nil, fmt.Errorf("GetIntIndex() index %v is out of bounds for lenght %v", idx, length)
}

// Get a float from the parsed positional arguments according to some index `idx`
func (slf *ParseResult) GetFloatIndex(idx int) (*float64, error) {
	length := len(slf.Positionals)
	if checkBounds(idx, length) {
		value, err := strconv.ParseFloat(slf.Positionals[idx], 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse value '%v' at index '%v' to float type", value, idx)
		}

		return &value, nil
	}

	return nil, fmt.Errorf("GetFloatIndex() index %v is out of bounds for lenght %v", idx, length)
}

// Get a string from the parsed flags by some key `key`
func (slf *ParseResult) GetStringFlag(key string) (*string, error) {
	if value, ok := slf.Flags[key]; ok {
		retval := strings.Clone(value)
		return &retval, nil
	}

	return nil, fmt.Errorf("no value for key %v", key)
}

// Get an int from the parsed flags by some key `key`
func (slf *ParseResult) GetIntFlag(key string) (*int, error) {
	if value, ok := slf.Flags[key]; ok {
		retval, err := strconv.Atoi(value)
		if err != nil {
			return nil, fmt.Errorf("unable to parse value '%v' from key '%v' to integer type", value, key)
		}
		return &retval, nil
	}

	return nil, fmt.Errorf("no value for key %v", key)
}

// Get a float from the parsed flags by some key `key`
func (slf *ParseResult) GetFloatFlag(key string) (*float64, error) {
	if value, ok := slf.Flags[key]; ok {
		retval, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to parse value '%v' from key '%v' to float64 type", value, key)
		}
		return &retval, nil
	}

	return nil, fmt.Errorf("no value for key %v", key)
}

// Get a bool from the parsed flags by some key `key`
func (slf *ParseResult) GetBoolFlag(key string) (*bool, error) {
	if _, ok := slf.Flags[key]; ok {
		return &ok, nil
	}

	return nil, fmt.Errorf("key %v not found in boolean flags", key)
}

func checkBounds(idx, size int) bool {
	return idx >= 0 && idx < size
}

// Parse arguments from a string.
// The input string is preprocessed with respect to strings within the input string.
// see parse.toFilteredSlice(string) for more information.
func ParseFromString(input string) *ParseResult {
	fixed, err := toFilteredSlice(input)
	if err != nil {
		panic(err)
	}
	return parse(fixed)
}

// Parse args from string slice `input`
// If you're unsure about the correctness of your input,
// see parse.toFilteredSlice(string) for more information.
func ParseFromSlice(input []string) *ParseResult {
	return parse(&input)
}

// Parse from os.Args
func Parse() *ParseResult {
	os_args := os.Args[1:]
	return parse(&os_args)
}

// Returns a guaranteed valid but possibly empty, `ParseResult`.
func parse(input *[]string) *ParseResult {
	flags, positional, err := parseFlags(input)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return &ParseResult{
		Flags: *flags,
		Positionals: *positional,
	}
}

// Parse the flags of some input string slice.
// Following scheme is used:
//	--file filename.txt -> file: "filename.txt";
//		Assigned to arg `file`.
//	-file filename.txt -> f: "", i: "", l: "", e: "filename.txt";
//		Assigned to arg `e`, since its the last one in the arg-stack.
//	-V -> V: "";
//		Arg `V` is an empty string. When getting its value using GetBoolFlag("V"),
//		you'd get `true` since it exists.
//		GetBoolFlag(string) can also be used to check the presence of any flag.
func parseFlags(input *[]string) (*map[string]string, *[]string, error) {
	assigned := make(map[string]string)
	positional := make([]string, 0)

	if len(*input) == 0 {
		return &assigned, &positional, nil
	}
	
	skip_next := false

	for i, item := range *input {
		if skip_next {
			skip_next = false
			continue
		}

		if strings.HasPrefix(item, "--") && len(item) >= 3 {
			if i+1 != len(*input) && !strings.HasPrefix((*input)[i+1], "-") {
				skip_next = true
				assigned[item[2:]] = (*input)[i+1]
			} else {
				assigned[item[2:]] = ""
			}
		} else if strings.HasPrefix(item, "-") && len(item) >= 2 {
			flags := item[1:]
			for j, c := range flags {
				if c == '-' {
					return nil, nil, fmt.Errorf("tried to parse illegal character '-' as a parameter name in stack '%v'", flags)
				}

				// If it's the last flag of the stack (eg. 'L' in -rfL), and the original input still got more items
				// we assign the next item to the last flag 'L'.
				result := ""
				if j+1 == len(flags) && i+1 != len(*input) && !strings.HasPrefix((*input)[i+1], "-") {
					skip_next = true
					result = (*input)[i+1]
				} 

				key := string(c)

				// We only want to add the kv-pair if it doesn't already exist.
				// This way, we handle double definitions of parameters be it with or without values.
				if _, ok := assigned[key]; !ok {
					assigned[key] = result
				} else {
					return nil, nil, fmt.Errorf("double-definition of paramater '%v' in parameter stack '%v'", key, flags)
				}
			}
		} else {
			positional = append(positional, item)
		}
	}

	return &assigned, &positional, nil
}

// Returns the input string split up by spaces ' ' while
// respecting any present strings
// 
// Example:
//	-d Hello --file "Some text file.txt" 'Hello there!'
// Would result in:
//	[-d, Hello, --file, Some text file.txt, Hello there!]
func toFilteredSlice(input string) (*[]string, error) {
	strs := make([]string, 0)

	start := 0
	current_delim := ' '
	in_str := false

	normal_start := 0

	for i, r := range input {
		// String parsing
		if !in_str && isValidDelim(r) {
			in_str = true
			start = i
			current_delim = r
			continue
		} else if in_str && current_delim == r && input[i-1] != '\\' {
			in_str = false
			normal_start = i+1
			strs = append(strs, input[start+1:i])
			continue
		} else if in_str {
			continue
		}

		if i+1 == len(input) {
			strs = append(strs, input[normal_start:i+1])
		}

		// Non-String parsing
		if r != ' ' {
			if input[normal_start] == ' ' && normal_start < i {
				normal_start = i
			}
			continue
		}

		strs = append(strs, input[normal_start:i])
		normal_start = i+1
	}

	if in_str {
		return nil, errors.New("unexpected end of input while parsing string")
	}

	return &strs, nil
}

func isValidDelim(r rune) bool {
	return r == '"' ||
		r == '\'' ||
		r == '`'
}
