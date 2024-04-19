package execx_test

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
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

	t.Run("RunWithConsumer", func(t *testing.T) {
		t.Run("Separator", func(t *testing.T) {
			for _, tc := range []struct {
				name       string
				stdin      io.Reader
				opt        []execx.Option
				want       []string
				wantResult io.Reader
			}{
				{
					name:  "lines",
					stdin: bytes.NewBufferString("line1\nline2"),
					opt:   []execx.Option{},
					want: []string{
						"line1",
						"line2",
					},
					wantResult: bytes.NewBufferString("line1\nline2"),
				},
				{
					name:  "csv lines",
					stdin: bytes.NewBufferString("line1\nline2"),
					opt: []execx.Option{
						execx.WithSplitSeparator([]byte(",")),
					},
					want: []string{
						"line1",
						"line2",
					},
					wantResult: bytes.NewBufferString("line1,line2"),
				},
				{
					name:  "csv",
					stdin: bytes.NewBufferString("line1"),
					opt: []execx.Option{
						execx.WithSplitSeparator([]byte(",")),
					},
					want: []string{
						"line1",
					},
					wantResult: bytes.NewBufferString("line1"),
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					c := execx.New("cat", "-")
					c.Stdin = tc.stdin
					var (
						lines []string
						opt   = append(tc.opt, execx.WithStdoutConsumer(func(x execx.Token) {
							lines = append(lines, x.String())
						}))
					)
					got, err := c.Run(context.TODO(), opt...)

					assert.Nil(t, err)
					assert.Equal(t, tc.want, lines)
					assertReader(t, tc.wantResult, got.Stdout)
				})
			}
		})

		for _, tc := range []struct {
			name  string
			stdin io.Reader
			opt   []execx.Option
			want  []string
		}{
			{
				name:  "lines",
				stdin: bytes.NewBufferString("line1\nline2\n"),
				opt:   []execx.Option{},
				want: []string{
					"line1",
					"line2",
				},
			},
			{
				name:  "words",
				stdin: bytes.NewBufferString("word1 word2 "),
				opt: []execx.Option{
					execx.WithSplitFunc(bufio.ScanWords),
				},
				want: []string{
					"word1",
					"word2",
				},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				c := execx.New("cat", "-")
				c.Stdin = tc.stdin
				var (
					lines []string
					opt   = append(tc.opt, execx.WithStdoutConsumer(func(x execx.Token) {
						lines = append(lines, x.String())
					}))
				)
				_, err := c.Run(context.TODO(), opt...)

				assert.Nil(t, err)
				assert.Equal(t, tc.want, lines)
			})
		}
	})

	t.Run("Run", func(t *testing.T) {
		t.Run("earlystop", func(t *testing.T) {
			const (
				input = "1,2,die,4"
				want  = "1,2"
			)
			var (
				wantLines = []string{"1", "2"}
			)

			split := func(data []byte, atEOF bool) (int, []byte, error) {
				i := bytes.IndexByte(data, ',')
				if i == -1 {
					if !atEOF {
						return 0, nil, nil
					}
					return 0, data, bufio.ErrFinalToken
				}

				if string(data[:i]) == "die" {
					return i + 1, nil, bufio.ErrFinalToken
				}

				return i + 1, data[:i], nil
			}

			var outLines []string

			cmd := execx.New("cat", "-")
			cmd.Stdin = bytes.NewBufferString(input)
			got, err := cmd.Run(
				context.TODO(),
				execx.WithSplitFunc(split),
				execx.WithSplitSeparator([]byte(",")),
				execx.WithStdoutConsumer(func(x execx.Token) {
					outLines = append(outLines, x.String())
				}),
			)

			assert.Nil(t, err)
			assert.Equal(t, wantLines, outLines)
			assertReader(t, got.Stdout, bytes.NewBufferString(want))
		})

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
