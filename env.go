package execx

import (
	"fmt"
	"os"
	"strings"
)

// Env represents a set of environment variables.
type Env map[string]string

// EnvFromEnviron creates a new [Env] from [os.Environ].
func EnvFromEnviron() Env {
	return EnvFromSlice(os.Environ())
}

// NewEnv creates a new empty [Env].
func NewEnv() Env {
	return Env(make(map[string]string))
}

// EnvFromSlice creates a new [Env] from strings, in the form "key=value".
func EnvFromSlice(envSlice []string) Env {
	env := NewEnv()
	for _, x := range envSlice {
		xs := strings.SplitN(x, "=", 2)
		if len(xs) != 2 {
			continue
		}
		env.Set(xs[0], xs[1])
	}
	return env
}

func (e Env) Get(key string) (string, bool) {
	v, ok := e[key]
	return v, ok
}

func (e Env) Set(key, value string) {
	e[key] = value
}

func (e Env) Merge(other Env) {
	for k, v := range other {
		e[k] = v
	}
}

// IntoSlice converts into os.Environ format.
func (e Env) IntoSlice() []string {
	var (
		i      int
		result = make([]string, len(e))
	)
	for k, v := range e {
		result[i] = fmt.Sprintf("%s=%s", k, v)
		i++
	}
	return result
}

func (e Env) get(key string) string {
	if value, found := e[key]; found {
		return value
	}
	// no changes
	return fmt.Sprintf("${%s}", key)
}

const expandMaxAttempts = 10

// Expand expands environment variables in target.
func (e Env) Expand(target string) string {
	var (
		result string
		count  int
	)
	for result = os.Expand(target, e.get); result != target && count < expandMaxAttempts; count++ {
		target = result
		result = os.Expand(result, e.get)
	}
	return result
}

// ExpandStrings expands environment variables in multiple targets.
func (e Env) ExpandStrings(target []string) []string {
	result := make([]string, len(target))
	for i, t := range target {
		result[i] = e.Expand(t)
	}
	return result
}
