package execx

import (
	"io"
	"sync"
)

// NullBuffer providing no data on Read and perfoming no actions on Write.
type NullBuffer struct{}

func (NullBuffer) Write(_ []byte) (int, error) {
	return 0, nil
}

func (NullBuffer) Read(_ []byte) (int, error) {
	return 0, io.EOF
}

type ConcurrentWriter struct {
	io.Writer
	mutex sync.Mutex
}

func (c *ConcurrentWriter) Write(p []byte) (n int, err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.Writer.Write(p)
}
