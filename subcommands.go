package flage

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const flageCmdTag = "flage-cmd"

type HelpInfo struct {
	Commands []FlagSetDefinition
	Flagsets []*flag.FlagSet

	Progname             string // optional, defaults to os.Args[0]
	About                string
	CommandPrefix        string
	SkipPrintingCommands bool // don't print commands
}

// MakeUsageWithSubcommands creates a flag.Usage function that prints subcommands and arguments for them.
func MakeUsageWithSubcommands(info HelpInfo) func() {
	return func() {
		if info.Progname == "" {
			info.Progname = os.Args[0]
		}
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "Usage: %s [GLOBAL_OPTIONS] (COMMAND [COMMAND_OPTIONS])+\n", info.Progname)
		if info.About != "" {
			fmt.Fprintf(out, "\n%s\n", info.About)
		}
		fmt.Fprintf(out, "\nGLOBAL_OPTIONS:\n")
		flag.PrintDefaults()
		if info.CommandPrefix != "" {
			fmt.Fprintf(out, "\n%s\n", info.CommandPrefix)
		}
		if !info.SkipPrintingCommands {
			fmt.Fprintf(out, "\nCOMMANDS: (type '%s COMMAND -help' for command specific help)\n", info.Progname)
			PrintCommands(out, info.Commands)
		}

		if flag.Parsed() {
			fmt.Fprintf(out, "\n")
			it := newFlagSetIterator(flag.Args(), info.Flagsets)
			for it.Next() {
				fs := it.FlagSet()
				for _, f := range info.Flagsets {
					if f == fs {
						fmt.Fprintf(out, "\n")
						f.Usage()
						break
					}
				}
			}
		} else {
			fmt.Fprintf(out, "FLAGS FOR ALL COMMANDS:\n")
			PrintFlagSets(out, info.Flagsets)
		}
	}
}

type FlagSetDefinition struct {
	Name   string
	Desc   string
	OutVar any
}

func NewFlagSetsAndDefsFromStruct(v any, handling flag.ErrorHandling) *FlagSetsAndDefs {
	rv := reflect.ValueOf(v)
	t := rv.Type()
	for t.Kind() == reflect.Ptr {
		rv = rv.Elem()
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("expected value to be a struct pointer, got: %s", t.Kind().String()))
	}
	cmds := make([]FlagSetDefinition, 0, t.NumField())
	for i, n := 0, t.NumField(); i < n; i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		name := strings.ToLower(f.Name)
		docstring := ""
		if raw := strings.TrimSpace(f.Tag.Get(flageCmdTag)); raw != "" {
			parts := strings.SplitN(raw, ",", 3)
			if len(parts) > 0 && parts[0] != "" {
				name = parts[0]
			}
			if len(parts) > 1 {
				docstring = strings.TrimSpace(parts[1])
			}
		}
		if name == "-" {
			continue
		}

		ptr := rv.Field(i).Addr().Interface()
		switch f.Type.Kind() {
		case reflect.Struct:
			cmds = append(cmds, FlagSetDefinition{name, docstring, ptr})
		default:
			panic(fmt.Errorf("%s: unsupported field type for 'flage.NewFlagSetsFromStruct' parsing: %s", f.Name, f.Type.Kind().String()))
		}
	}
	return NewFlagSets(cmds, handling)
}

type FlagSetsAndDefs struct {
	Defs []FlagSetDefinition
	Sets []*flag.FlagSet
}

func NewFlagSets(defs []FlagSetDefinition, handling flag.ErrorHandling) *FlagSetsAndDefs {
	sets := make([]*flag.FlagSet, len(defs))
	for i, def := range defs {
		sets[i] = FlagSetStruct(def.Name, handling, def.OutVar)
	}
	return &FlagSetsAndDefs{
		Defs: defs,
		Sets: sets,
	}
}
func (fss *FlagSetsAndDefs) Parse(args []string) *CommandIterator {
	return &CommandIterator{
		fss,
		newFlagSetIterator(args, fss.Sets),
	}
}

type CommandIterator struct {
	compiledFlags *FlagSetsAndDefs
	it            *flagSetIterator
}

func (it *CommandIterator) Next() bool   { return it.it.Next() }
func (it *CommandIterator) FlagPtr() any { return it.compiledFlags.OutVarFromFlagset(it.it.FlagSet()) }
func (it *CommandIterator) Err() error   { return it.it.Err() }

// DefinitionFromFlagset returns the FlagSetDefinition for a given flagset
func (fss *FlagSetsAndDefs) DefinitionFromFlagset(fs *flag.FlagSet) (*FlagSetDefinition, bool) {
	for i, s := range fss.Sets {
		if s == fs {
			return &fss.Defs[i], true
		}
	}
	return nil, false
}

func (fss *FlagSetsAndDefs) OutVarFromFlagset(fs *flag.FlagSet) any {
	if def, ok := fss.DefinitionFromFlagset(fs); ok {
		return def.OutVar
	}
	return nil
}

// PrintCommands prints flagset definitions
func PrintCommands(w io.Writer, defs []FlagSetDefinition) {
	maxSize := 0
	for _, cmd := range defs {
		s := len(cmd.Name)
		if maxSize < s {
			maxSize = s
		}
	}
	for _, cmd := range defs {
		fmt.Fprintf(w, "  %s%s\t%s\n", cmd.Name, strings.Repeat(" ", maxSize-len(cmd.Name)), cmd.Desc)
	}
}

// PrintFlagSets prints flagset usages with a newline separate in between
func PrintFlagSets(w io.Writer, fss []*flag.FlagSet) {
	for _, set := range fss {
		fmt.Fprintf(w, "\n")
		set.Usage()
	}
}

var (
	ErrNoMatchingFlagSet = errors.New("no matching commands")
	ErrUnknownCommand    = errors.New("unknown command")
)

/*
flagSetIterator provides a basic subcommand pattern to process subcommands and their flagsets

Example:

	// flagset configuration assumed (sysFlags, deployFlags)
	flag.Parse()

	itr := util.FlagSetIterator{
		Args: flag.Args(),
		Sets: []*flag.FlagSet{
			sysFlags,
			deployFlags,
		},
	}
	for itr.Next() {
		fs := itr.FlagSet()
		switch fs {
		case sysFlags:
			fmt.Printf("SYSTEM: %#v %v\n", sysf, fs.Args())
		case deployFlags:
			fmt.Printf("DEPLOY: %#v %v\n", deployf, fs.Args())
		}
	}
	if err := itr.Err(); err != nil {
		if !errors.Is(err, util.ErrNoMatchingFlagSet) && !errors.Is(err, flag.ErrHelp) {
			fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		}
		flag.Usage()
		return 1
	}

This above code would allow usages line:

	./myprogram sysflag -sysflag-arg deployflag -deployflag-arg
	./myprogram sysflag -sysflag-arg -- deployflag -deployflag-arg
*/
type flagSetIterator struct {
	Args []string        // ok to modify between Next() calls
	Sets []*flag.FlagSet // not ok to modify during iteration

	names     []string
	curr      *flag.FlagSet
	err       error
	parsedOne bool
}

func newFlagSetIterator(args []string, sets []*flag.FlagSet) *flagSetIterator {
	return &flagSetIterator{Args: args, Sets: sets}
}

func (it *flagSetIterator) Init(sets []*flag.FlagSet, args []string) {
	it.Sets = sets
	it.Args = args
	it.names = nil
	it.curr = nil
	it.err = nil
	it.parsedOne = false
}

func (it *flagSetIterator) findSet(name string) *flag.FlagSet {
	if it.names == nil {
		it.names = make([]string, len(it.Sets))
		for i, set := range it.Sets {
			it.names[i] = set.Name()
		}
	}
	for i, n := range it.names {
		if name == n {
			it.curr = it.Sets[i]
			return it.curr
		}
	}
	return nil
}

// Advance skips i argments; returning false if it reached the end of the arg slice.
// Note: Next() will still behave correctly if Advance goes beyond the arg slice.
func (it *flagSetIterator) Advance(i int) bool {
	size := len(it.Args)
	if size <= i {
		it.Args = it.Args[size:]
		return false
	} else {
		it.Args = it.Args[i:]
		return true
	}
}

// FlagSet returns the flagset that was matched from Next().
//
// Note: FlagSet still return all remaining args via Args() method. It is up to
// you to parse non-flag arguments and then modify the iterator's Arg slice or
// call Advance.
func (it *flagSetIterator) FlagSet() *flag.FlagSet { return it.curr }

// Next returns true if a flagset matches based on the next argument matching the flagset's name.
//
// Returns false that no flagset matched with an optional error via Err() which can return:
//   - ErrNoMatchingFlagSet is returned if the iterator has consumed all args and never matched a flagset
//   - Errors from FlagSet.Parse()
//   - ErrHelp if the FlagSet requests printing help
//
// Returns true if a flagset was parsed successfully.
func (it *flagSetIterator) Next() bool {
	it.err = nil
	if len(it.Args) > 0 {
		set := it.findSet(it.Args[0])
		if set == nil {
			it.err = fmt.Errorf("%w: %s", ErrUnknownCommand, it.Args[0])
			return false
		}
		set.VisitAll(func(f *flag.Flag) { Reset(f.Value) })
		it.err = set.Parse(it.Args[1:])
		if it.err != nil {
			return false
		}
		it.Args = it.Args[len(it.Args)-set.NArg():]
		it.parsedOne = true
		return true
	} else {
		if !it.parsedOne {
			it.err = ErrNoMatchingFlagSet
		}
		return false
	}
}

// Err returns the error when Next() was called
func (it *flagSetIterator) Err() error { return it.err }

// CommandString converts a struct into a series of command line args
func CommandString(v any) []string {
	if v == nil {
		return nil
	}
	rv := reflect.ValueOf(v)
	t := rv.Elem().Type()
	if t.Kind() != reflect.Struct {
		panic("expected value to be a struct pointer")
	}

	n := t.NumField()
	out := make([]string, 0, n*2)
	for i := 0; i < n; i++ {
		f := t.Field(i)
		name := strings.ToLower(f.Name)
		if raw := strings.TrimSpace(f.Tag.Get("arg")); raw != "" {
			parts := strings.SplitN(raw, ",", 2)
			if len(parts) > 0 && parts[0] != "" {
				name = parts[0]
			}
		}
		if name == "-" {
			continue
		}
		name = "-" + name
		rstruct := rv.Elem()
		switch f.Type.Kind() {
		case reflect.Bool:
			value := rstruct.Field(i).Bool()
			if value {
				out = append(out, name)
			}
		case reflect.String:
			value := rstruct.Field(i).String()
			if value != "" {
				out = append(out, name, value)
			}
		case reflect.Int, reflect.Int64:
			value := rstruct.Field(i).Int()
			if value != 0 {
				if d, ok := rstruct.Field(i).Interface().(time.Duration); ok {
					out = append(out, name, d.String())
				} else {
					out = append(out, name, strconv.FormatInt(value, 10))
				}
			}
		case reflect.Uint, reflect.Uint64:
			value := rstruct.Field(i).Uint()
			if value != 0 {
				out = append(out, name, strconv.FormatUint(value, 10))
			}
		case reflect.Slice:
			value := rstruct.Field(i)
			L := value.Len()
			for j := 0; j < L; j++ {
				val := value.Index(j)
				switch val.Type().Kind() {
				case reflect.Bool:
					value := val.Bool()
					if value {
						out = append(out, name)
					}
				case reflect.String:
					value := val.String()
					out = append(out, name, value)
				case reflect.Int, reflect.Int64:
					value := val.Int()
					out = append(out, name, strconv.FormatInt(value, 10))
				case reflect.Uint, reflect.Uint64:
					value := val.Uint()
					out = append(out, name, strconv.FormatUint(value, 10))
				default:
					panic(fmt.Errorf("%s: unsupported field type for 'flag' emitting: %s", f.Name, f.Type.Kind().String()))
				}
			}
		default:
			panic(fmt.Errorf("%s: unsupported field type for 'flag' emitting: %s", f.Name, f.Type.Kind().String()))
		}
	}
	return out
}
