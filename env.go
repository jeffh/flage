package flage

import (
	"context"
	"flag"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
)

// a dictionary key-value lookup interface
type Lookuper interface {
	Lookup(ctx context.Context, key string) ([]string, bool)
	Keys() []string
}

type contextKey int

var (
	isUnderLookupCtxKey contextKey
	isRequiredCtxKey    contextKey
	defvalueValueCtxKey contextKey
)

func withContext(ctx context.Context, required bool, defvalue []string) context.Context {
	if !isUnderLookup(ctx) {
		ctx = context.WithValue(ctx, &isRequiredCtxKey, required)
		if len(defvalue) > 0 {
			ctx = context.WithValue(ctx, &defvalueValueCtxKey, defvalue)
		}
		ctx = context.WithValue(ctx, &isUnderLookupCtxKey, true)
	}
	return ctx
}

func isUnderLookup(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	if v, ok := ctx.Value(&isUnderLookupCtxKey).(bool); ok {
		return v
	}
	return false
}

func contextIsLookingUpRequiredKey(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	if v, ok := ctx.Value(&isRequiredCtxKey).(bool); ok {
		return v
	}
	return false
}

func contextGetDefaultLookupValue(ctx context.Context) ([]string, bool) {
	if ctx == nil {
		return nil, false
	}
	if v, ok := ctx.Value(&defvalueValueCtxKey).([]string); ok {
		return v, true
	}
	return nil, false
}

type Env struct {
	Parent *Env
	Dict   Lookuper
}

func NewEnv(parent *Env, dict Lookuper) *Env {
	return &Env{Parent: parent, Dict: dict}
}

var sysEnv EnvMap
var sysEnvOnce sync.Once

func makeEnvMap() {
	sysEnv = make(EnvMap)
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		key := parts[0]
		sysEnv[key] = append(sysEnv[key], parts[1])
	}
}

func EnvSystem(parent *Env) *Env {
	sysEnvOnce.Do(makeEnvMap)
	return NewEnv(parent, sysEnv)
}

func EnvFile(parent *Env, filepath string) (*Env, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	environ, err := ParseEnvironFile(data)
	envmap := make(EnvMap)
	for _, pairs := range environ {
		envmap[pairs[0]] = append(envmap[pairs[0]], pairs[1])
	}
	return NewEnv(parent, envmap), nil
}

// Represents a map of environment variables. Environment variables are appended when they are set.
// Supports multiple assignments as a flag argument using the format KEY=VALUE.
type EnvMap map[string][]string

var _ flag.Value = (*EnvMap)(nil)

func (e EnvMap) Lookup(_ context.Context, key string) ([]string, bool) {
	if v, ok := e[key]; ok {
		return v, true
	}
	return nil, false
}

func (e EnvMap) Keys() []string {
	keys := make([]string, 0)
	for k := range e {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

type capturingEnvMap struct {
	Usages []EnvUsage
}

type EnvUsage struct {
	Key      string
	Default  []string
	Required bool
}

func (e *capturingEnvMap) UsagesAsEnviron(requiredValue string) [][2]string {
	var env [][2]string
	for _, u := range e.Usages {
		var value [2]string
		if len(u.Default) > 0 {
			for _, v := range u.Default {
				value = [2]string{u.Key, v}
				if !slices.Contains(env, value) {
					env = append(env, value)
				}
			}
		} else if u.Required {
			value = [2]string{u.Key, requiredValue}
			if !slices.Contains(env, value) {
				env = append(env, value)
			}
		} else {
			value = [2]string{u.Key, ""}
			if !slices.Contains(env, value) {
				env = append(env, value)
			}
		}
	}
	return env
}

func (e *capturingEnvMap) Lookup(ctx context.Context, key string) ([]string, bool) {
	if v, ok := contextGetDefaultLookupValue(ctx); ok {
		e.Usages = append(e.Usages, EnvUsage{Key: key, Default: v})
	} else if contextIsLookingUpRequiredKey(ctx) {
		e.Usages = append(e.Usages, EnvUsage{Key: key, Required: true})
	} else {
		e.Usages = append(e.Usages, EnvUsage{Key: key})
	}
	return nil, false // we always will defer to parent
}

func (e *capturingEnvMap) Keys() []string { return nil }

func (e *Env) lookupMany(ctx context.Context, key string) ([]string, bool) {
	ctx = withContext(ctx, false, nil)
	if e == nil {
		return nil, false
	}
	if e.Dict == nil {
		if e.Parent != nil {
			return e.Parent.lookupMany(ctx, key)
		}
		return nil, false
	}
	if v, ok := e.Dict.Lookup(ctx, key); ok {
		return v, true
	} else {
		if e.Parent != nil {
			return e.Parent.lookupMany(ctx, key)
		}
		return nil, false
	}
}
func (e *Env) lookup(ctx context.Context, key string) (string, bool) {
	if v, ok := e.lookupMany(ctx, key); ok {
		if len(v) > 0 {
			return v[0], true
		}
		return "", false
	}
	return "", false
}
func (e *Env) Lookup(key string) (string, bool) {
	return e.lookup(context.Background(), key)
}

func (e *Env) GetOrError(key, errorMsg string) (string, error) {
	ctx := withContext(context.Background(), true, nil)
	if v, ok := e.lookup(ctx, key); ok {
		return v, nil
	}
	return "", fmt.Errorf("require env var %s: %s", key, errorMsg)
}

func (e *Env) GetOr(key, defvalue string) string {
	ctx := withContext(context.Background(), false, []string{defvalue})
	if v, ok := e.lookup(ctx, key); ok {
		return v
	}
	return defvalue
}

func (e *Env) Get(key string) string { return e.GetOr(key, "") }

func (e *Env) Keys() []string {
	keys := e.Dict.Keys()
	if e.Parent != nil {
		parentKeys := e.Parent.Keys()
		for _, k := range keys {
			if !slices.Contains(parentKeys, k) {
				parentKeys = append(parentKeys, k)
			}
		}
	}
	return keys
}

func (e *Env) Map() map[string][]string {
	dict := make(map[string][]string)
	ctx := withContext(context.Background(), false, nil)
	for _, k := range e.Keys() {
		if v, ok := e.lookupMany(ctx, k); ok {
			dict[k] = v
		}
	}
	return dict
}

func (e *Env) Slice() [][2]string {
	pairs := make([][2]string, 0)
	ctx := withContext(context.Background(), false, nil)
	for _, k := range e.Keys() {
		if vs, ok := e.lookupMany(ctx, k); ok {
			for _, v := range vs {
				pairs = append(pairs, [2]string{k, v})
			}
		}
	}
	return pairs
}

func (e *EnvMap) Set(value string) error {
	parts := strings.SplitN(value, "=", 2)
	key := parts[0]
	if len(parts) == 1 {
		(*e)[key] = append((*e)[key], "")
	} else {
		(*e)[key] = append((*e)[key], parts[1])
	}
	return nil
}

func (e *EnvMap) String() string {
	var sb strings.Builder
	for k, vs := range *e {
		for _, v := range vs {
			sb.WriteString(k)
			sb.WriteString("=")
			sb.WriteString(strconv.Quote(v))
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

func (e *EnvMap) Reset() {
	*e = make(EnvMap)
}
