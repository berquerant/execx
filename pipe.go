package execx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
)

// PipedCmd orchestrates the execution of multiple commands,
// connecting the stdout of one command to the stdin of the next command.
type PipedCmd struct {
	cmds []*exec.Cmd
	// Stdin for the first command.
	Stdin io.Reader
	// Stdout for the last command.
	Stdout io.Writer
	// Stderr for the all commands.
	Stderr io.Writer
}

var (
	ErrNoCmd = errors.New("NoCmd")
)

func NewPipedCmd(cmd ...*exec.Cmd) (*PipedCmd, error) {
	if len(cmd) == 0 {
		return nil, ErrNoCmd
	}
	return &PipedCmd{
		cmds: cmd,
	}, nil
}

func (p *PipedCmd) Start(ctx context.Context) error {
	if len(p.cmds) == 1 {
		c := p.cmds[0]
		c.Stdin = p.Stdin
		c.Stdout = p.Stdout
		c.Stderr = p.Stderr
		return c.Start()
	}

	var stdout io.Reader
	stderr := &ConcurrentWriter{
		Writer: p.Stderr,
	}
	for i, c := range p.cmds {
		c.Stderr = stderr

		switch i {
		case 0:
			c.Stdin = p.Stdin
			outPipe, err := c.StdoutPipe()
			if err != nil {
				return fmt.Errorf("%w: failed to create stdout pipe of cmds[%d]", err, i)
			}
			stdout = outPipe
		case len(p.cmds) - 1:
			c.Stdin = stdout
			c.Stdout = p.Stdout
		default:
			c.Stdin = stdout
			outPipe, err := c.StdoutPipe()
			if err != nil {
				return fmt.Errorf("%w: failed to create stdout pipe of cmds[%d]", err, i)
			}
			stdout = outPipe
		}
	}

	var startedCmds []*exec.Cmd
	for i, c := range p.cmds {
		if err := c.Start(); err != nil {
			_ = p.killCmds(startedCmds...)
			_ = p.waitCmds(startedCmds...)
			return fmt.Errorf("%w: failed to start cmds[%d]", err, i)
		}
		startedCmds = append(startedCmds, c)
	}

	return nil
}

func (PipedCmd) waitCmds(cmd ...*exec.Cmd) error {
	var errs []error
	for i, c := range cmd {
		if err := c.Wait(); err != nil {
			errs = append(errs, fmt.Errorf("%w: failed to wait cmds[%d]", err, i))
		}
	}
	return errors.Join(errs...)
}

func (PipedCmd) killCmds(cmd ...*exec.Cmd) error {
	var errs []error
	for i, c := range cmd {
		if x := c.Process; x != nil {
			if err := x.Kill(); err != nil {
				errs = append(errs, fmt.Errorf("%w: failed to kill cmds[%d]", err, i))
			}
		}
	}
	return errors.Join(errs...)
}

func (p *PipedCmd) Wait() error {
	return p.waitCmds(p.cmds...)
}
