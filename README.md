flage
======

Extensions to go's built-in flag parsing. For when you want a little bit more, but not too much.

Install
-------

Run:

```bash
go get github.com/jeffh/flage
```

Structs
-------

This package can use a struct for easy parsing using go's flag package. Supported types supported
by the `flag` package or any type that supports the `flag.Value` interface.

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
StructVar(&opt, nil)
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
