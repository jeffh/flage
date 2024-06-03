package flage

import (
	"encoding"
	"flag"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

type ExampleMarshal struct {
	T time.Time
	N big.Int `flage:",1"`
}

func (e *ExampleMarshal) MarshalFlagField(name string) encoding.TextMarshaler {
	switch name {
	case "T":
		return time.Now().UTC().Truncate(0)
	case "N":
		return nil
		// return big.NewInt(1)
	default:
		panic("unreachable")
	}
}

func TestStructVarTextMarshaling(t *testing.T) {
	var example ExampleMarshal
	fs := FlagSetStruct("test", flag.ContinueOnError, &example)
	err := fs.Parse([]string{
		"-t", "2024-03-22T10:33:50Z",
		"-n", "100",
	})
	if err != nil {
		t.Errorf("failed to parse flags: %s", err.Error())
	}

	tim, err := time.Parse(time.RFC3339, "2024-03-22T10:33:50Z")
	if err != nil {
		t.Errorf("failed to parse time: %s", err.Error())
	}

	expected := ExampleMarshal{T: tim.UTC().Truncate(0)}
	expected.N.SetInt64(100)
	if !reflect.DeepEqual(expected, example) {
		t.Errorf("expected %#v, got %#v", expected, example)
	}
}

func TestStructVarParsing(t *testing.T) {
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
	var example Example

	fs := FlagSetStruct("test", flag.ContinueOnError, &example)
	err := fs.Parse([]string{
		"-bool",
		"-str", "hello",
		"-u", "1",
		"-u64", "1024",
		"-i", "-1",
		"-i64", "-1024",
		"-f64", "-3.5",
		"-d", "10s",
	})
	if err != nil {
		t.Errorf("failed to parse flags: %s", err.Error())
	}

	expected := Example{
		Bool: true,
		Str:  "hello",
		U:    1,
		U64:  1024,
		I:    -1,
		I64:  -1024,
		F64:  -3.5,
		D:    10 * time.Second,
	}

	if !reflect.DeepEqual(expected, example) {
		t.Errorf("expected %#v, got %#v", expected, example)
	}
}

func TestStructVarParsingWithTags(t *testing.T) {
	type Example struct {
		Bool bool          `flage:"b,true"`
		Str  string        `flage:",world"`
		U    uint          `flage:"uint,2"`
		U64  uint64        `flage:"uint64,6"`
		I    int           `flage:"int,-2"`
		I64  int64         `flage:"int64,-6"`
		F64  float64       `flage:"float64,64"`
		D    time.Duration `flage:"dur,15s"`
	}
	var example Example

	fs := FlagSetStruct("test", flag.ContinueOnError, &example)
	err := fs.Parse([]string{
		"-str", "hello",
		"-uint64", "1024",
		"-int64", "-1024",
		"-float64", "-3.5",
	})
	if err != nil {
		t.Errorf("failed to parse flags: %s", err.Error())
	}

	expected := Example{
		Bool: true,
		Str:  "hello",
		U:    2,
		U64:  1024,
		I:    -2,
		I64:  -1024,
		F64:  -3.5,
		D:    15 * time.Second,
	}

	if !reflect.DeepEqual(expected, example) {
		t.Errorf("expected %#v, got %#v", expected, example)
	}
}

func TestStructVarParsingNestedStructs(t *testing.T) {
	type Example struct {
		Bool bool          `flage:"b,true"`
		Str  string        `flage:",world"`
		U    uint          `flage:"uint,2"`
		U64  uint64        `flage:"uint64,6"`
		I    int           `flage:"int,-2"`
		I64  int64         `flage:"int64,-6"`
		F64  float64       `flage:"float64,64"`
		D    time.Duration `flage:"dur,15s"`
	}
	type Nested struct {
		Example Example `flage:"*"`
	}
	var example Nested
	fs := FlagSetStruct("test", flag.ContinueOnError, &example)
	err := fs.Parse([]string{
		"-str", "hello",
		"-uint64", "1024",
		"-int64", "-1024",
		"-float64", "-3.5",
	})
	if err != nil {
		t.Errorf("failed to parse flags: %s", err.Error())
	}

	expected := Nested{
		Example: Example{
			Bool: true,
			Str:  "hello",
			U:    2,
			U64:  1024,
			I:    -2,
			I64:  -1024,
			F64:  -3.5,
			D:    15 * time.Second,
		},
	}

	if !reflect.DeepEqual(expected, example) {
		t.Errorf("expected %#v, got %#v", expected, example)
	}
}

func TestStructVarParsingWithDefaults(t *testing.T) {
	type Example struct {
		Bool bool          `flage:",true"`
		Str  string        `flage:",world"`
		U    uint          `flage:",2"`
		U64  uint64        `flage:",1024"`
		I    int           `flage:",-2"`
		I64  int64         `flage:",-1024"`
		F64  float64       `flage:",-3.5"`
		D    time.Duration `flage:",15s"`
	}
	var example Example

	fs := FlagSetStruct("test", flag.ContinueOnError, &example)
	err := fs.Parse([]string{})
	if err != nil {
		t.Errorf("failed to parse flags: %s", err.Error())
	}

	expected := Example{
		Bool: true,
		Str:  "world",
		U:    2,
		U64:  1024,
		I:    -2,
		I64:  -1024,
		F64:  -3.5,
		D:    15 * time.Second,
	}

	if !reflect.DeepEqual(expected, example) {
		t.Errorf("expected %#v, got %#v", expected, example)
	}
}

type TypeWithTextMarshals struct{ X int }

func (t TypeWithTextMarshals) MarshalText() (text []byte, err error) {
	return []byte(fmt.Sprintf("%d", t.X)), nil
}
func (t *TypeWithTextMarshals) UnmarshalText(text []byte) error {
	v, err := strconv.ParseInt(string(text), 10, 64)
	if err == nil {
		t.X = int(v)
	}
	return err
}

type TypeWithNoImplementations struct{ X int }

type TypeWithNoTextMarshal struct{ X int }

func (t *TypeWithNoTextUnmarshal) UnmarshalText(text []byte) error {
	v, err := strconv.ParseInt(string(text), 10, 64)
	if err == nil {
		t.X = int(v)
	}
	return err
}

type TypeWithNoTextUnmarshal struct{ X int }

func (t TypeWithNoTextUnmarshal) MarshalText() (text []byte, err error) {
	return []byte(fmt.Sprintf("%d", t.X)), nil
}

type ExampleReturningFieldTypeMarshal struct {
	X       TypeWithNoTextUnmarshal
	returns encoding.TextMarshaler
}

func (e *ExampleReturningFieldTypeMarshal) MarshalFieldFlag(name string) encoding.TextMarshaler {
	return e.returns
}

func TestStructVarWithTextMarshaler(t *testing.T) {
	t.Run("works with default values", func(t *testing.T) {
		// defer expectPanic(t, "")
		type Example struct {
			A TypeWithTextMarshals `flage:"A,1"`
		}
		var example Example
		fs := FlagSetStruct("test", flag.ContinueOnError, &example)
		err := fs.Parse([]string{})
		if err != nil {
			t.Errorf("failed to parse flags: %s", err.Error())
		}
	})
	t.Run("panics when custom type is missing methods", func(t *testing.T) {
		defer expectPanic(t, "Example.A has an unsupported type: ")
		type Example struct {
			A TypeWithNoImplementations
		}
		var example Example
		fs := FlagSetStruct("test", flag.ContinueOnError, &example)
		err := fs.Parse([]string{})
		if err != nil {
			t.Errorf("failed to parse flags: %s", err.Error())
		}
	})
	t.Run("panics when UnmarshalText is missing", func(t *testing.T) {
		defer expectPanic(t, "Example.A must have a default value set.")
		type Example struct {
			A TypeWithNoTextUnmarshal
		}
		var example Example
		fs := FlagSetStruct("test", flag.ContinueOnError, &example)
		err := fs.Parse([]string{})
		if err != nil {
			t.Errorf("failed to parse flags: %s", err.Error())
		}
	})
	t.Run("panics when MarshalText is missing", func(t *testing.T) {
		defer expectPanic(t, "")
		type Example struct {
			A TypeWithNoTextMarshal
		}
		var example Example
		fs := FlagSetStruct("test", flag.ContinueOnError, &example)
		err := fs.Parse([]string{})
		if err != nil {
			t.Errorf("failed to parse flags: %s", err.Error())
		}
	})
	t.Run("panics when FieldTextMarshaler returns nil", func(t *testing.T) {
		defer expectPanic(t, "ExampleReturningFieldTypeMarshal.X must have a default value set.")
		var example ExampleReturningFieldTypeMarshal
		fs := FlagSetStruct("test", flag.ContinueOnError, &example)
		err := fs.Parse([]string{})
		if err != nil {
			t.Errorf("failed to parse flags: %s", err.Error())
		}
	})
	t.Run("panics when TextMarshaler has no default value", func(t *testing.T) {
		defer expectPanic(t, "Example.A must have a default value set.")
		type Example struct {
			A TypeWithTextMarshals
		}
		var example Example
		fs := FlagSetStruct("test", flag.ContinueOnError, &example)
		err := fs.Parse([]string{})
		if err != nil {
			t.Errorf("failed to parse flags: %s", err.Error())
		}
	})
}

func expectPanic(t *testing.T, msg string) {
	t.Helper()
	err := recover()
	if err == nil {
		t.Error("expected panic")
	} else if msg != "" {
		actual := ""
		switch err := err.(type) {
		case error:
			actual = err.Error()
		case string:
			actual = err
		default:
			t.Errorf("unknown panic value: %#v", err)
		}
		if !strings.Contains(actual, msg) {
			t.Errorf("expected panic to contain text %q, but got %q", msg, actual)
		}
	}
}
