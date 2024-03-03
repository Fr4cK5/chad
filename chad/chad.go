package chad

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
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
	argsRegistered bool
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
		make(map[string]Arg),
		nil,
		0,
		false,
	}
}

// Register the arguments to be parsed.
// This allows for the validation of the arguments.
func (slf *Chad) RegisterArgs(args []Arg, positionalArgCount int) {
	slf.argsRegistered = true
	slf.ExpectedPositionalCount = positionalArgCount
	for _, arg := range args {
		slf.DefinedFlags[arg.Name] = arg
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
}

// Parse from an input string slice
func (slf *Chad) ParseFromSlice(input []string) {
	slf.checkRegistered()
	slf.parse(parse.ParseFromSlice(input))
}

// Parse from os.Args
func (slf *Chad) Parse() {
	slf.checkRegistered()
	slf.parse(parse.Parse())
}

// This is cursed, but it should* work!
func (slf *Chad) parse(parsed_args *parse.ParseResult) {

	if len(parsed_args.Positionals) != slf.ExpectedPositionalCount {
		fmt.Printf("Received invalid amount of positional arguments. Expected %v, got %v\n", slf.ExpectedPositionalCount, len(parsed_args.Positionals))
		os.Exit(1)
	}

	slf.Result = &parse.ParseResult{
		Flags: make(map[string]string),
		Positionals: make([]string, 0),
	}

	check_supplied_but_undefined_flags:
	for parsed := range parsed_args.Flags {
		for defined := range slf.DefinedFlags {
			if parsed == defined {
				continue check_supplied_but_undefined_flags
			}
		}
		fmt.Printf("An unknown flag '%v' was supplied.\n", parsed)
		os.Exit(1)
	}

	for arg, value := range parsed_args.Flags {
		default_value := slf.DefinedFlags[arg].DefaultValue
		if !isTypeOk(default_value, value) {
			fmt.Printf("Flag '%v' expects input of type '%v' but recieved 'string'.\n", arg, reflect.TypeOf(default_value).Name())
			os.Exit(1)
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
					slf.Result.Flags[arg.Name] = ""
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

		fmt.Printf("Did not receive required flag '%v'.\n", arg.Name)
		os.Exit(1)
	}

	for k, v := range parsed_args.Flags {
		if _, ok := slf.Result.Flags[k]; !ok {
			slf.Result.Flags[k] = v
		}
	}

	slf.Result.Positionals = append(slf.Result.Positionals, parsed_args.Positionals...)
}

func (slf *Chad) IsFlagPresent(key string) bool {
	_, ok := slf.Result.Flags[key]
	return ok
}

func (slf *Chad) IsFlagDefault(key string) bool {
	if flag, ok := slf.Result.Flags[key]; ok {
		return flag == slf.DefinedFlags[key].DefaultValue
	}
	return false
}

func (slf *Chad) GetStringIndex(idx int) string {
	value, err := slf.Result.GetStringIndex(idx)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return *value
}

func (slf *Chad) GetIntIndex(idx int) int {
	value, err := slf.Result.GetIntIndex(idx)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return *value
}

func (slf *Chad) GetFloatIndex(idx int) float64 {
	value, err := slf.Result.GetFloatIndex(idx)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return *value
}

func (slf *Chad) GetStringFlag(key string) string {
	value, err := slf.Result.GetStringFlag(key)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return *value
}

func (slf *Chad) GetIntFlag(key string) int {
	value, err := slf.Result.GetIntFlag(key)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return *value
}

func (slf *Chad) GetFloatFlag(key string) float64 {
	value, err := slf.Result.GetFloatFlag(key)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return *value
}

func (slf *Chad) GetBoolFlag(key string) bool {
	value, err := slf.Result.GetBoolFlag(key)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return *value
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
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

func isValidFloat(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
