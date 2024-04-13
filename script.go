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

// Run executes the script.
//
// See [Cmd.Run for opt.
func (s Script) Run(ctx context.Context, opt ...Option) (*Result, error) {
	fp, err := os.CreateTemp("", "execx")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = os.Remove(fp.Name())
	}()

	cmd := New(s.Shell[0], append(s.Shell[1:], fp.Name())...)
	cmd.Env.Merge(s.Env)
	cmd.Stdin = s.Stdin
	cmd.Dir = s.Dir

	if _, err := io.WriteString(fp, cmd.Env.Expand(s.Content)); err != nil {
		return nil, err
	}
	if err := os.Chmod(fp.Name(), 0755); err != nil {
		return nil, err
	}

	return cmd.Run(ctx, opt...)
}
