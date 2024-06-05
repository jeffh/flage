package flage

import (
	"bytes"
	"flag"
	"fmt"
	"strconv"
	"strings"
)

type resetable interface{ Reset() }

// Reset zeros out the flag.Value given.
//
// If the flag.Value has a Reset() method, that is called instead.
// Otherwise, defaults to calling value.Set("").
//
// Implementers of Reset() should take case to not mutate the original value, in case it's
// used in other parts of the code base (post flag parsing).
//
// Example:
//
//	var args flage.StringSlice
//	flag.Var(&args, "arg", "additional arguments to pass. Can be used multiple times")
//	// ...
//	flag.Parse()
//	fmt.Printf("args are: %s", strings.Join(args, ", "))
//	flage.Reset(&args)
//	fmt.Printf("args are: %s", strings.Join(args, ", ")) // will be empty
func Reset(f flag.Value) {
	if r, ok := f.(resetable); ok && r != nil {
		r.Reset()
	} else {
		f.Set("")
	}
}

// Int64Slice is a slice where mutliple of the flag appends to the slice
// Use ResetValues() to clear the slice (for multi-stage flag parsing)
type Int64Slice []int64

// String returns a string with ", " joined between each element
func (i *Int64Slice) String() string {
	var b bytes.Buffer
	for j, f := range *i {
		if j != 0 {
			b.Write([]byte(", "))
		}
		fmt.Fprintf(&b, "%d", f)
	}
	return b.String()
}

// Set appends an int64 or returns error if it is an invalid int. Use Reset() to reset the string slice to an empty slice.
func (i *Int64Slice) Set(value string) error {
	if value != "" {
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		*i = append(*i, v)
	}
	return nil
}

// Reset creates a new slice to use
func (i *Int64Slice) Reset() { *i = make(Int64Slice, 0) }

// Uint64Slice is a slice where mutliple of the flag appends to the slice
// Use ResetValues() to clear the slice (for multi-stage flag parsing)
type Uint64Slice []uint64

// String returns a string with ", " joined between each element
func (i *Uint64Slice) String() string {
	var b bytes.Buffer
	for j, f := range *i {
		if j != 0 {
			b.Write([]byte(", "))
		}
		fmt.Fprintf(&b, "%d", f)
	}
	return b.String()
}

// Set appends an int64 or returns error if it is an invalid uint. Use Reset() to reset the string slice to an empty slice.
func (i *Uint64Slice) Set(value string) error {
	if value != "" {
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		*i = append(*i, v)
	}
	return nil
}

// Reset creates a new slice to use
func (i *Uint64Slice) Reset() { *i = make(Uint64Slice, 0) }

// FloatSlice is a slice where mutliple of the flag appends to the slice
// Use ResetValues() to clear the slice (for multi-stage flag parsing)
type FloatSlice []float64

// String returns a string with ", " joined between each element
func (i *FloatSlice) String() string {
	var b bytes.Buffer
	for j, f := range *i {
		if j != 0 {
			b.Write([]byte(", "))
		}
		fmt.Fprintf(&b, "%f", f)
	}
	return b.String()
}

// Set appends an int64 or returns error if it is an invalid float64. Use Reset() to reset the string slice to an empty slice.
func (i *FloatSlice) Set(value string) error {
	if value != "" {
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		*i = append(*i, v)
	}
	return nil
}
func (i *FloatSlice) Reset() { *i = make(FloatSlice, 0) }

// StringSlice is a slice where mutliple of the flag appends to the slice
// Use ResetValues() to clear the slice (for multi-stage flag parsing)
type StringSlice []string

// String returns a string with ", " joined between each element
func (i *StringSlice) String() string { return strings.Join(*i, ", ") }

// Set appends to the string slice. Use Reset() to reset the string slice to an empty slice.
func (i *StringSlice) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// Reset creates a new slice to use
func (i *StringSlice) Reset() { *i = make(StringSlice, 0) }
