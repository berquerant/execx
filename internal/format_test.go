package internal_test

import (
	"testing"

	"github.com/berquerant/execx/internal"
	"github.com/stretchr/testify/assert"
)

func TestSprintf(t *testing.T) {
	for _, tc := range []struct {
		title string
		s     string
		a     map[string]string
		want  string
	}{
		{
			title: "replace missing",
			s:     "test %[k] %[v] %[k]",
			a: map[string]string{
				"k": "key",
			},
			want: "test key %[v] key",
		},
		{
			title: "replace all",
			s:     "test %[k] %[v] %[k]",
			a: map[string]string{
				"k": "key",
				"v": "value",
			},
			want: "test key value key",
		},
		{
			title: "replace",
			s:     "test %[k]",
			a: map[string]string{
				"k": "key",
			},
			want: "test key",
		},
		{
			title: "nil map with verb",
			s:     "test %[k]",
			want:  "test %[k]",
		},
		{
			title: "nil map",
			s:     "test",
			want:  "test",
		},
	} {
		t.Run(tc.title, func(t *testing.T) {
			assert.Equal(t, tc.want, internal.Sprintf(tc.s, tc.a))
		})
	}
}
