package execx

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sync/errgroup"
)

//go:generate go tool goconfig -field "StdoutConsumer func(Token)|StderrConsumer func(Token)|Delim byte|CaptureStdout bool|CaptureStderr bool" -option -output exec_config_generated.go -configOption Option

// Cmd is an external command.
type Cmd struct {
	Args   []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Dir    string
	Env    Env
}

// Result is [Cmd] execution result.
type Result struct {
	// ExpandedArgs is actual command.
	ExpandedArgs []string
	// If [WithCaptureStdout] is true, records stdout.
	Stdout io.Reader
	// If [WithCaptureStderr] is true, records stderr.
	Stderr io.Reader
}

type SplitFunc = bufio.SplitFunc

type Token interface {
	String() string
	Bytes() []byte
}

var (
	_ Token = token(nil)
)

type token []byte

func (t token) String() string {
	return string(t)
}

func (t token) Bytes() []byte {
	return []byte(t)
}

// Create a new [Cmd].
//
// Set the current directory to [Cmd.Dir], current environment variables to [Cmd.Env].
func New(name string, arg ...string) *Cmd {
	return &Cmd{
		Args: append([]string{name}, arg...),
		Dir:  ".",
		Env:  NewEnv(),
	}
}

func (c Cmd) IntoExecCmd(ctx context.Context) *exec.Cmd {
	cmd, _ := c.prepare(ctx)
	cmd.Stdout = c.Stdout
	cmd.Stderr = c.Stderr
	return cmd
}

func (c Cmd) prepare(ctx context.Context) (*exec.Cmd, *Result) {
	args := c.Env.ExpandStrings(c.Args)

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Stdin = c.Stdin
	cmd.Dir = c.Dir
	cmd.Env = c.Env.IntoSlice()

	result := &Result{
		ExpandedArgs: args,
	}
	return cmd, result
}

type cmdWriters struct {
	stdout, stderr io.Writer
}

func (c Cmd) prepareWriters(result *Result, cfg *Config) *cmdWriters {
	r := &cmdWriters{
		stdout: c.Stdout,
		stderr: c.Stderr,
	}

	var (
		stdout, stderr bytes.Buffer
	)
	result.Stdout = &stdout
	result.Stderr = &stderr
	if cfg.CaptureStdout.Get() {
		r.stdout = &stdout
	}
	if cfg.CaptureStderr.Get() {
		r.stderr = &stderr
	}

	return r
}

// Run executes the command.
//
// If [WithStdoutConsumer] set, you can get the standard output of a command without waiting for the command to finish.
// If [WithStderrConsumer] set, you can get the standard error of a command without waiting for the command to finish.
// [WithSplitFunc] sets the split function for a scanner used in consumers, default is [bufio.ScanLines].
// default is `[]byte("\n")`.
func (c Cmd) Run(ctx context.Context, opt ...Option) (*Result, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	config := NewConfigBuilder().
		StdoutConsumer(func(Token) {}).
		StderrConsumer(func(Token) {}).
		Delim('\n').
		CaptureStdout(false).
		CaptureStderr(false).
		Build()
	config.Apply(opt...)

	cmd, result := c.prepare(ctx)
	writers := c.prepareWriters(result, config)

	if config.StdoutConsumer.IsModified() || config.StderrConsumer.IsModified() {
		return c.runWithLineConsumers(
			ctx,
			config,
			cmd,
			result,
			writers,
		)
	}

	cmd.Stdout = writers.stdout
	cmd.Stderr = writers.stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w: command run", err)
	}
	return result, nil
}

func (c Cmd) runWithLineConsumers(
	ctx context.Context,
	cfg *Config,
	cmd *exec.Cmd,
	result *Result,
	writers *cmdWriters,
) (*Result, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("%w: stdout pipe", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("%w: stderr pipe", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("%w: command start", err)
	}

	worker := func(w io.Writer, r io.Reader, consumer func(Token)) func() error {
		s := NewScanner(w, r, cfg.Delim.Get(), consumer)
		return s.Scan
	}
	eg, _ := errgroup.WithContext(ctx)
	eg.Go(worker(writers.stdout, stdout, cfg.StdoutConsumer.Get()))
	eg.Go(worker(writers.stderr, stderr, cfg.StderrConsumer.Get()))

	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("%w: read wait", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("%w: command wait", err)
	}

	return result, nil
}

// Exec invokes execve(2).
// This ignores Cmd.Stdin.
func (c Cmd) Exec() error {
	bin, err := exec.LookPath(c.Args[0])
	if err != nil {
		return fmt.Errorf("%w: exec look path %s", err, c.Args[0])
	}
	if err := os.Chdir(c.Dir); err != nil {
		return fmt.Errorf("%w: exec chdir %s", err, c.Dir)
	}
	return syscall.Exec(bin, c.Args, c.Env.IntoSlice())
}
