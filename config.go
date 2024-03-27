package flage

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"os"
	"path"
	"strings"

	"github.com/google/shlex"
)

// TemplateConfigRenderer manages rendering of config files that is used by other helper functions:
//
// - ParseConfigFile
// - ReadConfigFile
// - ReadConfig
// - ExtractEnvKeysFromConfigFile
// - PreviewConfigFile
// - PreviewConfig
//
// This allows you to set your own template variables and functions. Use DefaultTemplateFuncs to get
// default template functions used.
type TemplateConfigRenderer struct {
	Data  map[string]string
	Funcs template.FuncMap
}

func DefaultTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"env": func(key string) string {
			return os.Getenv(key)
		},
		"envOrError": func(key, msg string) (string, error) {
			if v, ok := os.LookupEnv(key); ok {
				return v, nil
			}
			if msg != "" {
				return "", fmt.Errorf("require env var: %q: %s", key, msg)
			} else {
				return "", fmt.Errorf("require env var: %q", key)
			}
		},
		"envOr": func(key, def string) string {
			if v, ok := os.LookupEnv(key); ok {
				return v
			}
			return def
		},
		"json": func(v interface{}) (string, error) {
			b, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
	}
}
func DefaultPreviewTemplateFuncs(keys *[][2]string) template.FuncMap {
	addKey := func(k, defvalue string) {
		for _, v := range *keys {
			if v[0] == k {
				return
			}
		}
		*keys = append(*keys, [2]string{k, defvalue})
	}
	return template.FuncMap{
		"env": func(key string) string {
			addKey(key, "")
			return os.Getenv(key)
		},
		"envOrError": func(key, msg string) (string, error) {
			addKey(key, "NEEDS_TO_BE_FILLED")
			return os.Getenv(key), nil
		},
		"envOr": func(key, def string) string {
			addKey(key, def)
			if v, ok := os.LookupEnv(key); ok {
				return v
			}
			return def
		},
		"json": func(v interface{}) (string, error) {
			b, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
	}
}

func (r *TemplateConfigRenderer) Render(data string, configPath string) (string, error) {
	if r.Data == nil {
		r.Data = make(map[string]string)
	}
	r.Data["configDir"] = path.Dir(configPath)

	if r.Funcs == nil {
		r.Funcs = DefaultTemplateFuncs()
	}

	tmpl, err := template.New("config").Funcs(r.Funcs).Parse(data)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, r.Data)
	if err != nil {
		return "", fmt.Errorf("failed to interpolate config file: %w", err)
	}
	return buf.String(), nil
}

func fileToCmdlineArgs(s string) string {
	var out bytes.Buffer
	for _, line := range strings.Split(s, "\n") {
		before, _, found := strings.Cut(line, "#")
		if found && strings.TrimSpace(before) == "" {
			continue
		}
		out.WriteString(line)
		out.WriteString(" ")
	}
	return out.String()[0 : out.Len()-1]
}

// PreviewConfig returns the contents of data by passing that is ready to pass to flag.FlagSet.Parse(...)
func PreviewConfig(configPath string, data string) (string, error) {
	r := TemplateConfigRenderer{}
	out, err := r.Render(data, configPath)
	if err != nil {
		return "", err
	}
	return fileToCmdlineArgs(out), nil
}

// PreviewConfigFile returns the contents of file by passing that is ready to pass to flag.FlagSet.Parse(...)
func PreviewConfigFile(file string) (string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("failed to open config file: %s: %w", file, err)
	}
	args, err := PreviewConfig(file, string(data))
	if err != nil {
		return "", fmt.Errorf("failed to read config file: %s: %w", file, err)
	}
	return args, nil
}

func observeEnvReadsTemplate(data string, configPath string) ([][2]string, error) {
	var keys [][2]string
	r := TemplateConfigRenderer{
		Funcs: DefaultPreviewTemplateFuncs(&keys),
	}
	_, err := r.Render(data, configPath)
	return keys, err
}

// ExtractEnvKeysFromConfigFile returns the environment keys that are read from a given config file.
func ExtractEnvKeysFromConfigFile(file string) ([][2]string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %s: %w", file, err)
	}
	keys, err := observeEnvReadsTemplate(string(data), file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %s: %w", file, err)
	}
	return keys, nil
}

// ReadConfigFile returns args to parse from a given file path
//
// Comments are lines that start with a '#' (and not in the middle).
//
// The configuration file format assumes:
//
// - any line that starts with a '#' is a comment and ignored (ignoring leading whitespace)
// - all lines are replaced with spaces
// - then input is passed to shlex.Split to split by shell parsing rules
//
// This is also templated, so the following variables are available:
//
// - configDir: the directory of the config file
//
// The following functions are available:
//
// - env(key): returns the value of the environment variable
// - envOr(key, def): returns the value of the environment variable or the default value
// - envOrError(key, msg): returns the value of the environment variable or errors with msg
//
// Example:
//
//	-load {{.configDir}}/file.txt -secret {{env "SECRET"}} -port {{envOr "PORT" "8080"}}
func ReadConfig(configPath string, data string) ([]string, error) {
	r := TemplateConfigRenderer{}
	out, err := r.Render(data, configPath)
	if err != nil {
		return nil, err
	}
	cmdline := fileToCmdlineArgs(out)
	args, err := shlex.Split(cmdline)
	if err != nil {
		narg := len(args)
		context := strings.Join(args[max(0, narg-4):narg], " ")
		return nil, fmt.Errorf("failed to parse config file: %w (maybe right after %q)", err, context)
	}
	return args, nil
}

// ReadConfigFile returns args to parse from a given file.
// If you have already opened the file, use ReadConfig instead.
//
// Comments are lines that start with a '#' (and not in the middle).
//
// The configuration file format assumes:
//
// - any line that starts with a '#' is a comment and ignored (ignoring leading whitespace)
// - all lines are replaced with spaces
// - then input is passed to shlex.Split to split by shell parsing rules
//
// This is also templated, so the following variables are available:
//
// - configDir: the directory of the config file
//
// The following functions are available:
//
// - env(key): returns the value of the environment variable
// - envOr(key, def): returns the value of the environment variable or the default value
//
// Example:
//
//	-load {{.configDir}}/file.txt -secret {{env "SECRET"}} -port {{envOr "PORT" "8080"}}
func ReadConfigFile(file string) ([]string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %s: %w", file, err)
	}
	args, err := ReadConfig(file, string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %s: %w", file, err)
	}
	return args, nil
}

// ParseConfigFile reads a given filepath and applies command line parsing to it.
// This is a quick and easy way to provide file-based configuration, a la pip.
//
// If you want to only read the file, use ReadConfig instead.
//
// Comments are lines that start with a '#' (and not in the middle).
//
// The configuration file format assumes:
//
// - any line that starts with a '#' is a comment and ignored (ignoring leading whitespace)
// - all lines are replaced with spaces
// - then input is passed to shlex.Split to split by shell parsing rules
//
// This is also templated, so the following variables are available:
//
// - configDir: the directory of the config file
//
// The following functions are available:
//
// - env(key): returns the value of the environment variable
// - envOr(key, def): returns the value of the environment variable or the default value
// - envOrError(key, msg): returns the value of the environment variable or errors with msg
//
// Example:
//
//	-load {{.configDir}}/file.txt -secret {{env "SECRET"}} -port {{envOr "PORT" "8080"}}
func ParseConfigFile(file string) error {
	args, err := ReadConfigFile(file)
	if err != nil {
		return err
	}
	return flag.CommandLine.Parse(args)
}

// ParseEnvFile reads a file like an enviroment file.
//
// File format:
//
//   - "#" are to-end-of-line comments
//   - each line is in KEY=VALUE format
//   - any line without an equal sign is ignored
func ParseEnvFile(file string) ([][2]string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
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

// SourceEnvFile parses a env file via ParseEnvFile and sets the process' environment variables
func SourceEnvFile(file string) error {
	pairs, err := ParseEnvFile(file)
	if err != nil {
		return err
	}
	for _, pair := range pairs {
		os.Setenv(pair[0], pair[1])
	}
	return nil
}
