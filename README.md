[![Go Report Card](https://goreportcard.com/badge/github.com/berquerant/execx)](https://goreportcard.com/report/github.com/berquerant/execx)
[![Go Reference](https://pkg.go.dev/badge/github.com/berquerant/execx.svg)](https://pkg.go.dev/github.com/berquerant/execx)

# execx

Provides `os/exec.Cmd` wrappers.

``` go
import (
	"bytes"
	"context"
	"fmt"

	"github.com/berquerant/execx"
)

func main() {
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
```
