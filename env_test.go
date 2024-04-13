package execx_test

import (
	"sort"
	"testing"

	"github.com/berquerant/execx"

	"github.com/stretchr/testify/assert"
)

func TestEnv(t *testing.T) {
	t.Run("slice", func(t *testing.T) {
		for _, x := range [][]string{
			{},
			{"KEY1=VALUE1"},
			{"KEY1=VALUE1", "KEY2=VALUE2"},
		} {
			got := execx.EnvFromSlice(x).IntoSlice()
			sort.Strings(x)
			sort.Strings(got)
			assert.Equal(t, x, got)
		}
	})

	t.Run("set", func(t *testing.T) {
		e := execx.NewEnv()
		{
			_, ok := e.Get("KEY")
			assert.False(t, ok)
		}
		e.Set("KEY", "VAL")
		{
			got, ok := e.Get("KEY")
			assert.True(t, ok)
			assert.Equal(t, "VAL", got)
		}
		assert.Equal(t, execx.EnvFromSlice([]string{"KEY=VAL"}), e)
	})

	t.Run("merge", func(t *testing.T) {
		e := execx.NewEnv()
		e.Merge(execx.NewEnv())
		t.Run("empty", func(t *testing.T) {
			assert.Equal(t, 0, len(e.IntoSlice()))
		})
		t.Run("addkey", func(t *testing.T) {
			e.Merge(execx.EnvFromSlice([]string{"k=v"}))
			assert.Equal(t, execx.EnvFromSlice([]string{"k=v"}), e)
		})
		t.Run("overwrite", func(t *testing.T) {
			e.Merge(execx.EnvFromSlice([]string{"k=vv", "k2=v2"}))
			assert.Equal(t, execx.EnvFromSlice([]string{"k=vv", "k2=v2"}), e)
		})
	})

	t.Run("expand", func(t *testing.T) {
		e := execx.EnvFromSlice([]string{"KEY=VAL"})
		assert.Equal(t, "value is VAL", e.Expand("value is $KEY"))

		e.Set("KEY2", "new${KEY}")
		assert.Equal(t, "value is newVAL", e.Expand("value is $KEY2"))
	})

	t.Run("expand exhausted", func(t *testing.T) {
		e := execx.EnvFromSlice([]string{"A=B", "B=C", "C=A"})
		t.Log(e.Expand("value is $A"))
	})

	t.Run("ignore unknown variables", func(t *testing.T) {
		e := execx.NewEnv()
		assert.Equal(t, "value is ${A}", e.Expand("value is ${A}"))
		assert.Equal(t, "value is ${A}", e.Expand("value is $A"))
	})

	t.Run("expand strings", func(t *testing.T) {
		e := execx.EnvFromSlice([]string{"SUBJECT=Alice", "LOCATION1=Virginia", "LOCATION2=Atlanta"})
		assert.Equal(t, []string{
			"Alice went to Virginia",
			"Alice went to Atlanta",
		}, e.ExpandStrings([]string{
			"$SUBJECT went to $LOCATION1",
			"$SUBJECT went to $LOCATION2",
		}))
	})

	t.Run("append", func(t *testing.T) {
		e := execx.NewEnv()
		e.Set("A", "a")
		e.Set("A", "b$A")
		assert.Equal(t, "ba", e.Expand("$A"))
	})
}
