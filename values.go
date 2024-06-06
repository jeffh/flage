package flage

import (
	"encoding"
	"errors"
	"flag"
	"fmt"
	"strconv"
	"time"

	"golang.org/x/exp/constraints"
)

var errParse = errors.New("parse error")

type resettableValue[T any] struct {
	ptr      *T
	defvalue T
	parser   func(string) (T, error)
	stringer func(T) string
	isBool   bool
}

func (b *resettableValue[T]) IsBoolFlag() bool { return b.isBool }
func (b *resettableValue[T]) Set(s string) error {
	v, err := b.parser(s)
	if err != nil {
		err = errParse
	}
	*b.ptr = v
	return err
}
func (b *resettableValue[T]) Get() any { return T(*b.ptr) }
func (b *resettableValue[T]) String() string {
	if b == nil {
		return ""
	}
	return b.stringer(*b.ptr)
}
func (b *resettableValue[T]) Reset() { *b.ptr = b.defvalue }

func newVar[T any](ptr *T, defvalue T, parser func(string) (T, error), stringer func(T) string, isBool bool) *resettableValue[T] {
	*ptr = defvalue
	return &resettableValue[T]{ptr: ptr, defvalue: defvalue, parser: parser, stringer: stringer, isBool: isBool}
}

type resettableFlagVar struct {
	flag.Value
	defval string
}

func (b *resettableFlagVar) Reset() {
	if v, ok := b.Value.(resetable); ok {
		v.Reset()
	} else {
		err := b.Set(b.defval)
		if err != nil {
			panic(fmt.Errorf("failed to set flag value: %w", err))
		}
	}
}

func Var(fs *flag.FlagSet, p flag.Value, name string, value string, usage string) {
	if v, ok := p.(resetable); ok {
		v.Reset()
	} else {
		err := p.Set(value)
		if err != nil {
			panic(fmt.Errorf("failed to set flag value: %w", err))
		}
	}
	fs.Var(&resettableFlagVar{p, value}, name, usage)
}

func BoolVar(fs *flag.FlagSet, p *bool, name string, value bool, usage string) {
	fs.Var(newVar(p, value, strconv.ParseBool, strconv.FormatBool, true), name, usage)
}

func parseInt[X constraints.Integer](s string) (X, error) {
	v, err := strconv.ParseInt(s, 10, 64)
	return X(v), err
}
func formatInt[X constraints.Integer](v X) string { return strconv.FormatInt(int64(v), 10) }

func parseUint[X constraints.Unsigned](s string) (X, error) {
	v, err := strconv.ParseUint(s, 10, 64)
	return X(v), err
}
func formatUint[X constraints.Integer](v X) string { return strconv.FormatUint(uint64(v), 10) }

func IntVar(fs *flag.FlagSet, p *int, name string, value int, usage string) {
	fs.Var(newVar(p, value, parseInt, formatInt, false), name, usage)
}
func Int64Var(fs *flag.FlagSet, p *int64, name string, value int64, usage string) {
	fs.Var(newVar(p, value, parseInt, formatInt, false), name, usage)
}
func UintVar(fs *flag.FlagSet, p *uint, name string, value uint, usage string) {
	fs.Var(newVar(p, value, parseUint, formatUint, false), name, usage)
}
func Uint64Var(fs *flag.FlagSet, p *uint64, name string, value uint64, usage string) {
	fs.Var(newVar(p, value, parseUint, formatUint, false), name, usage)
}

func parseFloat[X constraints.Float](s string) (X, error) {
	v, err := strconv.ParseFloat(s, 64)
	return X(v), err
}
func formatFloat[X constraints.Float](v X) string {
	return strconv.FormatFloat(float64(v), 'g', -1, 64)
}

func Float32Var(fs *flag.FlagSet, p *float32, name string, value float32, usage string) {
	fs.Var(newVar(p, value, parseFloat, formatFloat, false), name, usage)
}
func Float64Var(fs *flag.FlagSet, p *float64, name string, value float64, usage string) {
	fs.Var(newVar(p, value, parseFloat, formatFloat, false), name, usage)
}

func stringParser(s string) (string, error) { return s, nil }
func formatString(s string) string          { return s }

func StringVar(fs *flag.FlagSet, p *string, name string, value string, usage string) {
	fs.Var(newVar(p, value, stringParser, formatString, false), name, usage)
}

func DurationVar(fs *flag.FlagSet, p *time.Duration, name string, value time.Duration, usage string) {
	fs.Var(newVar(p, value, time.ParseDuration, time.Duration.String, false), name, usage)
}

type textMarshalVar struct {
	ptr      encoding.TextUnmarshaler
	defvalue string
}

func textMarshal(m any, s string) string {
	if m, ok := m.(encoding.TextMarshaler); ok {
		txt, err := m.MarshalText()
		if err == nil {
			return string(txt)
		}
	}
	return ""
}

func (b *textMarshalVar) Set(s string) error { return b.ptr.UnmarshalText([]byte(s)) }
func (b *textMarshalVar) Get() any           { return b.ptr }
func (b *textMarshalVar) String() string {
	if b == nil {
		return ""
	}
	return textMarshal(b.ptr, b.defvalue)
}
func (b *textMarshalVar) Reset() {
	err := b.ptr.UnmarshalText([]byte(b.defvalue))
	if err != nil {
		panic(fmt.Errorf("failed to reset value: %w", err))
	}
}

func TextVar(fs *flag.FlagSet, p encoding.TextUnmarshaler, name string, value string, usage string) {
	if err := p.UnmarshalText([]byte(value)); err != nil {
		panic(fmt.Errorf("failed to set flag value %q: %w", name, err))
	}
	fs.Var(&textMarshalVar{p, value}, name, usage)
}
