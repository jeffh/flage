package flage

import (
	"encoding"
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// FieldTextMarshaler is the interface a struct needs to implement if it has fields
// of type encoding.TextUnmarshaler to return the appropriate default value.
type FieldTextMarshaler interface {
	MarshalFlagField(field string) encoding.TextMarshaler
}

// FlagSetStruct makes a new flagset based on an output string to set to
func FlagSetStruct(name string, errHandling flag.ErrorHandling, out any) *flag.FlagSet {
	fs := flag.NewFlagSet(name, errHandling)
	StructVar(out, fs)
	return fs
}

// StructVar performs like flag.Var(...) but using a struct. Can optionally be annotated using tags.
// If fs is nil, then the global functions in the flag package are used instead.
//
// Tags use the "flag" key with the following values: "<flagName>,<defaultValue>,<description>"
// If <flagName> is empty, then the lowercase of the fieldname is used.
// If <defaultValue> is empty, then the zero value is used.
// If <description> is empty, then the empty string is used.
//
// As per flag package, the following types are supported:
//
//   - string
//   - float64
//   - uint / uint64
//   - int / int64
//   - bool
//   - flag.Value
//   - encoding.TextUnmarshaler | encoding.TextMarshaler
//
// If encoding.TextUnmarshler is used then a method on the struct must be used
// to provide defaults to encoding.TextUnmarshaler:
//
//	func (m *myOptions) MarshalFlagField(field string) encoding.TextMarshaler {
//	  if field == "myTextUnmarshalerField" {
//	    return big.NewInt(123)
//	  }
//	  return nil // panics will happen if we return nil
//	}
//
// Example:
//
//	type Flag struct {
//	  Install bool `flag:"install,,enables installation"`
//	  ConfigFile string `flag:"config,,optional config file to load"`
//	}
//
//	var f Flag
//	StructVar(&f, nil)
//	flag.Parse()
func StructVar(v any, fs *flag.FlagSet) {
	if fs == nil {
		fs = flag.CommandLine
	}

	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	t := rv.Type()
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("expected value to be a struct pointer, got: %s", t.Kind().String()))
	}
	for i, n := 0, t.NumField(); i < n; i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		name := strings.ToLower(f.Name)
		defaultValue := ""
		docstring := ""
		var isSplat bool
		if raw := strings.TrimSpace(f.Tag.Get("flage")); raw != "" {
			parts := strings.SplitN(raw, ",", 3)
			if len(parts) > 0 && parts[0] != "" {
				if parts[0] == "*" {
					isSplat = true
				} else {
					name = parts[0]
				}
			}
			if len(parts) > 1 {
				val := strings.TrimSpace(parts[1])
				// defaultValue = parts[1]
				if val != "" {
					switch f.Type.Kind() {
					case reflect.String:
						defaultValue = parts[1]
					default:
						defaultValue = val
					}
				}
			}
			if len(parts) > 2 {
				docstring = parts[2]
			}
		}
		numBase := 10
		if raw := strings.TrimSpace(f.Tag.Get("flage-base")); raw != "" {
			v, err := strconv.ParseInt(raw, 10, 64)
			if err != nil {
				panic(fmt.Errorf("%s flage-base tag is not an integer: %w", name, err))
			}
			numBase = int(v)
		}
		if name == "-" {
			continue
		}

		ptr := rv.Field(i).Addr().Interface()
		if pt, ok := ptr.(flag.Value); ok {
			fs.Var(pt, name, docstring)
		} else if pt, ok := ptr.(encoding.TextUnmarshaler); ok {
			if provider, ok := v.(FieldTextMarshaler); ok {
				defaultMarshalValue := provider.MarshalFlagField(f.Name)
				if defaultMarshalValue != nil {
					fs.TextVar(pt, name, defaultMarshalValue, docstring)
				} else {
					if defaultValue != "" {
						if um, ok := ptr.(encoding.TextMarshaler); ok {
							err := pt.UnmarshalText([]byte(defaultValue))
							if err != nil {
								panic(fmt.Errorf("error parsing default value on field %s.%s: %w", t.Name(), f.Name, err))
							}
							fs.TextVar(pt, name, um, docstring)
						} else {
							panic(fmt.Sprintf("%s.%s must have a type that implements encoding.TextMarshaler (in addition to encoding.TextUnmarshaler)", t.Name(), f.Name))
						}
					} else {
						panic(fmt.Sprintf(
							"%s.%s needs an explicitly defined default value. %s.MarshalFlagField method must return a non-nil encoding.TextMarshaler value",
							t.Name(),
							f.Name,
							t.Name(),
						))
					}
				}
			} else {
				if defaultValue != "" {
					if um, ok := ptr.(encoding.TextMarshaler); ok {
						err := pt.UnmarshalText([]byte(defaultValue))
						if err != nil {
							panic(fmt.Errorf("error parsing default value on field %q: %w", f.Name, err))
						}
						fs.TextVar(pt, name, um, docstring)
					} else {
						panic(fmt.Sprintf("%s must have a type that implements encoding.TextMarshaler", f.Name))
					}
				} else {
					panic(fmt.Sprintf(
						"%s.%s must have a default value set. Implement %s.MarshalFlagField method or implement encoding.TextMarshaler with a default value in flage tag",
						t.Name(),
						f.Name,
						t.Name(),
					))
				}
			}
		} else {
			switch f.Type.Kind() {
			case reflect.Bool:
				var def bool
				var err error
				if defaultValue == "" {
					def, err = false, nil
				} else {
					def, err = strconv.ParseBool(defaultValue)
				}
				if err != nil {
					panic(err)
				}
				fs.BoolVar(ptr.(*bool), name, def, docstring)
			case reflect.String:
				fs.StringVar(ptr.(*string), name, defaultValue, docstring)
			case reflect.Int:
				if defaultValue == "" {
					defaultValue = "0"
				}
				v, err := strconv.ParseInt(defaultValue, numBase, f.Type.Bits())
				if err != nil {
					panic(err)
				}
				fs.IntVar(ptr.(*int), name, int(v), docstring)
			case reflect.Int64:
				if _, ok := ptr.(*time.Duration); ok {
					var v time.Duration
					if defaultValue != "" {
						var err error
						v, err = time.ParseDuration(defaultValue)
						if err != nil {
							panic(fmt.Errorf("failed to parse default value for %s: %w", name, err))
						}
					}
					fs.DurationVar(ptr.(*time.Duration), name, v, docstring)
				} else {
					var v int64
					if defaultValue != "" {
						var err error
						v, err = strconv.ParseInt(defaultValue, numBase, f.Type.Bits())
						if err != nil {
							panic(fmt.Errorf("failed to parse %s as integer (%q): %w", name, v, err))
						}
					}
					fs.Int64Var(ptr.(*int64), name, v, docstring)
				}
			case reflect.Uint:
				var v uint64
				if defaultValue != "" {
					var err error
					v, err = strconv.ParseUint(defaultValue, numBase, f.Type.Bits())
					if err != nil {
						panic(fmt.Errorf("failed to parse default value for %s: %w", name, err))
					}
				}
				fs.UintVar(ptr.(*uint), name, uint(v), docstring)
			case reflect.Uint64:
				var v uint64
				if defaultValue != "" {
					var err error
					v, err = strconv.ParseUint(defaultValue, numBase, f.Type.Bits())
					if err != nil {
						panic(fmt.Errorf("failed to parse default value for %s: %w", name, err))
					}
				}
				fs.Uint64Var(ptr.(*uint64), name, v, docstring)
			case reflect.Float64:
				var v float64
				if defaultValue != "" {
					var err error
					v, err = strconv.ParseFloat(defaultValue, f.Type.Bits())
					if err != nil {
						panic(fmt.Errorf("failed to parse default value for %s: %w", name, err))
					}
				}
				fs.Float64Var(ptr.(*float64), name, v, docstring)
			case reflect.Struct:
				if isSplat {
					StructVar(ptr, fs)
				} else {
					panic(fmt.Errorf("%s.%s has an unsupported type: %s", t.Name(), f.Name, f.Type.String()))
				}
			default:
				panic(fmt.Errorf("%s.%s has an unsupported type: %s", t.Name(), f.Name, f.Type.String()))
			}
		}
	}
}
