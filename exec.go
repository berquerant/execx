package execx

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"

	"golang.org/x/sync/errgroup"
)

//go:generate go run github.com/berquerant/goconfig -field "StdoutConsumer func(Token)|StderrConsumer func(Token)|SplitFunc SplitFunc|StdoutWriter io.Writer|StderrWriter io.Writer" -option -output exec_config_generated.go -configOption Option

// Cmd is an external command.
type Cmd struct {
	Args  []string
	Stdin io.Reader
	Dir   string
	Env   Env
}

// Result is [Cmd] execution result.
type Result struct {
	// ExpandedArgs is actual command.
	ExpandedArgs []string
	Stdout       io.Reader
	Stderr       io.Reader
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

func (Cmd) prepareWriter(result *Result, cfg *Config) (stdout, stderr io.Writer) {
	var (
		stdoutBuf bytes.Buffer
		stderrBuf bytes.Buffer
	)
	result.Stdout = &stdoutBuf
	result.Stderr = &stderrBuf

	if cfg.StdoutWriter.IsModified() {
		// write to given writer instead of stdoutBuf
		// result.Stdout will be empty
		stdout = cfg.StdoutWriter.Get()
	} else {
		stdout = &stdoutBuf
	}

	if cfg.StderrWriter.IsModified() {
		// write to given writer instead of stderrBuf
		// result.Stderr will be empty
		stderr = cfg.StderrWriter.Get()
	} else {
		stderr = &stderrBuf
	}
	return
}

// Run executes the command.
//
// If [WithStdoutConsumer] set, you can get the standard output of a command without waiting for the command to finish.
// If [WithStderrConsumer] set, you can get the standard error of a command without waiting for the command to finish.
// [WithSplitFunc] sets the split function for a scanner used in consumers, default is [bufio.ScanLines].
// default is `[]byte("\n")`.
// If [WithStdoutWriter] set, write the standard output into it instead of [Result.Stdout].
// If [WithStderrWriter] set, write the standard error into it instead of [Result.Stderr].
func (c Cmd) Run(ctx context.Context, opt ...Option) (*Result, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	config := NewConfigBuilder().
		StdoutConsumer(func(Token) {}).
		StderrConsumer(func(Token) {}).
		SplitFunc(bufio.ScanLines).
		StdoutWriter(nil).
		StderrWriter(nil).
		Build()
	config.Apply(opt...)

	var (
		cmd, result    = c.prepare(ctx)
		stdout, stderr = c.prepareWriter(result, config)
	)

	if config.StdoutConsumer.IsModified() || config.StderrConsumer.IsModified() {
		return c.runWithLineConsumers(
			ctx,
			config,
			cmd,
			result,
			stdout,
			stderr,
		)
	}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

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
	stdoutWriter, stderrWriter io.Writer,
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
		s := NewScanner(w, r, cfg.SplitFunc.Get(), consumer)
		return s.Scan
	}
	eg, _ := errgroup.WithContext(ctx)
	eg.Go(worker(stdoutWriter, stdout, cfg.StdoutConsumer.Get()))
	eg.Go(worker(stderrWriter, stderr, cfg.StderrConsumer.Get()))

	if err := eg.Wait(); err != nil {
		return nil, fmt.Errorf("%w: read wait", err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("%w: command wait", err)
	}

	return result, nil
}
