package execx

import (
	"bufio"
	"errors"
	"io"
)

// Scanner reads data from Reader and simultaneously writes it to Writer while passing it to the consumer.
type Scanner struct {
	r        io.Reader
	w        io.Writer
	consumer func(Token)
	delim    byte
}

func NewScanner(w io.Writer, r io.Reader, delim byte, consumer func(Token)) *Scanner {
	if w == nil {
		w = &NullBuffer{}
	}
	return &Scanner{
		w:        w,
		r:        r,
		consumer: consumer,
		delim:    delim,
	}
}

func (s *Scanner) consume(buf []byte) {
	bufLen := len(buf)
	if bufLen == 0 {
		return
	}
	if buf[bufLen-1] == s.delim {
		buf = buf[:bufLen-1]
	}
	s.consumer(token(buf))
}

func (s *Scanner) Scan() error {
	r := bufio.NewReader(io.TeeReader(s.r, s.w))
	for {
		buf, err := r.ReadBytes(s.delim)
		if err == nil {
			s.consume(buf)
			continue
		}
		if errors.Is(err, io.EOF) {
			s.consume(buf)
			return nil
		}
		return err
	}
}
