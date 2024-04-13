package execx_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/berquerant/execx"
	"github.com/stretchr/testify/assert"
)

func TestScript(t *testing.T) {
	t.Run("Run", func(t *testing.T) {
		const (
			content = `echo line1
echo ${line2}
cat -
echo line3 > /dev/stderr`
			wantStdout = `line1
LINE2
from stdin
`
			wantStderr = `line3
`
		)
		script := execx.NewScript(content, "sh")
		script.Stdin = bytes.NewBufferString("from stdin\n")
		script.Env.Set("line2", "LINE2")

		var (
			outLines []string
			errLines []string
		)

		got, err := script.Run(context.TODO(), execx.WithStdoutConsumer(func(x string) {
			outLines = append(outLines, x)
		}), execx.WithStderrConsumer(func(x string) {
			errLines = append(errLines, x)
		}))
		assert.Nil(t, err)

		assertReader(t, bytes.NewBufferString(wantStdout), got.Stdout)
		assertReader(t, bytes.NewBufferString(wantStderr), got.Stderr)
		assert.Equal(t, strings.Split(strings.TrimSpace(wantStdout), "\n"), outLines)
		assert.Equal(t, strings.Split(strings.TrimSpace(wantStderr), "\n"), errLines)
	})
}

func ExampleScript_Run() {
	script := execx.NewScript(
		`echo line1
echo ${line2}
cat -
echo line3 > /dev/stderr`,
		"sh",
	)
	script.Stdin = bytes.NewBufferString("from stdin\n")
	script.Env.Set("line2", "LINE2")

	var (
		outLines []string
		errLines []string
	)

	if _, err := script.Run(
		context.TODO(),
		execx.WithStdoutConsumer(func(x string) {
			outLines = append(outLines, x)
		}), execx.WithStderrConsumer(func(x string) {
			errLines = append(errLines, x)
		})); err != nil {
		panic(err)
	}

	fmt.Println(strings.Join(outLines, "\n"))
	fmt.Println("---")
	fmt.Println(strings.Join(errLines, "\n"))
	// Output:
	// line1
	// LINE2
	// from stdin
	// ---
	// line3
}
