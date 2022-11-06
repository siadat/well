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
	Global() Environment
	SetDebug(bool)
}

type mapEnv struct {
	global Environment
	parent Environment
	store  map[string]Object
	debug  bool
}

func NewEnvironment() Environment {
	var store = make(map[string]Object)
	store["true"] = &Boolean{Value: true}
	store["false"] = &Boolean{Value: false}
	var env = &mapEnv{
		parent: nil,
		store:  store,
	}
	env.global = env
	return env
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
		fmt.Fprintf(&buf, "[environment] Get %q in [", name)
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
		fmt.Printf("[environment] Set %q to %#v\n", name, obj)
	}
	if _, ok := env.store[name]; ok {
		return fmt.Errorf("duplicate env key %q", name)
	}
	env.store[name] = obj
	return nil
}

func (env *mapEnv) Global() Environment {
	return env.global
}

func (env *mapEnv) NewScope() Environment {
	if env.debug {
		fmt.Printf("[environment] NewScope\n")
	}
	newEnv := &mapEnv{
		global: env.global,
		parent: env,
		store:  make(map[string]Object),
		debug:  env.debug,
	}
	for k, v := range env.store {
		newEnv.store[k] = v
	}
	return newEnv
}
