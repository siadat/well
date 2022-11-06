package interpreter

import (
	"bytes"
	"fmt"
	"strings"
)

type Environment interface {
	Get(string) (Object, error)
	Set(string, Object) error
	NewScope() Environment
	SetDebug(bool)
}

type mapEnv struct {
	parent Environment
	store  map[string]Object
	debug  bool
}

func NewEnvironment() Environment {
	env := make(map[string]Object)
	env["true"] = &Boolean{Value: true}
	env["false"] = &Boolean{Value: false}
	return &mapEnv{
		parent: nil,
		store:  env,
	}
}

func (env *mapEnv) SetDebug(v bool) {
	env.debug = v
}

func (env *mapEnv) Get(name string) (Object, error) {
	if env.debug {
		var keys []string
		for k := range env.store {
			keys = append(keys, fmt.Sprintf("%q", k))
		}

		var buf bytes.Buffer
		fmt.Fprintf(&buf, "DEBUG: looking for %q in [", name)
		fmt.Fprintf(&buf, strings.Join(keys, " "))
		fmt.Fprintf(&buf, "]")
		fmt.Println(buf.String())
	}
	if obj, ok := env.store[name]; ok {
		return obj, nil
	}
	return nil, fmt.Errorf("undefined object %q", name)
}

func (env *mapEnv) Set(name string, obj Object) error {
	if env.debug {
		fmt.Printf("DEBUG: setting %q to %#v\n", name, obj)
	}
	if _, ok := env.store[name]; ok {
		return fmt.Errorf("duplicate env key %q", name)
	}
	env.store[name] = obj
	return nil
}

func (env *mapEnv) NewScope() Environment {
	newEnv := &mapEnv{
		parent: env,
		store:  make(map[string]Object),
	}
	for k, v := range env.store {
		newEnv.store[k] = v
	}
	return newEnv
}
