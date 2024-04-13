# execx

Provides `os/exec.Cmd` wrappers.

``` go
import (
	"bytes"
	"context"
	"fmt"
	"strings"

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
	script.Stdin = bytes.NewBufferString("from stdin\n")
	script.Env.Set("line2", "LINE2")

	var (
		outLines []string
		errLines []string
	)

	if _, err := script.Run(
		context.Background(),
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
```
