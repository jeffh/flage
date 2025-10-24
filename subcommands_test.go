package flage

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestMakeUsageWithSubcommands(t *testing.T) {
	// Save original values
	origArgs := os.Args
	origCommandLine := flag.CommandLine
	defer func() {
		os.Args = origArgs
		flag.CommandLine = origCommandLine
	}()

	t.Run("basic usage", func(t *testing.T) {
		// Reset flag.CommandLine
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		var buf bytes.Buffer
		flag.CommandLine.SetOutput(&buf)

		defs := []FlagSetDefinition{
			{Name: "deploy", Desc: "Deploy application"},
			{Name: "test", Desc: "Run tests"},
		}

		fs1 := flag.NewFlagSet("deploy", flag.ContinueOnError)
		fs2 := flag.NewFlagSet("test", flag.ContinueOnError)

		info := HelpInfo{
			Commands: defs,
			Flagsets: []*flag.FlagSet{fs1, fs2},
			Progname: "myapp",
			About:    "My application",
		}

		usageFunc := MakeUsageWithSubcommands(info)
		usageFunc()

		output := buf.String()
		if !strings.Contains(output, "myapp") {
			t.Error("Expected output to contain program name")
		}
		if !strings.Contains(output, "My application") {
			t.Error("Expected output to contain about text")
		}
		if !strings.Contains(output, "deploy") {
			t.Error("Expected output to contain deploy command")
		}
	})

	t.Run("with command prefix", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		var buf bytes.Buffer
		flag.CommandLine.SetOutput(&buf)

		info := HelpInfo{
			Commands:      []FlagSetDefinition{},
			Flagsets:      []*flag.FlagSet{},
			Progname:      "myapp",
			CommandPrefix: "Available commands:",
		}

		usageFunc := MakeUsageWithSubcommands(info)
		usageFunc()

		output := buf.String()
		if !strings.Contains(output, "Available commands:") {
			t.Error("Expected output to contain command prefix")
		}
	})

	t.Run("skip printing commands", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		var buf bytes.Buffer
		flag.CommandLine.SetOutput(&buf)

		defs := []FlagSetDefinition{
			{Name: "deploy", Desc: "Deploy application"},
		}

		info := HelpInfo{
			Commands:             defs,
			Flagsets:             []*flag.FlagSet{},
			Progname:             "myapp",
			SkipPrintingCommands: true,
		}

		usageFunc := MakeUsageWithSubcommands(info)
		usageFunc()

		output := buf.String()
		// When SkipPrintingCommands is true, PrintCommands should not be called
		// So we shouldn't see the deploy command description in the output
		// The heading "COMMANDS:" won't appear either since !info.SkipPrintingCommands is false
		if strings.Contains(output, "Deploy application") {
			t.Error("Expected output to skip command descriptions")
		}
	})
}

func TestNewFlagSetsAndDefsFromStruct(t *testing.T) {
	type DeployCmd struct {
		Env string `flage:"env,development,Environment to deploy to"`
	}

	type TestCmd struct {
		Verbose bool `flage:"verbose,false,Enable verbose output"`
	}

	type Commands struct {
		Deploy DeployCmd `flage-cmd:"deploy,Deploy application"`
		Test   TestCmd   `flage-cmd:"test,Run tests"`
	}

	t.Run("basic struct parsing", func(t *testing.T) {
		cmds := &Commands{}
		fss := NewFlagSetsAndDefsFromStruct(cmds, flag.ContinueOnError)

		if len(fss.Defs) != 2 {
			t.Errorf("Expected 2 definitions, got %d", len(fss.Defs))
		}
		if len(fss.Sets) != 2 {
			t.Errorf("Expected 2 flagsets, got %d", len(fss.Sets))
		}

		// Check that definitions are correct
		if fss.Defs[0].Name != "deploy" {
			t.Errorf("Expected first command to be 'deploy', got '%s'", fss.Defs[0].Name)
		}
		if fss.Defs[1].Name != "test" {
			t.Errorf("Expected second command to be 'test', got '%s'", fss.Defs[1].Name)
		}
	})

	t.Run("ignore unexported fields", func(t *testing.T) {
		type CommandsWithPrivate struct {
			Deploy  DeployCmd `flage-cmd:"deploy,Deploy application"`
			private TestCmd
		}

		cmds := &CommandsWithPrivate{}
		fss := NewFlagSetsAndDefsFromStruct(cmds, flag.ContinueOnError)

		if len(fss.Defs) != 1 {
			t.Errorf("Expected 1 definition, got %d", len(fss.Defs))
		}
	})

	t.Run("skip fields marked with dash", func(t *testing.T) {
		type CommandsWithSkip struct {
			Deploy DeployCmd `flage-cmd:"deploy,Deploy application"`
			Skip   TestCmd   `flage-cmd:"-"`
		}

		cmds := &CommandsWithSkip{}
		fss := NewFlagSetsAndDefsFromStruct(cmds, flag.ContinueOnError)

		if len(fss.Defs) != 1 {
			t.Errorf("Expected 1 definition, got %d", len(fss.Defs))
		}
	})

	t.Run("panic on non-struct", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for non-struct type")
			}
		}()

		var notStruct string
		NewFlagSetsAndDefsFromStruct(&notStruct, flag.ContinueOnError)
	})

	t.Run("panic on unsupported field type", func(t *testing.T) {
		type InvalidCommands struct {
			InvalidField int `flage-cmd:"invalid,Invalid field"`
		}

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for unsupported field type")
			}
		}()

		cmds := &InvalidCommands{}
		NewFlagSetsAndDefsFromStruct(cmds, flag.ContinueOnError)
	})
}

func TestNewFlagSets(t *testing.T) {
	type DeployCmd struct {
		Env string `flage:"env,development,Environment"`
	}

	t.Run("creates flagsets from definitions", func(t *testing.T) {
		deploy := &DeployCmd{}
		defs := []FlagSetDefinition{
			{Name: "deploy", Desc: "Deploy app", OutVar: deploy},
		}

		fss := NewFlagSets(defs, flag.ContinueOnError)

		if len(fss.Defs) != 1 {
			t.Errorf("Expected 1 definition, got %d", len(fss.Defs))
		}
		if len(fss.Sets) != 1 {
			t.Errorf("Expected 1 flagset, got %d", len(fss.Sets))
		}
		if fss.Sets[0].Name() != "deploy" {
			t.Errorf("Expected flagset name 'deploy', got '%s'", fss.Sets[0].Name())
		}
	})
}

func TestCommandIterator(t *testing.T) {
	type DeployCmd struct {
		Env     string `flage:"env,development,Environment"`
		Verbose bool   `flage:"verbose,false,Verbose output"`
	}

	type TestCmd struct {
		Coverage bool `flage:"coverage,false,Run with coverage"`
	}

	type Commands struct {
		Deploy DeployCmd `flage-cmd:"deploy,Deploy application"`
		Test   TestCmd   `flage-cmd:"test,Run tests"`
	}

	t.Run("iterate through commands", func(t *testing.T) {
		cmds := &Commands{}
		fss := NewFlagSetsAndDefsFromStruct(cmds, flag.ContinueOnError)

		args := []string{"deploy", "-env", "production", "test", "-coverage"}
		it := fss.Parse(args)

		// First command: deploy
		if !it.Next() {
			t.Fatal("Expected first command to be parsed")
		}
		def := it.FlagDef()
		if def.Name != "deploy" {
			t.Errorf("Expected 'deploy', got '%s'", def.Name)
		}
		deployPtr := it.FlagPtr().(*DeployCmd)
		if deployPtr.Env != "production" {
			t.Errorf("Expected env='production', got '%s'", deployPtr.Env)
		}

		// Second command: test
		if !it.Next() {
			t.Fatal("Expected second command to be parsed")
		}
		def = it.FlagDef()
		if def.Name != "test" {
			t.Errorf("Expected 'test', got '%s'", def.Name)
		}
		testPtr := it.FlagPtr().(*TestCmd)
		if !testPtr.Coverage {
			t.Error("Expected coverage=true")
		}

		// No more commands
		if it.Next() {
			t.Error("Expected no more commands")
		}
		if it.Err() != nil {
			t.Errorf("Expected no error, got %v", it.Err())
		}
	})

	t.Run("error on unknown command", func(t *testing.T) {
		cmds := &Commands{}
		fss := NewFlagSetsAndDefsFromStruct(cmds, flag.ContinueOnError)

		args := []string{"unknown"}
		it := fss.Parse(args)

		if it.Next() {
			t.Error("Expected Next() to return false for unknown command")
		}
		if it.Err() == nil {
			t.Error("Expected error for unknown command")
		}
		if !strings.Contains(it.Err().Error(), "unknown") {
			t.Errorf("Expected error about unknown command, got: %v", it.Err())
		}
	})

	t.Run("no matching flagset when no args", func(t *testing.T) {
		cmds := &Commands{}
		fss := NewFlagSetsAndDefsFromStruct(cmds, flag.ContinueOnError)

		args := []string{}
		it := fss.Parse(args)

		if it.Next() {
			t.Error("Expected Next() to return false for empty args")
		}
		if it.Err() != ErrNoMatchingFlagSet {
			t.Errorf("Expected ErrNoMatchingFlagSet, got %v", it.Err())
		}
	})
}

func TestDefinitionFromFlagset(t *testing.T) {
	type DeployCmd struct {
		Env string `flage:"env,development,Environment"`
	}

	deploy := &DeployCmd{}
	defs := []FlagSetDefinition{
		{Name: "deploy", Desc: "Deploy app", OutVar: deploy},
	}

	fss := NewFlagSets(defs, flag.ContinueOnError)

	t.Run("find definition for flagset", func(t *testing.T) {
		def, ok := fss.DefinitionFromFlagset(fss.Sets[0])
		if !ok {
			t.Error("Expected to find definition")
		}
		if def.Name != "deploy" {
			t.Errorf("Expected 'deploy', got '%s'", def.Name)
		}
	})

	t.Run("return false for unknown flagset", func(t *testing.T) {
		unknownFS := flag.NewFlagSet("unknown", flag.ContinueOnError)
		_, ok := fss.DefinitionFromFlagset(unknownFS)
		if ok {
			t.Error("Expected to not find definition for unknown flagset")
		}
	})
}

func TestOutVarFromFlagset(t *testing.T) {
	type DeployCmd struct {
		Env string `flage:"env,development,Environment"`
	}

	deploy := &DeployCmd{}
	defs := []FlagSetDefinition{
		{Name: "deploy", Desc: "Deploy app", OutVar: deploy},
	}

	fss := NewFlagSets(defs, flag.ContinueOnError)

	t.Run("get outvar for flagset", func(t *testing.T) {
		outVar := fss.OutVarFromFlagset(fss.Sets[0])
		if outVar == nil {
			t.Error("Expected to get outvar")
		}
		if outVar != deploy {
			t.Error("Expected outvar to match original")
		}
	})

	t.Run("return nil for unknown flagset", func(t *testing.T) {
		unknownFS := flag.NewFlagSet("unknown", flag.ContinueOnError)
		outVar := fss.OutVarFromFlagset(unknownFS)
		if outVar != nil {
			t.Error("Expected nil for unknown flagset")
		}
	})
}

func TestPrintCommands(t *testing.T) {
	defs := []FlagSetDefinition{
		{Name: "deploy", Desc: "Deploy application"},
		{Name: "test", Desc: "Run tests"},
		{Name: "build", Desc: "Build project"},
	}

	var buf bytes.Buffer
	PrintCommands(&buf, defs)

	output := buf.String()
	if !strings.Contains(output, "deploy") {
		t.Error("Expected output to contain 'deploy'")
	}
	if !strings.Contains(output, "Deploy application") {
		t.Error("Expected output to contain deploy description")
	}
	if !strings.Contains(output, "test") {
		t.Error("Expected output to contain 'test'")
	}
}

func TestPrintFlagSets(t *testing.T) {
	fs1 := flag.NewFlagSet("deploy", flag.ContinueOnError)
	fs1.String("env", "dev", "Environment")

	fs2 := flag.NewFlagSet("test", flag.ContinueOnError)
	fs2.Bool("verbose", false, "Verbose output")

	var buf bytes.Buffer
	fs1.SetOutput(&buf)
	fs2.SetOutput(&buf)

	PrintFlagSets(&buf, []*flag.FlagSet{fs1, fs2})

	output := buf.String()
	// Should print usage for both flagsets with newlines between
	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		t.Error("Expected multiple lines of output")
	}
}

func TestFlagSetIterator(t *testing.T) {
	type DeployFlags struct {
		Env string `flage:"env,development,Environment"`
	}

	type TestFlags struct {
		Verbose bool `flage:"verbose,false,Verbose output"`
	}

	t.Run("basic iteration", func(t *testing.T) {
		deployFlags := &DeployFlags{}
		testFlags := &TestFlags{}

		deployFS := FlagSetStruct("deploy", flag.ContinueOnError, deployFlags)
		testFS := FlagSetStruct("test", flag.ContinueOnError, testFlags)

		it := newFlagSetIterator(
			[]string{"deploy", "-env", "prod", "test", "-verbose"},
			[]*flag.FlagSet{deployFS, testFS},
		)

		// First iteration - deploy
		if !it.Next() {
			t.Fatal("Expected Next() to return true")
		}
		if it.FlagSet() != deployFS {
			t.Error("Expected deploy flagset")
		}
		if deployFlags.Env != "prod" {
			t.Errorf("Expected env='prod', got '%s'", deployFlags.Env)
		}

		// Second iteration - test
		if !it.Next() {
			t.Fatal("Expected Next() to return true for second command")
		}
		if it.FlagSet() != testFS {
			t.Error("Expected test flagset")
		}
		if !testFlags.Verbose {
			t.Error("Expected verbose=true")
		}

		// No more
		if it.Next() {
			t.Error("Expected Next() to return false")
		}
	})

	t.Run("advance method", func(t *testing.T) {
		it := newFlagSetIterator(
			[]string{"arg1", "arg2", "arg3"},
			[]*flag.FlagSet{},
		)

		// Advance by 1
		if !it.Advance(1) {
			t.Error("Expected Advance(1) to return true")
		}
		if len(it.Args) != 2 {
			t.Errorf("Expected 2 args remaining, got %d", len(it.Args))
		}

		// Advance beyond end
		if it.Advance(10) {
			t.Error("Expected Advance(10) to return false")
		}
		if len(it.Args) != 0 {
			t.Errorf("Expected 0 args remaining, got %d", len(it.Args))
		}
	})

	t.Run("unknown command error", func(t *testing.T) {
		deployFS := flag.NewFlagSet("deploy", flag.ContinueOnError)

		it := newFlagSetIterator(
			[]string{"unknown"},
			[]*flag.FlagSet{deployFS},
		)

		if it.Next() {
			t.Error("Expected Next() to return false for unknown command")
		}
		if it.Err() == nil {
			t.Error("Expected error for unknown command")
		}
	})

	t.Run("Init method", func(t *testing.T) {
		it := &flagSetIterator{}
		deployFS := flag.NewFlagSet("deploy", flag.ContinueOnError)

		it.Init([]*flag.FlagSet{deployFS}, []string{"deploy"})

		if len(it.Sets) != 1 {
			t.Error("Expected Sets to be initialized")
		}
		if len(it.Args) != 1 {
			t.Error("Expected Args to be initialized")
		}
	})
}

func TestCommandString(t *testing.T) {
	t.Run("basic types", func(t *testing.T) {
		type Flags struct {
			Name    string `arg:"name"`
			Verbose bool   `arg:"verbose"`
			Count   int    `arg:"count"`
			Port    uint   `arg:"port"`
		}

		flags := &Flags{
			Name:    "test",
			Verbose: true,
			Count:   5,
			Port:    8080,
		}

		result := CommandString(flags)
		expected := []string{"-name", "test", "-verbose", "-count", "5", "-port", "8080"}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("zero values omitted", func(t *testing.T) {
		type Flags struct {
			Name  string `arg:"name"`
			Count int    `arg:"count"`
		}

		flags := &Flags{
			Name: "test",
			// Count is zero, should be omitted
		}

		result := CommandString(flags)
		expected := []string{"-name", "test"}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("duration type", func(t *testing.T) {
		type Flags struct {
			Timeout time.Duration `arg:"timeout"`
		}

		flags := &Flags{
			Timeout: 5 * time.Second,
		}

		result := CommandString(flags)
		expected := []string{"-timeout", "5s"}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("slice types", func(t *testing.T) {
		type Flags struct {
			Tags []string `arg:"tag"`
			IDs  []int    `arg:"id"`
		}

		flags := &Flags{
			Tags: []string{"a", "b"},
			IDs:  []int{1, 2},
		}

		result := CommandString(flags)
		expected := []string{"-tag", "a", "-tag", "b", "-id", "1", "-id", "2"}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("skip dash fields", func(t *testing.T) {
		type Flags struct {
			Name   string `arg:"name"`
			Ignore string `arg:"-"`
		}

		flags := &Flags{
			Name:   "test",
			Ignore: "should be ignored",
		}

		result := CommandString(flags)
		expected := []string{"-name", "test"}

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("nil input", func(t *testing.T) {
		result := CommandString(nil)
		if result != nil {
			t.Errorf("Expected nil, got %v", result)
		}
	})

	t.Run("panic on non-struct", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for non-struct type")
			}
		}()

		var notStruct int
		CommandString(&notStruct)
	})

	t.Run("panic on unsupported type", func(t *testing.T) {
		type Flags struct {
			Invalid map[string]string `arg:"invalid"`
		}

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for unsupported type")
			}
		}()

		flags := &Flags{
			Invalid: map[string]string{"key": "value"},
		}
		CommandString(flags)
	})
}

func TestFlagSetIteratorReset(t *testing.T) {
	// This test verifies that flags are reset between command iterations
	type Flags struct {
		Value string `flage:"value,default,A value"`
	}

	flags := &Flags{}
	fs := FlagSetStruct("cmd", flag.ContinueOnError, flags)

	it := newFlagSetIterator(
		[]string{"cmd", "-value", "first", "cmd"},
		[]*flag.FlagSet{fs},
	)

	// First iteration
	if !it.Next() {
		t.Fatal("Expected first Next() to succeed")
	}
	if flags.Value != "first" {
		t.Errorf("Expected value='first', got '%s'", flags.Value)
	}

	// Second iteration - value should be reset to default
	if !it.Next() {
		t.Fatal("Expected second Next() to succeed")
	}
	if flags.Value != "default" {
		t.Errorf("Expected value='default' after reset, got '%s'", flags.Value)
	}
}

func ExampleMakeUsageWithSubcommands() {
	type DeployCmd struct {
		Env string `flage:"env,production,Environment to deploy to"`
	}

	type Commands struct {
		Deploy DeployCmd `flage-cmd:"deploy,Deploy the application"`
	}

	cmds := &Commands{}
	fss := NewFlagSetsAndDefsFromStruct(cmds, flag.ContinueOnError)

	// Create custom usage function
	flag.Usage = MakeUsageWithSubcommands(HelpInfo{
		Commands:      fss.Defs,
		Flagsets:      fss.Sets,
		Progname:      "myapp",
		About:         "Example application",
		CommandPrefix: "Use one of the following commands:",
	})

	fmt.Println("Usage function created successfully")
	// Output: Usage function created successfully
}
