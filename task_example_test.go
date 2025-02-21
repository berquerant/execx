package execx_test

import (
	"context"
	"fmt"

	"github.com/berquerant/execx"
)

func ExampleExecutableTasks() {
	err := execx.NewExecutableTasks(
		execx.NewTasks().
			Add(execx.NewTask("f", `echo "f($*)"`)).
			Add(execx.NewTask("g", `r="$(f "$@")"
echo "g(${r})"`)).
			Add(execx.NewTask("h", `g "A" "$A"`)),
		execx.EnvFromSlice([]string{"A=envA"}),
		"h",
	).IntoScript("sh").Runner(func(cmd *execx.Cmd) error {
		_, err := cmd.Run(context.TODO(), execx.WithStdoutConsumer(func(x execx.Token) {
			fmt.Println(x.String())
		}))
		return err
	})
	if err != nil {
		panic(err)
	}

	// Output:
	// g(f(A envA))
}
