package execx_test

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/berquerant/execx"
	"github.com/stretchr/testify/assert"
)

func TestScript(t *testing.T) {
	assertScriptStdout := func(t *testing.T, want io.Reader, script *execx.Script) {
		var stdout io.Reader
		assert.Nil(t, script.Runner(func(cmd *execx.Cmd) error {
			f, err := cmd.Run(context.TODO())
			if err != nil {
				return err
			}
			stdout = f.Stdout
			return nil
		}))
		assertReader(t, want, stdout)
	}

	t.Run("Runner", func(t *testing.T) {
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
					wantStdout:       bytes.NewBufferString("line1 1\nline2 2"),
					wantStderr:       bytes.NewBufferString("err1 1\nerr2 2"),
					wantOutLines:     []string{"line1 1", "line2 2"},
					wantErrLines:     []string{"err1 1", "err2 2"},
					withStdoutWriter: true,
					withStderrWriter: true,
				},
				{
					name: "split by words joint by space",
					script: execx.NewScript(`cat -
echo err1 1 > /dev/stderr
echo err2 2 > /dev/stderr`, "bash"),
					opt: []execx.Option{
						execx.WithSplitFunc(bufio.ScanWords),
						execx.WithSplitSeparator([]byte(" ")),
					},
					stdin:        bytes.NewBufferString("line1 1\nline2 2\n"),
					wantStdout:   bytes.NewBufferString("line1 1 line2 2"),
					wantStderr:   bytes.NewBufferString("err1 1 err2 2"),
					wantOutLines: []string{"line1", "1", "line2", "2"},
					wantErrLines: []string{"err1", "1", "err2", "2"},
				},
				{
					name: "split by words",
					script: execx.NewScript(`cat -
echo err1 1 > /dev/stderr
echo err2 2 > /dev/stderr`, "bash"),
					opt: []execx.Option{
						execx.WithSplitFunc(bufio.ScanWords),
					},
					stdin:        bytes.NewBufferString("line1 1\nline2 2\n"),
					wantStdout:   bytes.NewBufferString("line1\n1\nline2\n2"),
					wantStderr:   bytes.NewBufferString("err1\n1\nerr2\n2"),
					wantOutLines: []string{"line1", "1", "line2", "2"},
					wantErrLines: []string{"err1", "1", "err2", "2"},
				},
				{
					name: "plain",
					script: execx.NewScript(`cat -
echo err1 1 > /dev/stderr
echo err2 2 > /dev/stderr`, "bash"),
					stdin: bytes.NewBufferString("line1 1\nline2 2\n"),
					// trailing newlines are removed due to consumers
					wantStdout:   bytes.NewBufferString("line1 1\nline2 2"),
					wantStderr:   bytes.NewBufferString("err1 1\nerr2 2"),
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
							opt = append(opt, execx.WithStdoutWriter(&gotStdoutWriter))
						}
						if tc.withStderrWriter {
							opt = append(opt, execx.WithStderrWriter(&gotStderrWriter))
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

					assert.Equal(t, tc.wantOutLines, gotOutLines)
					assert.Equal(t, tc.wantErrLines, gotErrLines)

					if tc.withStdoutWriter {
						assertReader(t, tc.wantStdout, &gotStdoutWriter)
						assertReader(t, bytes.NewBufferString(""), gotResult.Stdout)
					} else {
						assertReader(t, tc.wantStdout, gotResult.Stdout)
					}

					if tc.withStdoutWriter {
						assertReader(t, tc.wantStderr, &gotStderrWriter)
						assertReader(t, bytes.NewBufferString(""), gotResult.Stderr)
					} else {
						assertReader(t, tc.wantStderr, gotResult.Stderr)
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
					r, err := cmd.Run(context.TODO())
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
