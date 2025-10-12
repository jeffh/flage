# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`flage` is a Go library that extends Go's built-in `flag` package with additional functionality for struct-based flag parsing, slice types, config file support, subcommands, and environment variable handling. It's designed to provide "a little bit more, but not too much" than the standard library.

## Building and Testing

```bash
# Build the project
go build -v ./...

# Run all tests
go test -v ./...

# Run tests with race detector
go test -race -v ./...

# Run go vet
go vet ./...
```

## Core Architecture

The library is organized into several key functional areas:

### 1. Struct-based Flag Parsing (struct.go)

`StructVar()` is the core function that allows defining flags using struct tags. It uses reflection to traverse struct fields and register them with a flagset.

- Tag format: `flage:"<flagName>,<defaultValue>,<docString>"`
- Special tag values:
  - `"-"` to ignore a field
  - `"*"` to recursively parse nested struct as top-level flags
  - `"$type"` in docstrings is replaced with the type name
- Additional tag `flage-base` controls the number base for integer parsing
- Supports: primitives (bool, string, int/int64, uint/uint64, float32/float64), time.Duration, flag.Value interface, and encoding.TextUnmarshaler interface

### 2. Resettable Values (values.go)

All flag values implement a resettable pattern through the `Reset()` method. This is critical for multi-stage flag parsing (e.g., subcommands).

- `resettableValue[T]` is a generic wrapper that handles primitive types
- `resettableFlagVar` wraps custom flag.Value types
- Generic parsers like `parseInt[X]`, `parseUint[X]`, `parseFloat[X]` use Go constraints for type safety

### 3. Slice Types (slices.go)

Provides flag types that accumulate values across multiple invocations:
- `StringSlice`, `Int64Slice`, `Uint64Slice`, `FloatSlice`
- Each implements `flag.Value` interface with `Set()`, `String()`, and `Reset()` methods
- `Reset()` is essential for clearing values between subcommand parses

### 4. Config File Parsing (config.go)

Simple config file format that converts to command-line arguments:
- Comments: lines starting with `#` (ignoring leading whitespace)
- Newlines converted to spaces
- Uses `github.com/google/shlex` for shell-like parsing
- `ParseConfigFile(string)` for parsing file contents
- `ReadConfigFile(string)` for reading from disk
- Also provides `ParseEnvironFile()` and `ReadEnvironFile()` for KEY=VALUE format files

### 5. Subcommand Support (subcommands.go)

Complex but powerful system for chaining multiple commands:

- `FlagSetDefinition` defines a command with name, description, and output struct
- `NewFlagSetsFromStruct()` creates commands from a struct where each field is a command struct
- `flagSetIterator` is the core iteration mechanism that matches args to flagsets
- `CommandIterator` provides high-level iteration with access to parsed flag structs
- `MakeUsageWithSubcommands()` creates comprehensive help output
- `CommandString()` converts a struct back into command-line args (useful for testing)

Key pattern: Commands are matched by name, parsed, then remaining args are passed to next iteration.

### 6. Environment Variable Handling (env.go)

Hierarchical environment lookup system:

- `Env` struct with `Parent` chain allows layering (e.g., file env -> system env)
- `Lookuper` interface enables custom key-value sources
- `EnvMap` implements both `Lookuper` and `flag.Value`
- `capturingEnvMap` tracks env var usage (useful for generating example configs)
- Context-based tracking of required vs optional lookups and default values
- Methods: `Get()`, `GetOr()`, `GetOrError()` for different lookup patterns

## Important Implementation Details

1. **Tag Parsing**: The `flage` tag uses comma separation, but only splits on first 3 commas (allowing commas in docstrings)

2. **Reset Pattern**: When iterating subcommands, all flags in a flagset are reset via `VisitAll()` before parsing the next command instance (subcommands.go:309)

3. **Type Annotation**: The `$type` placeholder in docstrings is replaced with the actual type name (struct.go:20-22)

4. **Generic Constraints**: Uses `golang.org/x/exp/constraints` for `Integer`, `Unsigned`, and `Float` type constraints

5. **Number Base Parsing**: The `flage-base` tag allows specifying base (0 for auto-detect, 10 for decimal, 16 for hex, etc.)

6. **Error Handling**: Default value parsing panics on failure (during flag registration) while runtime parsing returns errors

## Development Notes

- Requires Go 1.22.0+
- Uses GitHub Actions for CI (test, race detector, vet)
- No external dependencies except `github.com/google/shlex` and `golang.org/x/exp`
- Test files follow `*_test.go` convention
