package execx_test

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/berquerant/execx"
)

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

func ExampleCmd_Exec() {
	cmd := execx.New("echo", "Hello, ${NAME}!")
	cmd.Env.Set("NAME", "world")
	if err := cmd.Exec(); err != nil {
		panic(err)
	}
}
