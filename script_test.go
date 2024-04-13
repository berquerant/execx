package execx_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/berquerant/execx"
	"github.com/stretchr/testify/assert"
)

func TestScript(t *testing.T) {
	t.Run("Run", func(t *testing.T) {
		os.Setenv("test_script_append_env1", "append1")
		const (
			content = `echo line1
echo ${line2}
echo ${test_script_append_env1}
cat -
echo line3 > /dev/stderr`
			stdin      = "from stdin\n"
			wantStdout = `line1
LINE2
added:append1
from stdin
`
			wantStderr = `line3
`
		)
		script := execx.NewScript(content, "sh")
		script.Env.Merge(execx.EnvFromEnviron())
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
}

func ExampleScript_Runner() {
	script := execx.NewScript(
		`echo line1
echo ${line2}
cat -
echo line3 > /dev/stderr`,
		"sh",
	)
	script.Env.Set("line2", "LINE2")

	if err := script.Runner(func(cmd *execx.Cmd) error {
		cmd.Stdin = bytes.NewBufferString("from stdin\n")
		_, err := cmd.Run(
			context.TODO(),
			execx.WithStdoutConsumer(func(x string) {
				fmt.Printf("1:%s\n", x)
			}),
			execx.WithStderrConsumer(func(x string) {
				fmt.Printf("2:%s\n", x)
			}),
		)
		return err
	}); err != nil {
		panic(err)
	}

	// Output:
	// 1:line1
	// 1:LINE2
	// 1:from stdin
	// 2:line3
}
