package execx

import (
	"bufio"
	"io"
)

// Scanner reads data from Reader and simultaneously writes it to Writer while passing it to the consumer.
type Scanner struct {
	r        io.Reader
	w        io.Writer
	consumer func(Token)
	split    SplitFunc
}

func NewScanner(w io.Writer, r io.Reader, split SplitFunc, consumer func(Token)) *Scanner {
	return &Scanner{
		w:        w,
		r:        r,
		consumer: consumer,
		split:    split,
	}
}

func (s *Scanner) Scan() error {
	r := io.TeeReader(s.r, s.w)
	sc := bufio.NewScanner(r)
	sc.Split(s.split)
	for sc.Scan() {
		s.consumer(token(sc.Bytes()))
	}
	return sc.Err()
}
