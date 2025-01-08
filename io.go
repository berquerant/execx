package execx

import "io"

// NullBuffer providing no data on Read and perfoming no actions on Write.
type NullBuffer struct{}

func (NullBuffer) Write(_ []byte) (int, error) {
	return 0, nil
}

func (NullBuffer) Read(_ []byte) (int, error) {
	return 0, io.EOF
}
