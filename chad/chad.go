package chad

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"
	"unicode"

	"github.com/Fr4cK5/chad/internal/parse"
)

// A datastructure representing a defined argument.
// Used in the validation step.
type Arg struct {
	Name, Help string
	DefaultValue interface{}
	Required bool
}

func NewArg(name, help string, defaultValue interface{}, required bool) *Arg {
	return &Arg{
		Name: name,
		Help: help,
		DefaultValue: defaultValue,
		Required: required,
	}
}

// A datastructure representing a true Chad.
// The Chad holds all important data for validation
// and parsed results after validation.
type Chad struct {
	DefinedFlags map[string]Arg
	Result *parse.ParseResult
	ExpectedPositionalCount int
	PositionalNames []string
	argsRegistered bool
	originalResult *parse.ParseResult
}

// Usage:
//	chad := chad.NewChad()
//	chad.RegisterArgs(
//		[]chad.Arg{ // These are our arguments.
//			*chad.NewArg("file", "The file to be read", "default_filename", true),
//		},
//		0, // The number of positional arguments we expect.
//	)
//	chad.Parse() // or any other parse<...> function.
// Before parsing anything, you have to register the arguments
// for the validation process to be able to execute.
func NewChad() *Chad {
	return &Chad{
		DefinedFlags: make(map[string]Arg),
		Result: nil,
		originalResult: nil,
		ExpectedPositionalCount: 0,
		PositionalNames: nil,
		argsRegistered: false,
	}
}

// Register the arguments to be parsed.
// This allows for the validation of the arguments.
func (slf *Chad) RegisterArgs(args []Arg, positionalNames []string) {
	slf.argsRegistered = true
	slf.ExpectedPositionalCount = len(positionalNames)
	slf.PositionalNames = positionalNames

	// Sneak the help flag in there!
	slf.DefinedFlags["help"] = *NewArg("help", "Print help", false, false)

	for _, arg := range args {
		if _, ok := slf.DefinedFlags[arg.Name]; !ok {
			slf.DefinedFlags[arg.Name] = arg
		} else {
			panic(fmt.Errorf("tried to create flag '%v' twice", arg.Name))
		}
	}

	for _, name := range slf.PositionalNames {
		if _, ok := slf.DefinedFlags[name]; ok {
			panic(fmt.Errorf("arg '%v' is present in flags as well as positional args", name))
		}
	}
}

// If no args were registered we can't really validate them so there goes that >.<
func (slf *Chad) checkRegistered() {
	if !slf.argsRegistered {
		panic("cannot validate parsed args without registering any; use Chad.RegisterArgs(...) to do so.")
	}
}

// Parse from an input string
func (slf *Chad) ParseFromString(input string) {
	slf.checkRegistered()
	slf.parse(parse.ParseFromString(input))
	slf.exitIfFlagProvided()
}

// Parse from an input string slice
func (slf *Chad) ParseFromSlice(input []string) {
	slf.checkRegistered()
	slf.parse(parse.ParseFromSlice(input))
	slf.exitIfFlagProvided()
}

// Parse from os.Args
func (slf *Chad) Parse() {
	slf.checkRegistered()
	slf.parse(parse.Parse())
	slf.exitIfFlagProvided()
}

// This is cursed, but it should* work!
func (slf *Chad) parse(parsed_args *parse.ParseResult) {

	slf.Result = &parse.ParseResult{
		Flags: make(map[string]string),
		Positionals: make([]string, 0),
	}
	// We need to keep track of the original parse result in order to query if flags have been provided or not.
	slf.originalResult = parsed_args

	// the user provided --help
	if slf.IsFlagPresent("help") {
		slf.exitWithHelp("")
	}

	if len(parsed_args.Positionals) != slf.ExpectedPositionalCount {
		slf.exitWithHelp(fmt.Sprintf("Received invalid amount of positional arguments. Expected %v, got %v.", slf.ExpectedPositionalCount, len(parsed_args.Positionals)))
	}

	check_supplied_but_undefined_flags:
	for parsed := range parsed_args.Flags {
		for defined := range slf.DefinedFlags {
			if parsed == defined {
				continue check_supplied_but_undefined_flags
			}
		}
		slf.exitWithHelp(fmt.Sprintf("An unknown flag '%v' was supplied.", parsed))
	}

	for arg, value := range parsed_args.Flags {
		default_value := slf.DefinedFlags[arg].DefaultValue
		if !isTypeOk(default_value, value) {
			slf.exitWithHelp(fmt.Sprintf("Flag '%v' expects input of type '%v' but recieved 'string'.", arg, reflect.TypeOf(default_value).Name()))
		}
	}

	flag_check:
	for _, arg := range slf.DefinedFlags {
		// If the arg isn't required but still present,
		// we want to put it into the results anyways.
		if !arg.Required {
			if value, ok := parsed_args.Flags[arg.Name]; ok {
				// Not required but present
				slf.Result.Flags[arg.Name] = value
			} else {
				// Not required and not present, which means we need to parse the default value.
				def_value := slf.DefinedFlags[arg.Name].DefaultValue
				switch value := def_value.(type) {
				case int, int8, int16, int32, int64:
					slf.Result.Flags[arg.Name] = fmt.Sprintf("%d", value)
				case uint, uint8, uint16, uint32, uint64:
					slf.Result.Flags[arg.Name] = fmt.Sprintf("%d", value)
				case float32, float64:
					slf.Result.Flags[arg.Name] = fmt.Sprintf("%f", value)
				case bool:
					if value {
						slf.Result.Flags[arg.Name] = "true"
					} else {
						slf.Result.Flags[arg.Name] = "false"
					}
				case string:
					slf.Result.Flags[arg.Name] = value
				default:
					panic(fmt.Errorf("encountered unknown type while trying to parse default value of defined argument %v", slf.DefinedFlags[arg.Name]))
				}
				
			}
			continue flag_check
		}

		for name := range parsed_args.Flags {
			if name == arg.Name {
				continue flag_check
			}
		}

		slf.exitWithHelp(fmt.Sprintf("Did not receive required flag '%v'.", arg.Name))
	}

	for k, v := range parsed_args.Flags {
		if _, ok := slf.Result.Flags[k]; !ok {
			slf.Result.Flags[k] = v
		}
	}

	slf.Result.Positionals = append(slf.Result.Positionals, parsed_args.Positionals...)
}

// Check if a flag is present in the parsed arguments.
func (slf *Chad) IsFlagPresent(key string) bool {
	_, ok := slf.originalResult.Flags[key]
	return ok
}

// Check if a flag retained its default value.
func (slf *Chad) IsFlagDefault(key string) bool {
	if result_value, ok := slf.Result.Flags[key]; ok {
		def_value := slf.DefinedFlags[key].DefaultValue

		switch v := def_value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return result_value == fmt.Sprintf("%d", v)
		case float32, float64:
			return result_value == fmt.Sprintf("%f", v)
		case bool:
			return v
		case string:
			return v == result_value
		}
	}
	return false
}

// Get a positional value my its index in form of a string
func (slf *Chad) StringIndex(idx int) string {
	value, err := slf.Result.GetStringIndex(idx)
	if err != nil {
		slf.exitWithHelp(err.Error())
	}
	return *value
}

// Get a positional value my its index in form of an int
func (slf *Chad) IntIndex(idx int) int {
	value, err := slf.Result.GetIntIndex(idx)
	if err != nil {
		slf.exitWithHelp(err.Error())
	}
	return *value
}

// Get a positional value my its index in form of a float64
func (slf *Chad) FloatIndex(idx int) float64 {
	value, err := slf.Result.GetFloatIndex(idx)
	if err != nil {
		slf.exitWithHelp(err.Error())
	}
	return *value
}

// Get a positional value my its name in form of a string
func (slf *Chad) StringPosName(name string) string {
	idx := slf.posIndexByName(name)
	if idx == -1 {
		panic(fmt.Errorf("positional arg '%v' not found", name))
	}

	return slf.StringIndex(idx)
}

// Get a positional value my its name in form of an int
func (slf *Chad) IntPosName(name string) int {
	idx := slf.posIndexByName(name)
	if idx == -1 {
		panic(fmt.Errorf("positional arg '%v' not found", name))
	}

	return slf.IntIndex(idx)
}

// Get a positional value my its name in form of a float64
func (slf *Chad) FloatPosName(name string) float64 {
	idx := slf.posIndexByName(name)
	if idx == -1 {
		panic(fmt.Errorf("positional arg '%v' not found", name))
	}

	return slf.FloatIndex(idx)
}

// Get a flag's value by its name in form of a string
func (slf *Chad) StringFlag(key string) string {
	value, err := slf.Result.GetStringFlag(key)
	if err != nil {
		slf.exitWithHelp(err.Error())
	}
	return *value
}

// Get a flag's value by its name in form of an int
func (slf *Chad) IntFlag(key string) int {
	value, err := slf.Result.GetIntFlag(key)
	if err != nil {
		slf.exitWithHelp(err.Error())
	}
	return *value
}

// Get a flag's value by its name in form of a float64
func (slf *Chad) FloatFlag(key string) float64 {
	value, err := slf.Result.GetFloatFlag(key)
	if err != nil {
		slf.exitWithHelp(err.Error())
	}
	return *value
}

// Check if a flag is supplied, synonymous to Chad.IsFlagPresent(key)
func (slf *Chad) BoolFlag(key string) bool {
	return slf.IsFlagPresent(key)
}

func isTypeOk(expected interface{}, actual string) bool {
	switch expected.(type) {
	case string, bool:
		return true
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return isValidInt(actual)
	case float32, float64:
		return isValidFloat(actual)
	default:
		return false
	}
}

func isValidInt(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) && rune(c) != '-' && rune(c) != '+' {
			return false
		}
	}
	return true
}

func isValidFloat(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func (slf *Chad) genFlagHelp() string {

	buf_slice := make([]byte, 0, 1e4)
	buf := bytes.NewBuffer(buf_slice)
	writer := tabwriter.NewWriter(buf, 0, 0, 1, ' ', 0)

	// Put the required args first,
	sorted := make([]Arg, 0, len(slf.DefinedFlags))
	for _, arg := range slf.DefinedFlags {
		if !arg.Required || arg.Name == "help" {
			continue
		}
		sorted = append(sorted, arg)
	}
	// the optional ones second,
	for _, arg := range slf.DefinedFlags {
		if arg.Required || arg.Name == "help" {
			continue
		}
		sorted = append(sorted, arg)
	}
	// the help flag last.
	sorted = append(sorted, slf.DefinedFlags["help"])

	for _, arg := range sorted {
		var default_value string
		switch value := arg.DefaultValue.(type) {
		case string:
			default_value = "\"" + default_value + "\""

		// Is this necessary?
		case bool:
			if value {
				default_value = "true"
			} else {
				default_value = "false"
			}
		case int, int8, int16, int32, int64,
				uint, uint8, uint16, uint32, uint64:
			default_value = fmt.Sprintf("%d", value)
		case float32, float64:
			default_value = fmt.Sprintf("%f", value)
		}

		var arg_name string
		if len(arg.Name) > 1 {
			arg_name = "--" + arg.Name
		} else {
			arg_name = "-" + arg.Name
		}

		var extra string
		if arg.Required {
			extra = "Required"
		} else {
			extra = fmt.Sprintf("Default = %v", default_value) 
		}

		fmt.Fprintf(writer, "    %v\t\t%v\t\t[%v]\n", arg_name, arg.Help, extra)
	}

	writer.Flush()

	return buf.String()
}

func (slf *Chad) genPositionals() string {
	positionals := ""
	for i, str := range slf.PositionalNames {
		positionals += "<" + strings.ToUpper(str) + ">"
		if i < len(slf.PositionalNames)-1 {
			positionals += " "
		}
	}
	return positionals
}

func (slf *Chad) genHelp() string {
	help := ""
	binary := getBinaryName()
	arg_help := slf.genFlagHelp()

	help += fmt.Sprintf("Usage: %v", binary)
	if slf.ExpectedPositionalCount > 0 {
		help += fmt.Sprintf(" %v", slf.genPositionals())
	}
	help += " [Flags]\n\n"

	help += "Flags:\n"
	help += arg_help

	return help
}

func (slf *Chad) exitWithHelp(err string) {
	if strings.Trim(err, " \t\r\n") != "" {
		fmt.Println("Error:")
		fmt.Printf("    %v\n\n", err)
	}
	fmt.Println(slf.genHelp())
	os.Exit(1)
}

func (slf *Chad) exitIfFlagProvided() {
	if _, ok := slf.originalResult.Flags["h"]; ok {
		slf.exitWithHelp("")
	}
}

func getBinaryName() string {
	exec_path := os.Args[0]
	path_parts := strings.Split(exec_path, string(os.PathSeparator))
	return path_parts[len(path_parts)-1]
}

func (slf *Chad) posIndexByName(name string) int {
	idx := -1
	for i, iter_name := range slf.PositionalNames {
		if name == iter_name {
			idx = i
		}
	}
	return idx
}
