package execx

import (
	"fmt"
	"strings"

	"github.com/berquerant/execx/internal"
)

// Task is a named script function.
type Task struct {
	Name   string
	Script string
}

func NewTask(name, script string) *Task {
	return &Task{
		Name:   name,
		Script: script,
	}
}

func (t Task) String() string {
	return fmt.Sprintf(`%s() {
%s
}`, t.Name, internal.IndentN(t.Script, 2))
}

type Tasks []*Task

func NewTasks() Tasks {
	return []*Task{}
}

func (t Tasks) Add(task *Task) Tasks {
	return append(t, task)
}

func (t Tasks) String() string {
	ss := make([]string, len(t))
	for i, x := range t {
		ss[i] = x.String()
	}
	return strings.Join(ss, "\n")
}

type ExecutableTasks struct {
	Tasks      Tasks
	Env        Env
	Entrypoint []string // task names to execute
}

func NewExecutableTasks(
	tasks Tasks,
	env Env,
	entrypoint ...string,
) *ExecutableTasks {
	return &ExecutableTasks{
		Tasks:      tasks,
		Env:        env,
		Entrypoint: entrypoint,
	}
}

func (t ExecutableTasks) IntoScript(shell string, arg ...string) *Script {
	s := NewScript(
		t.asString(false),
		shell,
		arg...,
	)
	s.Env = t.Env
	return s
}

func (t ExecutableTasks) String() string {
	return t.asString(true)
}

func (t ExecutableTasks) asString(dry bool) string {
	var (
		b strings.Builder
		w = func(f string, a ...any) {
			b.WriteString(fmt.Sprintf(f, a...) + "\n")
		}
	)

	if dry {
		for k, v := range t.Env {
			w(`%s="%s"`, k, internal.EscapeQuote(v))
		}
	}
	w(t.Tasks.String())
	w(strings.Join(t.Entrypoint, "\n"))

	return b.String()
}
