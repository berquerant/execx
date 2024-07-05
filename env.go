package execx

import (
	"fmt"
	"os"
	"regexp"
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
	if _, ok := e[key]; ok {
		value = e.Expand(value)
	}
	e[key] = value
}

func (e Env) Merge(other Env) {
	for k, v := range other {
		e.Set(k, v)
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

const expandMaxAttempts = 10

// Expand expands environment variables in target.
func (e Env) Expand(target string) string {
	var (
		originalTarget = target
		result         string
		count          int
		missingKeys    = map[string]bool{}
	)

	get := func(key string) string {
		if value, found := e[key]; found {
			return value
		}
		missingKeys[key] = true
		return fmt.Sprintf("${%s}", key)
	}

	for result = os.Expand(target, get); result != target && count < expandMaxAttempts; count++ {
		target = result
		result = os.Expand(result, get)
	}

	missingKey2Replaces := map[string][]bool{}
	for key := range missingKeys {
		// match with $var or ${var}
		re, err := regexp.Compile(fmt.Sprintf(`\$(%[1]s|\{%[1]s\})`, key))
		if err != nil {
			continue
		}
		found := re.FindAllString(originalTarget, -1)
		replaces := make([]bool, len(found))
		for i, x := range found {
			// true if original variable is raw (e.g. $var)
			replaces[i] = !strings.Contains(x, "{")
		}
		missingKey2Replaces[key] = replaces
	}

	// revert missing variables to original format
	countMap := map[string]int{}
	replaceMissing := func(key string) string {
		index := countMap[key]
		countMap[key]++
		if replaces, found := missingKey2Replaces[key]; found {
			if index < len(replaces) && replaces[index] {
				return fmt.Sprintf("$%s", key)
			}
		}
		return fmt.Sprintf("${%s}", key)
	}
	return os.Expand(result, replaceMissing)
}

// ExpandStrings expands environment variables in multiple targets.
func (e Env) ExpandStrings(target []string) []string {
	result := make([]string, len(target))
	for i, t := range target {
		result[i] = e.Expand(t)
	}
	return result
}
