package execx_test

import (
	"bytes"
	"context"
	"os/exec"
	"testing"

	"github.com/berquerant/execx"
	"github.com/stretchr/testify/assert"
)

func TestPipedCmd(t *testing.T) {
	t.Run("cannot be created without cmds", func(t *testing.T) {
		_, err := execx.NewPipedCmd()
		assert.ErrorIs(t, err, execx.ErrNoCmd)
	})

	for _, tc := range []struct {
		title      string
		cmd        []string
		stdin      string
		wantStdout string
		wantStderr string
		errMsg     string
	}{
		{
			title: "should run 1 cmd",
			cmd:   []string{"grep msg"},
			stdin: `first
msg
last
`,
			wantStdout: `msg
`,
		},
		{
			title: "shoudl run 2 cmds",
			cmd: []string{
				"grep msg",
				"grep hello",
			},
			stdin: `first
msg
msg hello
last
`,
			wantStdout: `msg hello
`,
		},
		{
			title: "should run 1 cmd without stdin",
			cmd: []string{
				`echo first
echo>&2 second`,
			},
			wantStdout: `first
`,
			wantStderr: `second
`,
		},
		{
			title: "should run 2 cmds without stdin",
			cmd: []string{
				`echo first
echo >&2 second
echo last`,
				`echo >&2 third
grep last`,
			},
			wantStdout: `last
`,
			wantStderr: `second
third
`,
		},
		{
			title: "should run 3 cmds",
			cmd: []string{
				`echo first
echo second
echo third
echo last`,
				`cat -`,
				`grep first`,
			},
			wantStdout: `first
`,
		},
		{
			title: "should fail 1 cmd",
			cmd: []string{
				"exit 1",
			},
			errMsg: "failed to wait cmds[0]",
		},
		{
			title: "should fail 2nd cmd",
			cmd: []string{
				"echo first",
				"exit 1",
			},
			errMsg: "failed to wait cmds[1]",
		},
	} {
		t.Run(tc.title, func(t *testing.T) {
			cmds := make([]*exec.Cmd, len(tc.cmd))
			for i, x := range tc.cmd {
				cmds[i] = exec.Command("bash", "-c", x)
			}
			p, err := execx.NewPipedCmd(cmds...)
			if !assert.Nil(t, err) {
				return
			}
			if s := tc.stdin; s != "" {
				p.Stdin = bytes.NewBufferString(s)
			}
			var stdout, stderr bytes.Buffer
			p.Stdout = &stdout
			p.Stderr = &stderr
			if !assert.Nil(t, p.Start(context.TODO())) {
				return
			}
			err = p.Wait()
			if s := tc.errMsg; s != "" {
				assert.ErrorContains(t, err, s)
				return
			}
			if !assert.Nil(t, err) {
				return
			}
			assert.Equal(t, tc.wantStdout, stdout.String())
			assert.Equal(t, tc.wantStderr, stderr.String())
		})
	}
}
