package flage

import (
	"flag"
	"reflect"
	"testing"
)

func TestInt64Slice(t *testing.T) {
	cases := []struct {
		Desc     string
		Input    []string
		Expected []int64
	}{
		{"1 arg", []string{"-append", "1"}, []int64{1}},
		{"2 args", []string{"-append", "1", "-append", "2"}, []int64{1, 2}},
		{"3 args", []string{"-append", "1", "-append", "2", "-append", "3"}, []int64{1, 2, 3}},
	}

	for _, tc := range cases {
		t.Run(tc.Desc, func(t *testing.T) {
			var ss Int64Slice
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.Var(&ss, "append", "append a float")
			err := fs.Parse(tc.Input)
			if err != nil {
				t.Errorf("expected to parse cli args, got: %s", err.Error())
			}

			if !reflect.DeepEqual([]int64(ss), tc.Expected) {
				t.Errorf("expected to get %#v, got %#v", tc.Expected, ss)
			}

			Reset(&ss)
			if len(ss) != 0 {
				t.Error("expected Reset() to empty flag")
			}
		})
	}
}

func TestUint64Slice(t *testing.T) {
	cases := []struct {
		Desc     string
		Input    []string
		Expected []uint64
	}{
		{"1 arg", []string{"-append", "1"}, []uint64{1}},
		{"2 args", []string{"-append", "1", "-append", "2"}, []uint64{1, 2}},
		{"3 args", []string{"-append", "1", "-append", "2", "-append", "3"}, []uint64{1, 2, 3}},
	}

	for _, tc := range cases {
		t.Run(tc.Desc, func(t *testing.T) {
			var ss Uint64Slice
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.Var(&ss, "append", "append a float")
			err := fs.Parse(tc.Input)
			if err != nil {
				t.Errorf("expected to parse cli args, got: %s", err.Error())
			}

			if !reflect.DeepEqual([]uint64(ss), tc.Expected) {
				t.Errorf("expected to get %#v, got %#v", tc.Expected, ss)
			}
			Reset(&ss)
			if len(ss) != 0 {
				t.Error("expected Reset() to empty flag")
			}
		})
	}
}

func TestFloatSlice(t *testing.T) {
	cases := []struct {
		Desc     string
		Input    []string
		Expected []float64
	}{
		{"1 arg", []string{"-append", "1"}, []float64{1}},
		{"2 args", []string{"-append", "1", "-append", "2"}, []float64{1, 2}},
		{"3 args", []string{"-append", "1", "-append", "2", "-append", "3.5"}, []float64{1, 2, 3.5}},
	}

	for _, tc := range cases {
		t.Run(tc.Desc, func(t *testing.T) {
			var ss FloatSlice
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.Var(&ss, "append", "append a float")
			err := fs.Parse(tc.Input)
			if err != nil {
				t.Errorf("expected to parse cli args, got: %s", err.Error())
			}

			if !reflect.DeepEqual([]float64(ss), tc.Expected) {
				t.Errorf("expected to get %#v, got %#v", tc.Expected, ss)
			}
			Reset(&ss)
			if len(ss) != 0 {
				t.Error("expected Reset() to empty flag")
			}
		})
	}
}

func TestStringSlice(t *testing.T) {
	cases := []struct {
		Desc     string
		Input    []string
		Expected []string
	}{
		{"1 arg", []string{"-append", "1"}, []string{"1"}},
		{"2 args", []string{"-append", "1", "-append", "2"}, []string{"1", "2"}},
		{"3 args", []string{"-append", "1", "-append", "2", "-append", "hello"}, []string{"1", "2", "hello"}},
	}

	for _, tc := range cases {
		t.Run(tc.Desc, func(t *testing.T) {
			var ss StringSlice
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.Var(&ss, "append", "append a string")
			err := fs.Parse(tc.Input)
			if err != nil {
				t.Errorf("expected to parse cli args, got: %s", err.Error())
			}

			if !reflect.DeepEqual([]string(ss), tc.Expected) {
				t.Errorf("expected to get %#v, got %#v", tc.Expected, ss)
			}
			Reset(&ss)
			if len(ss) != 0 {
				t.Error("expected Reset() to empty flag")
			}
		})
	}
}
