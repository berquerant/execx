package execx_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/berquerant/execx"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func TestScript(t *testing.T) {
	assertScriptStdout := func(t *testing.T, want io.Reader, script *execx.Script) {
		var stdout io.Reader
		assert.Nil(t, script.Runner(func(cmd *execx.Cmd) error {
			f, err := cmd.Run(context.TODO(), execx.WithCaptureStdout(true))
			if err != nil {
				return err
			}
			stdout = f.Stdout
			return nil
		}))
		assertReader(t, want, stdout)
	}

	t.Run("Runner", func(t *testing.T) {
		t.Run("PassArg", func(t *testing.T) {
			script := execx.NewScript("echo arg=$*", "sh")
			err := script.Runner(func(cmd *execx.Cmd) error {
				cmd.Args = append(cmd.Args, "ARG")
				r, err := cmd.Run(context.TODO(), execx.WithCaptureStdout(true))
				if err != nil {
					return err
				}
				b, err := io.ReadAll(r.Stdout)
				if err != nil {
					return err
				}
				assert.Equal(t, "arg=ARG\n", string(b))
				return nil
			})
			assert.Nil(t, err)
		})

		t.Run("ConcurrentRace", func(t *testing.T) {
			script := execx.NewScript("echo concurrent", "sh")
			script.KeepScriptFile = true

			var eg errgroup.Group
			for i := 0; i < 8; i++ {
				eg.Go(func() error {
					return script.Runner(func(cmd *execx.Cmd) error {
						r, err := cmd.Run(context.TODO(), execx.WithCaptureStdout(true))
						if err != nil {
							return err
						}
						assertReader(t, bytes.NewBufferString("concurrent\n"), r.Stdout)
						return nil
					})
				})
			}
			assert.Nil(t, eg.Wait())
		})

		t.Run("ChangeScript", func(t *testing.T) {
			t.Run("KeepScript", func(t *testing.T) {
				script := execx.NewScript("echo keep", "sh")
				script.KeepScriptFile = true

				assertScriptStdout(t, bytes.NewBufferString("keep\n"), script)
				script.Content = "echo changed"
				assertScriptStdout(t, bytes.NewBufferString("keep\n"), script)
			})

			script := execx.NewScript("", "sh")

			for _, tc := range []struct {
				name    string
				content string
				env     execx.Env
				want    io.Reader
			}{
				{
					name:    "echo 1st",
					content: "echo $v",
					env:     execx.EnvFromSlice([]string{"v=1"}),
					want:    bytes.NewBufferString("1\n"),
				},
				{
					name:    "change env",
					content: "echo $v",
					env:     execx.EnvFromSlice([]string{"v=2"}),
					want:    bytes.NewBufferString("2\n"),
				},
				{
					name:    "change script",
					content: "echo ${v}!",
					env:     execx.EnvFromSlice([]string{"v=2"}),
					want:    bytes.NewBufferString("2!\n"),
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					script.Content = tc.content
					script.Env.Merge(tc.env)
					assertScriptStdout(t, tc.want, script)
				})
			}
		})

		t.Run("Run", func(t *testing.T) {
			for _, tc := range []struct {
				name             string
				script           *execx.Script
				env              execx.Env
				opt              []execx.Option
				stdin            io.Reader
				wantStdout       io.Reader
				wantStderr       io.Reader
				wantOutLines     []string
				wantErrLines     []string
				withStdoutWriter bool
				withStderrWriter bool
			}{
				{
					name: "with writers",
					script: execx.NewScript(`cat -
echo err1 1 > /dev/stderr
echo err2 2 > /dev/stderr`, "bash"),
					stdin:            bytes.NewBufferString("line1 1\nline2 2\n"),
					wantStdout:       bytes.NewBufferString("line1 1\nline2 2\n"),
					wantStderr:       bytes.NewBufferString("err1 1\nerr2 2\n"),
					wantOutLines:     []string{"line1 1", "line2 2"},
					wantErrLines:     []string{"err1 1", "err2 2"},
					withStdoutWriter: true,
					withStderrWriter: true,
				},
				{
					name: "split by words",
					script: execx.NewScript(`cat -
echo err1 1 > /dev/stderr
echo err2 2 > /dev/stderr`, "bash"),
					opt: []execx.Option{
						execx.WithDelim(' '),
					},
					stdin:        bytes.NewBufferString("line1 1 line2 2"),
					wantStdout:   bytes.NewBufferString("line1 1 line2 2"),
					wantStderr:   bytes.NewBufferString("err1 1\nerr2 2\n"),
					wantOutLines: []string{"line1", "1", "line2", "2"},
					wantErrLines: []string{"err1", "1\nerr2", "2\n"},
				},
				{
					name: "plain",
					script: execx.NewScript(`cat -
echo err1 1 > /dev/stderr
echo err2 2 > /dev/stderr`, "bash"),
					stdin: bytes.NewBufferString("line1 1\nline2 2\n"),
					// trailing newlines are removed due to consumers
					wantStdout:   bytes.NewBufferString("line1 1\nline2 2\n"),
					wantStderr:   bytes.NewBufferString("err1 1\nerr2 2\n"),
					wantOutLines: []string{"line1 1", "line2 2"},
					wantErrLines: []string{"err1 1", "err2 2"},
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					var (
						gotOutLines []string
						gotErrLines []string
						gotResult   *execx.Result

						gotStdoutWriter bytes.Buffer
						gotStderrWriter bytes.Buffer
					)

					if !assert.Nil(t, tc.script.Runner(func(cmd *execx.Cmd) error {
						cmd.Stdin = tc.stdin
						cmd.Env.Merge(tc.env)

						opt := append(tc.opt, execx.WithStdoutConsumer(func(x execx.Token) {
							gotOutLines = append(gotOutLines, x.String())
						}), execx.WithStderrConsumer(func(x execx.Token) {
							gotErrLines = append(gotErrLines, x.String())
						}))
						if tc.withStdoutWriter {
							cmd.Stdout = &gotStdoutWriter
						} else {
							opt = append(opt, execx.WithCaptureStdout(true))
						}
						if tc.withStderrWriter {
							cmd.Stderr = &gotStderrWriter
						} else {
							opt = append(opt, execx.WithCaptureStderr(true))
						}

						got, err := cmd.Run(context.TODO(), opt...)
						if err != nil {
							return err
						}

						gotResult = got
						return nil
					})) {
						return
					}

					assert.Equal(t, tc.wantOutLines, gotOutLines, "out lines")
					assert.Equal(t, tc.wantErrLines, gotErrLines, "err lines")

					if tc.withStdoutWriter {
						assertReader(t, tc.wantStdout, &gotStdoutWriter, "stdout")
						assertReader(t, bytes.NewBufferString(""), gotResult.Stdout, "result stdout")
					} else {
						assertReader(t, tc.wantStdout, gotResult.Stdout, "result stdout")
					}

					if tc.withStderrWriter {
						assertReader(t, tc.wantStderr, &gotStderrWriter, "stderr")
						assertReader(t, bytes.NewBufferString(""), gotResult.Stderr, "result stderr")
					} else {
						assertReader(t, tc.wantStderr, gotResult.Stderr, "result stderr")
					}
				})
			}

			t.Run("Example", func(t *testing.T) {
				const (
					content = `echo line1
echo ${line2}
echo ${test_script_append_env1}
cat -
echo line3 > /dev/stderr
echo line4 > /dev/stderr`
					stdin      = "from stdin\n"
					wantStdout = `line1
LINE2
added:append1
from stdin
`
					wantStderr = `line3
line4
`
				)
				script := execx.NewScript(content, "sh")
				script.Env.Set("test_script_append_env1", "append1")
				script.Env.Set("line2", "LINE2")
				script.Env.Set("test_script_append_env1", "added:${test_script_append_env1}")

				var (
					stdout io.Reader
					stderr io.Reader
				)

				assert.Nil(t, script.Runner(func(cmd *execx.Cmd) error {
					cmd.Stdin = bytes.NewBufferString(stdin)
					r, err := cmd.Run(context.TODO(), execx.WithCaptureStdout(true), execx.WithCaptureStderr(true))
					if err != nil {
						return err
					}
					stdout = r.Stdout
					stderr = r.Stderr
					return nil
				}))

				assertReader(t, bytes.NewBufferString(wantStdout), stdout)
				assertReader(t, bytes.NewBufferString(wantStderr), stderr)
			})
		})
	})
}
