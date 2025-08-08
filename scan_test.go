package execx_test

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/berquerant/execx"
	"github.com/stretchr/testify/assert"
)

func TestScanner(t *testing.T) {
	for _, tc := range []struct {
		title string
		input string
		split execx.SplitFunc
		want  []string
	}{
		{
			title: "long long string",
			split: bufio.ScanLines,
			input: strings.Repeat("a", bufio.MaxScanTokenSize+1),
			want:  []string{strings.Repeat("a", bufio.MaxScanTokenSize+1)},
		},
		{
			title: "null input",
			split: bufio.ScanLines,
		},
		{
			title: "a line",
			split: bufio.ScanLines,
			input: `line`,
			want:  []string{"line"},
		},
		{
			title: "2 lines",
			split: bufio.ScanLines,
			input: `line1
line2`,
			want: []string{"line1", "line2"},
		},
		{
			title: "scan words",
			split: bufio.ScanWords,
			input: `line1 end
line2 end`,
			want: []string{"line1", "end", "line2", "end"},
		},
	} {
		t.Run(tc.title, func(t *testing.T) {
			var (
				r   = bytes.NewBufferString(tc.input)
				w   bytes.Buffer
				got []string
			)
			s := execx.NewScanner(&w, r, tc.split, func(t execx.Token) {
				got = append(got, t.String())
			})
			assert.Nil(t, s.Scan())
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.input, w.String())
		})
	}
}
