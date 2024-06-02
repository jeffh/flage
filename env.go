package flage

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
)

// a dictionary key-value lookup interface
type Lookuper interface {
	Lookup(ctx context.Context, key string) ([]string, bool)
	Keys() []string
}

var (
	isUnderLookupCtxKey int
	isRequiredCtxKey    int
	defvalueValueCtxKey int
)

func withContext(ctx context.Context, required bool, defvalue string) context.Context {
	if !isUnderLookup(ctx) {
		ctx = context.WithValue(ctx, &isRequiredCtxKey, required)
		if defvalue != "" {
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

func contextGetDefaultLookupValue(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	if v, ok := ctx.Value(&defvalueValueCtxKey).(string); ok {
		return v, true
	}
	return "", false
}

type Env struct {
	Parent *Env
	Dict   Lookuper
}

func NewEnv(parent *Env, dict Lookuper) *Env {
	return &Env{Parent: parent, Dict: dict}
}

func EnvMap(parent *Env, m map[string][]string) *Env {
	return NewEnv(parent, StringsMap(m))
}

var sysEnv StringsMap

func EnvSystem(parent *Env) *Env {
	if sysEnv == nil {
		L := make(StringsMap)
		for _, e := range os.Environ() {
			parts := strings.SplitN(e, "=", 2)
			L[parts[0]] = append(L[parts[0]], parts[1])
		}
		sysEnv = L
	}
	return NewEnv(parent, sysEnv)
}

func EnvFile(parent *Env, filepath string) (*Env, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	environ, err := ParseEnvironFile(data)
	envmap := make(StringsMap)
	for _, pairs := range environ {
		envmap[pairs[0]] = append(envmap[pairs[0]], pairs[1])
	}
	return NewEnv(parent, envmap), nil
}

type StringsMap map[string][]string

func (e StringsMap) Lookup(_ context.Context, key string) ([]string, bool) {
	if v, ok := e[key]; ok {
		return v, true
	}
	return nil, false
}

func (e StringsMap) Keys() []string {
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
	Default  string
	Required bool
}

func (e *capturingEnvMap) UsagesAsEnviron(requiredValue string) [][2]string {
	var env [][2]string
	for _, u := range e.Usages {
		var value [2]string
		if u.Default != "" {
			value = [2]string{u.Key, u.Default}
		} else if u.Required {
			value = [2]string{u.Key, requiredValue}
		} else {
			value = [2]string{u.Key, ""}
		}
		if !slices.Contains(env, value) {
			env = append(env, value)
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

func (e *Env) Lookup(ctx context.Context, key string) (string, bool) {
	ctx = withContext(ctx, false, "")
	if e == nil {
		return "", false
	}
	if e.Dict == nil {
		if e.Parent != nil {
			return e.Parent.Lookup(ctx, key)
		}
		return "", false
	}
	if v, ok := e.Dict.Lookup(ctx, key); ok {
		if len(v) > 0 {
			return v[0], true
		}
		return "", false
	} else {
		if e.Parent != nil {
			return e.Parent.Lookup(ctx, key)
		}
		return "", false
	}
}

func (e *Env) GetOrError(key, errorMsg string) (string, error) {
	ctx := withContext(context.Background(), true, "")
	if v, ok := e.Lookup(ctx, key); ok {
		return v, nil
	}
	return "", fmt.Errorf("require env var %s: %s", key, errorMsg)
}

func (e *Env) GetOr(key, defvalue string) string {
	ctx := withContext(context.Background(), false, defvalue)
	if v, ok := e.Lookup(ctx, key); ok {
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
