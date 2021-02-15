package base

import (
	"github.com/pkg/errors"
	"time"
)

// run with timeout when the timeout duration isn't zero
// run without timeout when the timeout duration is zero
func WithTimeout(timeout time.Duration, f func() error) error {
	if timeout.Nanoseconds() == 0 {
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
