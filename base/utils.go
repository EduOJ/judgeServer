package base

import (
	"bufio"
	"bytes"
	"github.com/pkg/errors"
	"io"
	"time"
)

// run with timeout when the timeout duration isn't zero
// run without timeout when the timeout duration is zero
func WithTimeout(timeout time.Duration, f func() error) error {
	if timeout == 0 {
		return f()
	}
	done := make(chan bool, 1)
	var err error
	go func() {
		err = f()
		done <- true
	}()
	select {
	case <-done:
		return err
	case <-time.After(timeout):
		return errors.New("timeout")
	}
}

type StrippedReader struct {
	Inner *bufio.Reader
	buf   bytes.Buffer
}

func (r *StrippedReader) Read(p []byte) (int, error) {
	for r.buf.Len() < len(p) {
		ch, _, err := r.Inner.ReadRune()
		if err == io.EOF {
			break
		} else if err != nil {
			return 0, err
		}
		if ch != ' ' && ch != '\n' {
			r.buf.WriteRune(ch)
		}
	}
	return r.buf.Read(p)
}
