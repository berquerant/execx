package execx_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/berquerant/execx"
	"github.com/stretchr/testify/assert"
)

func assertReader(t *testing.T, want, got io.Reader) {
	wantOut, err := io.ReadAll(want)
	assert.Nil(t, err)
	gotOut, err := io.ReadAll(got)
	assert.Nil(t, err)
	assert.Equal(t, wantOut, gotOut)
}

func TestCmd(t *testing.T) {
	based := t.TempDir()

	t.Run("RunWithStdinConsumer", func(t *testing.T) {
		c := execx.New("cat", "-")
		c.Stdin = bytes.NewBufferString("line1\nline2\n")

		var lines []string
		got, err := c.Run(context.TODO(), execx.WithStdoutConsumer(func(x execx.Token) {
			lines = append(lines, x.String())
		}))

		assert.Nil(t, err)
		assert.Equal(t, []string{"cat", "-"}, got.ExpandedArgs)
		assertReader(t, bytes.NewBufferString("line1\nline2\n"), got.Stdout)
		assertReader(t, bytes.NewBufferString(""), got.Stderr)
		assert.Equal(t, []string{"line1", "line2"}, lines)
	})

	t.Run("Run", func(t *testing.T) {
		t.Run("cancel", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.TODO())
			cancel()
			_, err := execx.New("sleep", "1").Run(ctx)
			assert.ErrorIs(t, err, context.Canceled)
		})

		t.Run("append", func(t *testing.T) {
			os.Setenv("test_cmd_append_env1", "append1")
			cmd := execx.New("echo", "${test_cmd_append_env1}")
			cmd.Env.Merge(execx.EnvFromEnviron())
			cmd.Env.Set("test_cmd_append_env1", "added:${test_cmd_append_env1}")
			r, err := cmd.Run(context.TODO())
			assert.Nil(t, err)
			assertReader(t, bytes.NewBufferString("added:append1\n"), r.Stdout)
		})

		for _, tc := range []struct {
			name string
			cmd  *execx.Cmd
			want *execx.Result
			err  bool
		}{
			{
				name: "not executable",
				cmd:  execx.New(based + "/unknown_cmd"),
				err:  true,
			},
			{
				name: "cat stdin",
				cmd: func() *execx.Cmd {
					c := execx.New("cat", "-")
					c.Stdin = bytes.NewBufferString("from stdin")
					return c
				}(),
				want: &execx.Result{
					ExpandedArgs: []string{"cat", "-"},
					Stdout:       bytes.NewBufferString("from stdin"),
					Stderr:       bytes.NewBufferString(""),
				},
			},
			{
				name: "echo env",
				cmd: func() *execx.Cmd {
					c := execx.New("echo", "i${nternationalizatio}n")
					c.Env.Set("nternationalizatio", "18")
					return c
				}(),
				want: &execx.Result{
					ExpandedArgs: []string{"echo", "i18n"},
					Stdout:       bytes.NewBufferString("i18n\n"),
					Stderr:       bytes.NewBufferString(""),
				},
			},
			{
				name: "echo",
				cmd:  execx.New("echo", "me"),
				want: &execx.Result{
					ExpandedArgs: []string{"echo", "me"},
					Stdout:       bytes.NewBufferString("me\n"),
					Stderr:       bytes.NewBufferString(""),
				},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				got, err := tc.cmd.Run(context.TODO())
				if tc.err {
					t.Logf("err=%v", err)
					assert.NotNil(t, err)
					return
				}
				assert.Nil(t, err)
				assert.Equal(t, tc.want.ExpandedArgs, got.ExpandedArgs)
				assertReader(t, tc.want.Stdout, got.Stdout)
				assertReader(t, tc.want.Stderr, got.Stderr)
			})
		}
	})
}

func ExampleCmd_Run() {
	cmd := execx.New("echo", "Hello, ${NAME}!")
	cmd.Env.Set("NAME", "world")
	r, err := cmd.Run(context.TODO())
	if err != nil {
		panic(err)
	}
	fmt.Println(strings.Join(r.ExpandedArgs, " "))
	b, err := io.ReadAll(r.Stdout)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(b))

	// Output:
	// echo Hello, world!
	// Hello, world!
}
