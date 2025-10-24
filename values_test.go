package flage

import (
	"flag"
	"testing"
)

func TestFloat32Var(t *testing.T) {
	var f float32
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	Float32Var(fs, &f, "float", 1.5, "A float32 value")

	err := fs.Parse([]string{"-float", "3.14"})
	if err != nil {
		t.Errorf("Failed to parse: %v", err)
	}
	if f != 3.14 {
		t.Errorf("Expected 3.14, got %f", f)
	}

	// Test reset
	fs.VisitAll(func(fl *flag.Flag) {
		Reset(fl.Value)
	})
	if f != 1.5 {
		t.Errorf("Expected 1.5 after reset, got %f", f)
	}
}

func TestVar(t *testing.T) {
	t.Run("with resettable value", func(t *testing.T) {
		var s StringSlice
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		Var(fs, &s, "slice", "", "A string slice")

		err := fs.Parse([]string{"-slice", "a", "-slice", "b"})
		if err != nil {
			t.Errorf("Failed to parse: %v", err)
		}
		if len(s) != 2 {
			t.Errorf("Expected 2 items, got %d", len(s))
		}

		// Test reset
		fs.VisitAll(func(fl *flag.Flag) {
			Reset(fl.Value)
		})
		if len(s) != 0 {
			t.Errorf("Expected 0 items after reset, got %d", len(s))
		}
	})

	t.Run("with non-resettable value", func(t *testing.T) {
		// Create a custom flag.Value that doesn't implement resetable
		type customValue struct {
			value string
		}
		var cv customValue
		impl := &struct {
			flag.Value
			val *customValue
		}{
			Value: &customStringValue{&cv.value},
			val:   &cv,
		}

		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		Var(fs, impl, "custom", "default", "A custom value")

		// Verify the default value was set
		if cv.value != "default" {
			t.Errorf("Expected 'default', got '%s'", cv.value)
		}
	})
}

// Helper type for testing Var with non-resettable value
type customStringValue struct {
	ptr *string
}

func (c *customStringValue) String() string {
	if c.ptr == nil {
		return ""
	}
	return *c.ptr
}

func (c *customStringValue) Set(s string) error {
	*c.ptr = s
	return nil
}

func TestResettableValueWithNilStringer(t *testing.T) {
	var s string
	rv := &resettableValue[string]{
		ptr:      &s,
		defvalue: "",
		parser:   func(s string) (string, error) { return s, nil },
		stringer: nil, // nil stringer
		isBool:   false,
	}

	if rv.String() != "" {
		t.Error("Expected empty string when stringer is nil")
	}
}

func TestResettableValueEdgeCases(t *testing.T) {
	// Test String() with nil pointer
	var s string
	rv := &resettableValue[string]{
		ptr:      &s,
		defvalue: "default",
		parser:   func(s string) (string, error) { return s, nil },
		stringer: func(s string) string { return s },
		isBool:   false,
	}

	if rv.String() != "" {
		t.Errorf("Expected empty string, got %s", rv.String())
	}

	// Set a value
	err := rv.Set("test")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if rv.String() != "test" {
		t.Errorf("Expected 'test', got %s", rv.String())
	}

	// Reset
	rv.Reset()
	if rv.String() != "default" {
		t.Errorf("Expected 'default' after reset, got %s", rv.String())
	}
}

func TestResettableFlagVarEdgeCases(t *testing.T) {
	// Test with nil Value
	rv := &resettableFlagVar{
		Value:  nil,
		defval: "default",
	}

	if rv.String() != "default" {
		t.Errorf("Expected 'default', got %s", rv.String())
	}
}
