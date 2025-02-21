package execx_test

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/berquerant/execx"
)

func ExampleScript_Runner_args() {
	script := execx.NewScript(`echo "$*"`, "sh")
	if err := script.Runner(func(cmd *execx.Cmd) error {
		cmd.Args = append(cmd.Args, "hello", "world")
		r, err := cmd.Run(context.TODO())
		if err != nil {
			return err
		}
		out, err := io.ReadAll(r.Stdout)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", out)
		return nil
	}); err != nil {
		panic(err)
	}

	// Output:
	// hello world
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
			execx.WithStdoutConsumer(func(x execx.Token) {
				fmt.Printf("1:%s\n", x)
			}),
			execx.WithStderrConsumer(func(x execx.Token) {
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
