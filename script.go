package execx

import (
	"context"
	"io"
	"os"
)

// Script is an executable script, set of commands.
type Script struct {
	// Shell executes Content.
	Shell []string
	// Content is a set of commands.
	Content string
	Stdin   io.Reader
	Dir     string
	Env     Env
}

// NewScript creates a new [Script].
func NewScript(content string, shell string, arg ...string) *Script {
	return &Script{
		Shell:   append([]string{shell}, arg...),
		Content: content,
		Env:     NewEnv(),
	}
}

// Runner creates a new [Cmd] and pass it to f.
func (s Script) Runner(f func(*Cmd) error) error {
	fp, err := os.CreateTemp("", "execx")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(fp.Name())
	}()

	cmd := New(s.Shell[0], append(s.Shell[1:], fp.Name())...)
	cmd.Env.Merge(s.Env)
	cmd.Stdin = s.Stdin
	cmd.Dir = s.Dir

	if _, err := io.WriteString(fp, cmd.Env.Expand(s.Content)); err != nil {
		return err
	}
	if err := os.Chmod(fp.Name(), 0755); err != nil {
		return err
	}

	return f(cmd)
}

// Run executes the script.
//
// See [Cmd.Run] for opt.
func (s Script) Run(ctx context.Context, opt ...Option) (*Result, error) {
	var result *Result
	if err := s.Runner(func(cmd *Cmd) error {
		r, err := cmd.Run(ctx, opt...)
		result = r
		return err
	}); err != nil {
		return nil, err
	}
	return result, nil
}
