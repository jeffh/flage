package flage

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/google/shlex"
)

func fileToCmdlineArgs(s string) string {
	var out bytes.Buffer
	r := bufio.NewScanner(strings.NewReader(s))
	for r.Scan() {
		line := r.Text()
		before, _, found := strings.Cut(line, "#")
		if found && strings.TrimSpace(before) == "" {
			continue
		}
		out.WriteString(line)
		out.WriteString(" ")
	}
	return out.String()[0 : out.Len()-1]
}

// ParseConfigFile reads a config file and converts it to command line arguments.
// This is a quick and easy way to provide file-based configuration, a la pip.
//
// Comments are lines that start with a '#' (and not in the middle).
// If env is given to nil, it defaults to EnvSystem(nil).
//
// The configuration file format assumes:
//
// - any line that starts with a '#' is a comment and ignored (ignoring leading whitespace)
// - all lines are replaced with spaces
// - then input is passed to shlex.Split to split by shell parsing rules
//
// Example:
//
//	-load ./file.txt -secret $SECRET
func ParseConfigFile(fileContents string) ([]string, error) {
	cmdline := fileToCmdlineArgs(fileContents)
	args, err := shlex.Split(cmdline)
	if err != nil {
		narg := len(args)
		context := strings.Join(args[max(0, narg-4):narg], " ")
		return nil, fmt.Errorf("failed to parse config file: %w (maybe right after %q)", err, context)
	}
	return args, nil
}

// ReadConfigFile reads a given filepath and converts it to command line arguments.
// This is a quick and easy way to provide file-based configuration, a la pip.
//
// If you have the contents of the file already, use ParseConfigFile instead.
//
// Comments are lines that start with a '#' (and not in the middle).
// If env is given to nil, it defaults to EnvSystem(nil).
//
// The configuration file format assumes:
//
// - any line that starts with a '#' is a comment and ignored (ignoring leading whitespace)
// - all lines are replaced with spaces
// - then input is passed to shlex.Split to split by shell parsing rules
//
// Example:
//
//	-load ./file.txt -secret $SECRET
func ReadConfigFile(file string) ([]string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return ParseConfigFile(string(data))
}

// ParseEnvironFile reads bytes like an enviroment file.
//
// File format:
//
//   - "#" are to-end-of-line comments and must be at the start of the line
//   - each line is in KEY=VALUE format
//   - any line without an equal sign is ignored
func ParseEnvironFile(data []byte) ([][2]string, error) {
	lines := strings.Split(string(data), "\n")
	var res [][2]string
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		res = append(res, [2]string{parts[0], parts[1]})
	}
	return res, nil
}

// ReadEnvironFile reads a file like an enviroment file.
//
// File format:
//
//   - "#" are to-end-of-line comments
//   - each line is in KEY=VALUE format
//   - any line without an equal sign is ignored
func ReadEnvironFile(file string) ([][2]string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return ParseEnvironFile(data)
}
