# Chad
An absolutely **BASED GIGACHAD** of a Go command line parser

## Features
  - Positional & Flag arguments.
  - Automatic validation.
    - Missing arguments will be caught and the program exits upon detection.
    - Supplied but not defined arguments will also cause the program to exit.
    - Basic type checking to respect every arg's default value's type.
  - Suitable for different use cases.
    - Normal stdin arg parsing.
    - (TODO) Continuous parsing of different types of / possibly faulty inputs (Maybe you're making some cli game?)

## The Chad's Basics
The Chad follows a simple yet effective parsing scheme.

Have a look:
  - `--foo "bar baz"` will assign the string `bar baz` to `foo`.
  - `-foo "bar baz"` will not work, since the arg-stack `foo` contains two equal flags which would result in a double-definition error.
  - `-bar 'foo baz'` will assign the string `foo baz` to `r`. `b` and `a` will be empty since they aren't the last flag in the arg-stack.
    - The first value, after an arg-stack given its not an arg or arg-stack itself will be the last flag's value of the current arg-stack.
  - `--HEY_how-are-you? 'good'` will assign the string `good` to `HEY_how-are-you?`.
    - The Chad allows for a lot of creativity :D
    - All arg names are allowed, as long as they don't start with a number. This is needed to be able to parse negative numbers like `-3`.
  - Different string delimiters are also allowed.
    - `"`, `'`, `<The backtick used to write these inline code "blocks">`.
    - Keep in mind that some may not work due to the way your terminal parses strings.

## How to properly work with the Chad
```go
func main() {
    c := chad.NewChad()
    c.RegisterArgs(
        []chad.Arg{
            *chad.NewArg("value", "Help for value", 69.420, false),
//                        ^^^^^    ^^^^^^^^^^^^^^   ^^^^^^  ^^^^^
//                        |        |                |       |- Is it required?
//                        |        |                |- The default value.
//                        |        |                   This can be of type string, bool, any int.. / uint.. or float..
//                        |        |- Help string
//                        |- The arg name; Used like this: go run main.go --value 420.69
        },
        []string{ /* ... */ },
//                ^^^^^^^^^
//                |- The positional arguments. Their names will be reflected in the program's help as well as your code.
//                   Positional arguments are possible to access via name or index. Use whichever suits you the best.
    )
    c.Parse()

    theValue := c.GetFloatFlag("value")
//                              ^^^^^
//                              |- The same name as above. If incorrect, the program will exit.
    fmt.Println(theValue)
}
```

## An example since everyone likes them
Feel free to try it out.
```go
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Fr4cK5/chad/chad"
)

func main() {
    c := chad.NewChad()
    c.RegisterArgs(
        []chad.Arg{
            *chad.NewArg("i", "Invert the image's colors", false, false),
            *chad.NewArg("c", "Compress the image", 1, false),
            *chad.NewArg("b", "Blur the image", 0, false),
            *chad.NewArg("p", "Posterize the image", 255, false), // 255 is the default range for colors in our image.
        },
        []string{"filename"},
    )
    c.Parse()

    // Access the positional argument by name.
    filename := c.StringPosName("filename")

    // We could also access it by index:
    // filename := c.StringIndex(0)

    // "Validate" the file format
	if !strings.HasSuffix(filename, ".png") && !strings.HasSuffix(filename, ".jpeg") {
		fmt.Printf("File '%v' is of unknown format.\n", filename)
		os.Exit(1)
	}

    fmt.Printf("Loading file '%v'...\n", filename)

    // If the flag is not defaulted / changed by the user's input, we can use it to do something with the image.
	if !c.IsFlagDefault("c") {
		fmt.Printf("Compressing image by a factor of %v...\n", c.IntFlag("c"))
	}

	if !c.IsFlagDefault("b") {
		fmt.Printf("Reducing color range from 255 to %v...\n", c.IntFlag("p"))
	}

	if !c.IsFlagDefault("b") {
		fmt.Printf("Blurring image by %v pixels using gaussian blur...\n", c.IntFlag("b"))
	}

	if c.BoolFlag("i") {
		fmt.Printf("Inverting image colors...\n")
	}

	extSep := strings.Index(filename, ".")
	fmt.Printf("Saving image to '%v-01.%v'...\n", filename[:extSep], filename[extSep+1:])
	fmt.Println("Conversion finished.")
}
```
Now try running these commands (some might fail):
  - `go run main.go my-image.png -p 100 -ib 3`
  - `go run main.go my-image.png -p 100 -i -b 3`
  - `go run main.go -ip 32`
  - `go run main.go the-other-image.jpeg -c 8 -b 4`
  - `go run main.go --help`

## This may interest you
  - Currently, this is on my list to make Chad even more based
    - Allow the program to persist if an arg's validation fails instead of just exiting.
  - What Chad will likely* never have:
    - Subcommands
