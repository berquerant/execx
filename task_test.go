package execx_test

import (
	"context"
	"io"
	"testing"

	"github.com/berquerant/execx"
	"github.com/stretchr/testify/assert"
)

func TestExecutableTasks(t *testing.T) {
	for _, tc := range []struct {
		title string
		tasks *execx.ExecutableTasks
		want  string
	}{
		{
			title: "empty tasks",
			tasks: execx.NewExecutableTasks(
				execx.NewTasks(),
				execx.NewEnv(),
				"echo empty",
			),
			want: `empty
`,
		},
		{
			title: "call task",
			tasks: execx.NewExecutableTasks(
				execx.NewTasks().
					Add(execx.NewTask("f", `echo "got $1"`)).
					Add(execx.NewTask("g", `f g`)),
				execx.NewEnv(),
				"g",
			),
			want: `got g
`,
		},
		{
			title: "call tasks",
			tasks: execx.NewExecutableTasks(
				execx.NewTasks().
					Add(execx.NewTask("f", `echo "got $1"`)).
					Add(execx.NewTask("g", `f g`)),
				execx.NewEnv(),
				"g",
				"f x",
			),
			want: `got g
got x
`,
		},
		{
			title: "call task with env",
			tasks: execx.NewExecutableTasks(
				execx.NewTasks().
					Add(execx.NewTask("f", `echo "got $*"`)).
					Add(execx.NewTask("g", `f g "$A"`)),
				execx.EnvFromSlice([]string{
					"A=ea",
				}),
				"g",
			),
			want: `got g ea
`,
		},
	} {
		t.Run(tc.title, func(t *testing.T) {
			t.Run("into script", func(t *testing.T) {
				var got string
				assert.Nil(t, tc.tasks.IntoScript("sh").Runner(func(cmd *execx.Cmd) error {
					r, err := cmd.Run(context.TODO())
					if err != nil {
						return err
					}
					out, err := io.ReadAll(r.Stdout)
					if err != nil {
						return err
					}
					got = string(out)
					return nil
				}))
				assert.Equal(t, tc.want, got)
			})
			t.Run("from string", func(t *testing.T) {
				var got string
				execx.NewScript(tc.tasks.String(), "sh").Runner(func(cmd *execx.Cmd) error {
					r, err := cmd.Run(context.TODO())
					if err != nil {
						return err
					}
					out, err := io.ReadAll(r.Stdout)
					if err != nil {
						return err
					}
					got = string(out)
					return nil
				})
				assert.Equal(t, tc.want, got)
			})
		})
	}
}

func TestTasks(t *testing.T) {
	for _, tc := range []struct {
		title string
		tasks execx.Tasks
		want  string
	}{
		{
			title: "empty",
			tasks: execx.NewTasks(),
			want:  "",
		},
		{
			title: "1 function",
			tasks: execx.NewTasks().
				Add(execx.NewTask("f", `echo from f`)),
			want: `f() {
  echo from f
}`,
		},
		{
			title: "2 functions",
			tasks: execx.NewTasks().
				Add(execx.NewTask("f", `echo from f`)).
				Add(execx.NewTask("g", `echo from g
echo end`)),
			want: `f() {
  echo from f
}
g() {
  echo from g
  echo end
}`,
		},
	} {
		t.Run(tc.title, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.tasks.String())
		})
	}
}
