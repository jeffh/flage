package flage

import (
	"encoding"
	"flag"
	"math/big"
	"reflect"
	"testing"
	"time"
)

type ExampleMarshal struct {
	T time.Time
	N big.Int
}

func (e *ExampleMarshal) MarshalFlagField(name string) encoding.TextMarshaler {
	switch name {
	case "T":
		return time.Now().UTC().Truncate(0)
	case "N":
		return big.NewInt(1)
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
