# Chad
An absolutely **BASED GIGACHAD** of a Go command line parser

## The Chad's Basics
The Chad follows a simple yet effective parsing scheme.

Have a look:
  - `--foo "bar baz"` will assign the string `bar baz` to `foo`
  - `-foo "bar baz"` will not work, since the arg-stack `foo` contains two equal flags which would result in a double-definition error.
  - `-bar 'foo baz'` will assign the string `foo baz` to `r`. `b` and `a` will be empty since they aren't the last flag in the arg-stack.
    - The first value, after an arg-stack given its not an arg or arg-stack itself will be the last flag's value of the current arg-stack.
  - `--HEY_how-are-you? 'good'` will assign the string `good` to `HEY_how-are-you?`. The Chad allows for a lot of creativity :D
  - Different string delimiters are also allowed: `"`, `'`, `<The backtick used to write these inline code "blocks">`

## How to properly work with the Chad
```go
func main() {
    c := chad.NewChad()
    c.RegisterArgs(
        []chad.Arg{
            *chad.NewArg("value", "Help for value", 69.420, false),
//                        ^^^^^    ^^^^^^^^^^^^^^   ^^^^^^  ^^^^^
//                        |        |                |       |- Is it required?
//                        |        |                |- The default value. This can be of any type.
//                        |        |- Help string
//                        |- The arg name; Used like this: go run main.go --value 420.69
        },
        0,
    )
    c.Parse()

    theValue := c.GetStringFlag("value")
//                               ^^^^^
//                               |- The same name as above. If incorrect, the program will exit.
    fmt.Println(theValue)
}
```

## A small example
If you'd like, you can try out this small snippet below.
```go
package main

import (
    "fmt"

    "github.com/Fr4cK5/chad/chad"
)

func main() {
    c := chad.NewChad()
    c.RegisterArgs(
        []chad.Arg{
            *chad.NewArg("first-name", "Your first name", "", false),
            *chad.NewArg("last-name", "Your last name", "", true),
        },
        0,
    )
    c.Parse()

    last_name := c.GetStringFlag("last-name")
    if !c.IsFlagDefault("first-name") {
        first_name := c.GetStringFlag("first-name")
        fmt.Printf("Hello %v %v, enjoy your stay.", first_name, last_name)
    } else  {
        fmt.Printf("Hello Mr. / Mrs. %v, enjoy your stay.", last_name)
    }
}
```
Now, try running these two commands:
  - `go run main.go --last-name Smith`
  - `go run main.go --last-name Smith --first-name John`.
  - `go run main.go --last-name Smith --first-name John --age 69`.

## This may interest you
  - Validation is performed automatically
    - Missing arguments will be caught and the program exits upon detection.
    - Supplied but not defined arguments will also cause the program to exit.
  - Currently, this is on my list to make Chad even more based
    - Automatic generation of a help command.
    - Better error handling scheme when trying to access invalid arguments / arguments are found to be missing.
  - What Chad will likely* never have:
    - Subcommands.
