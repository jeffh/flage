# AGENTS.md

Guidelines for AI coding agents working in this repository.

## Project Overview

`flage` is a Go library extending the standard `flag` package with struct-based parsing, slice types, config file support, subcommands, and environment variable handling.

## Build, Test, and Lint Commands

```bash
# Build the project
go build -v ./...

# Run all tests
go test -v ./...

# Run a single test by name
go test -v -run TestStructVarParsing ./...

# Run tests matching a pattern
go test -v -run "TestStructVar.*" ./...

# Run a specific subtest
go test -v -run "TestStructVarWithTextMarshaler/works_with_default_values" ./...

# Run tests with race detector
go test -race -v ./...

# Run static analysis
go vet ./...

# Format code (check before committing)
gofmt -s -w .
```

## Project Structure

```
flage/
├── struct.go        # Core struct-based flag parsing via reflection
├── values.go        # Resettable flag value wrappers with generics
├── slices.go        # Slice types (StringSlice, Int64Slice, etc.)
├── config.go        # Config file and .env file parsing
├── env.go           # Hierarchical environment variable handling
├── subcommands.go   # Subcommand support with iterator pattern
└── *_test.go        # Test files (same package)
```

## Code Style Guidelines

### Import Ordering

Standard Go import grouping: stdlib first, then external packages, alphabetically sorted within groups:

```go
import (
    "context"
    "errors"
    "flag"
    "fmt"

    "golang.org/x/exp/constraints"
)
```

### Naming Conventions

| Element | Style | Examples |
|---------|-------|----------|
| Exported types | PascalCase | `FlagSetDefinition`, `StringSlice`, `EnvMap` |
| Exported functions | PascalCase | `StructVar`, `Reset`, `ParseConfigFile` |
| Private functions | camelCase | `fileToCmdlineArgs`, `insertType`, `withContext` |
| Variables | camelCase | `errParse`, `sysEnv`, `defvalue` |
| Exported errors | PascalCase with `Err` prefix | `ErrNoMatchingFlagSet`, `ErrUnknownCommand` |
| Private errors | camelCase with `err` prefix | `errParse` |
| Receiver names | Single letter, type-appropriate | `b`, `i`, `e`, `fs`, `it` |

### Error Handling

1. **Sentinel errors** for known conditions:
   ```go
   var errParse = errors.New("parse error")
   var ErrNoMatchingFlagSet = errors.New("no matching commands")
   ```

2. **Error wrapping** with `fmt.Errorf` and `%w`:
   ```go
   return fmt.Errorf("failed to parse default value for %s: %w", name, err)
   return fmt.Errorf("%w: %s", ErrUnknownCommand, it.Args[0])
   ```

3. **Panic for programming errors** (invalid configuration during setup):
   ```go
   panic(fmt.Errorf("failed to parse default value for %s: %w", name, err))
   panic("expected value to be a struct pointer")
   ```

4. **Return errors for runtime failures** (user input errors during parsing).

### Nil Receiver Safety

Methods should handle nil receivers gracefully:

```go
func (b *resettableValue[T]) Set(s string) error {
    if b == nil {
        return nil
    }
    // ...
}

func (b *resettableValue[T]) String() string {
    if b == nil || b.stringer == nil {
        return ""
    }
    return b.stringer(*b.ptr)
}
```

### Interface Satisfaction Assertions

Use compile-time assertions to verify interface implementation:

```go
var _ flag.Value = (*EnvMap)(nil)
```

### Generic Type Constraints

Use `golang.org/x/exp/constraints` for numeric generics:

```go
func parseInt[X constraints.Integer](s string) (X, error) {
    v, err := strconv.ParseInt(s, 0, 64)
    return X(v), err
}
```

### Struct Tags

The `flage` tag format: `"<flagName>,<defaultValue>,<docString>"`
- Only splits on first 3 commas (allows commas in docstrings)
- Use `"-"` to ignore a field
- Use `"*"` for nested struct splat (top-level flags)
- Use `$type` in docstrings as placeholder for type name

```go
type Example struct {
    Config  string        `flage:"config,./config.json,path to $type config file"`
    Verbose bool          `flage:"v,false,enable verbose output"`
    Timeout time.Duration `flage:"timeout,30s,request timeout"`
    Inner   InnerFlags    `flage:"*"` // splat nested struct
    Ignored string        `flage:"-"` // ignored field
}
```

## Testing Patterns

### Table-Driven Tests with Subtests

```go
cases := []struct {
    Desc     string
    Input    []string
    Expected []int64
}{
    {"empty input", []string{}, nil},
    {"single value", []string{"1"}, []int64{1}},
}
for _, tc := range cases {
    t.Run(tc.Desc, func(t *testing.T) {
        // test logic
    })
}
```

### Helper Functions

Mark helpers with `t.Helper()`:

```go
func expectPanic(t *testing.T, msg string) {
    t.Helper()
    err := recover()
    if err == nil {
        t.Error("expected panic")
    }
    // ...
}
```

### Test Isolation

Each test should create its own `flag.FlagSet`:

```go
fs := flag.NewFlagSet("test", flag.ContinueOnError)
```

### Assertions

Use `reflect.DeepEqual` for complex comparisons:

```go
if !reflect.DeepEqual(expected, actual) {
    t.Errorf("expected %#v, got %#v", expected, actual)
}
```

## Key Implementation Details

1. **Reset Pattern**: All flag values must implement `Reset()` for multi-stage parsing (subcommands). Use `fs.VisitAll(func(f *flag.Flag) { Reset(f.Value) })` to reset all flags.

2. **Default Value Parsing**: Panics on invalid defaults during registration; returns errors during runtime parsing.

3. **Type Annotation**: `$type` in docstrings is replaced with the actual type name via `insertType()`.

4. **Number Base**: Use `flage-base` tag to control integer parsing base (0=auto, 10=decimal, 16=hex).

## Dependencies

- Go 1.22.0+
- `github.com/google/shlex` - shell-like parsing for config files
- `golang.org/x/exp/constraints` - generic type constraints
