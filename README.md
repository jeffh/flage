flage
======

[![Go Reference](https://pkg.go.dev/badge/github.com/jeffh/flage.svg)](https://pkg.go.dev/github.com/jeffh/flage)

Extensions to go's built-in flag parsing. For when you want a little bit more, but not too much.

Install
-------

Run:

```bash
go get github.com/jeffh/flage
```

Structs
-------

This package can use a struct for easy parsing using go's flag package. Supported types are:

 - types supported by the `flag` package
 - any type that supports the `flag.Value` interface
 - any type that supports `encoding.TextMarshal` and `encoding.TextUnmarshal` interfaces

Example:

```go
type Example struct {
    Bool bool
    Str  string
    U    uint
    U64  uint64
    I    int
    I64  int64
    F64  float64
    D    time.Duration
}
var opt Example
StructVar(&opt, nil) // this nil can be an optional flagset, otherwise, assumes flag.CommandLine
flag.Parse()
// opt will be populated
```

The argument names are the field names, lower-cased. You can add flage tags to customize them:

```go
type Example struct {
    Bool bool `flage:"yes"`
}
```

The tag is comma separated with the following format:

```
{FlagName},{DefaultValue},{DocString}

FlagName = optional, use "-" to ignore it, leave blank to use lowercase field name behavior
DefaultValue = default value, parsed as if it was an argument flag. Causes panics on failure to parse
DocString = docstring for when -help is used. Commas are accepted.
```

Finally, you can use structs to create flagsets via `FlagSetStruct`.


Slices
------

This package provides types that allow them to be used multiple time to build a slice:

```go
var args flage.StringSlice
flag.Var(&args, "arg", "additional arguments to pass. Can be used multiple times")
// ...
flag.Parse()

fmt.Printf("args are: %s", strings.Join(args, ", "))

// slices can be "reset" to clear them
flage.Reset(&args)

fmt.Printf("args are: %s", strings.Join(args, ", "))

// usage: myprogram -arg 1 -arg 2
// output:
// args are: 1, 2
// args are:
```

The following slices are supported:

 - `StringSlice` for slices of strings
 - `FloatSlice` for slices of float64
 - `IntSlice` for slices of int64
 - `UintSlice` for slices of uint64

These slices also support calling `Reset` on them to clear those slices, which can be useful
if you're reusing them in flagsets.

Config Files
------------

This feature is WIP and subject to change.

Sometimes using a bunch of flags is laborious and it would be nice to save to a
file. flage provides some helpers to do this:

```go
type Example struct {
    Config string

    Bool bool
    Str  string
    U    uint
    U64  uint64
    I    int
    I64  int64
    F64  float64
    D    time.Duration
}
var opt Example
StructVar(&opt, nil)
flag.Parse()

if opt.Config != "" {
    err := flage.ParseConfigFile(opt.Config)
    if err != nil {
        // ...
    }
}
// opt will be populated
```

The above code will allow `-config <file>` to point to a file that looks like:

```txt
# this is a comment and is ignored, # must be at the start of the line (ignoring only whitesepace)
-bool
-str "str"
-u 1 -u64 2
```

This is the same as passing in arguments to the command line argument (except
for `-config`) with a couple of differences:

 - `#` are single lined comments
 - Newlines are converted to spaces
 - Some template variables and functions are available a la go's text/template syntax


This is templated the following template context is available:

 - `{{.configDir}}` points to the directory that holds the config file specified via `-config <file>`
 - `{{env "MY_ENV_VAR"}}` returns the value of reading the environment variable `MY_ENV_VAR`
 - `{{envOr "MY_ENV_VAR" "DEFAULT"}}` returns the value of reading the environment variable `MY_ENV_VAR` or returns `"DEFAULT"` if not present
 - `{{envOrError "MY_ENV_VAR" "my error message"}}` returns the value of reading the environment variable `MY_ENV_VAR` or returns an error with `"my error message"` included

More may be added. You can define your own set by using
`TemplateConfigRenderer`, which the config functions wrap.
