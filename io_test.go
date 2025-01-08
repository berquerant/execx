package execx_test

import (
	"io"
	"testing"

	"github.com/berquerant/execx"
	"github.com/stretchr/testify/assert"
)

func TestNullBuffer(t *testing.T) {
	b := new(execx.NullBuffer)
	{
		n, err := b.Write([]byte(`null`))
		assert.Equal(t, 0, n)
		assert.Nil(t, err)
	}
	{
		x, err := io.ReadAll(b)
		assert.Equal(t, 0, len(x))
		assert.Nil(t, err)
	}
	{
		x, err := io.ReadAll(b)
		assert.Equal(t, 0, len(x))
		assert.Nil(t, err)
	}
	{
		n, err := b.Write([]byte(`null`))
		assert.Equal(t, 0, n)
		assert.Nil(t, err)
	}
}
