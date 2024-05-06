package execx

import (
	"fmt"
	"os"
	"sync"
)

// Script is an executable script, set of commands.
type Script struct {
	// Shell executes Content.
	Shell []string
	// Content is a set of commands.
	Content string
	Env     Env
	// If KeepScriptFile is true, then do not regenerate script files,
	// and not reflect changes in Content and Env when calling Runner.
	KeepScriptFile bool

	script *scriptFile
	mux    *sync.Mutex
}

// NewScript creates a new [Script].
func NewScript(content string, shell string, arg ...string) *Script {
	var mux sync.Mutex
	return &Script{
		Shell:   append([]string{shell}, arg...),
		Content: content,
		Env:     NewEnv(),
		mux:     &mux,
	}
}

// Close removes a temporary script file.
func (s *Script) Close() error {
	if s.script != nil {
		x := s.script
		s.script = nil
		return x.close()
	}
	return nil
}

func (s *Script) isExecutable() bool {
	return s.script != nil && s.script.isExecutable()
}

func (s *Script) requireNewScript() bool {
	return !s.isExecutable() || !s.KeepScriptFile
}

func (s *Script) newScript() error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if !s.requireNewScript() {
		return nil
	}

	content := s.Env.Expand(s.Content)
	f, err := newScriptFile(content)
	if err != nil {
		return err
	}
	s.script = f
	return nil
}

func (s *Script) prepare() (*Cmd, error) {
	if err := s.newScript(); err != nil {
		return nil, err
	}
	cmd := New(s.Shell[0], append(s.Shell[1:], s.script.path)...)
	cmd.Env.Merge(s.Env)
	return cmd, nil
}

// Runner creates a new [Cmd] and pass it to f.
func (s *Script) Runner(f func(*Cmd) error) error {
	cmd, err := s.prepare()
	if err != nil {
		return fmt.Errorf("%w: prepare script file", err)
	}

	if !s.KeepScriptFile {
		defer s.Close()
	}
	return f(cmd)
}

type scriptFile struct {
	path string
}

func newScriptFile(content string) (*scriptFile, error) {
	f, err := os.CreateTemp("", "execx")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return nil, err
	}
	if err := os.Chmod(f.Name(), 0755); err != nil {
		return nil, err
	}
	return &scriptFile{
		path: f.Name(),
	}, nil
}

func (s scriptFile) close() error {
	return os.Remove(s.path)
}

func (s scriptFile) isExecutable() bool {
	return isExecutable(s.path)
}

func isExecutable(path string) bool {
	f, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !f.IsDir() && f.Mode() == 0755
}
