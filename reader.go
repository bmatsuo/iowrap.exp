package iowrap

import (
	"fmt"
	"io"
)

type Reader struct {
	r []io.Reader
}

func NewReader(r io.Reader) *Reader {
	if r == nil {
		return new(Reader)
	}
	return &Reader{
		r: []io.Reader{r},
	}
}

func (r *Reader) NumR() int {
	return len(r.r)
}

func (r *Reader) R(i int) io.Reader {
	return r.r[len(r.r)-1-i]
}

func (r *Reader) Wrap(wrapper io.Reader, err error) error {
	if err != nil {
		return err
	}
	r.r = append(r.r, wrapper)
	return nil
}

func (r *Reader) Read(bs []byte) (int, error) {
	return r.R(0).Read(bs)
}

func (r *Reader) Close() error {
	var err error
	for i, n := 0, r.NumR(); i < n; i++ {
		var _err error
		term := termReader(r.R(i))
		if term != nil {
			_err = term()
		}
		if err == nil && _err != nil {
			err = _err
		}
	}
	return err
}

// Reset will the call the Reset method on all writters in the stack from the
// bottom upward, making rbase the new base writer.  An error is returned if a
// writer is encountered that does not implement WriteResetter.
func (r *Reader) Reset(rbase io.Reader) error {
	var _r io.Reader = rbase
	for i := range r.r {
		if rw, ok := r.r[i].(ReadResetter); ok {
			err := rw.Reset(_r)
			if err != nil {
				return err
			}
			_r = r.r[i]
			continue
		}
		if i == 0 {
			r.r[i] = rbase
			continue
		}
		return fmt.Errorf("not a ReadResetter")
	}
	return nil
}

func termReader(w io.Reader) func() error {
	switch w := w.(type) {
	case io.Closer:
		return w.Close
	default:
		return nil
	}
}

// ReadResetter allows reuse of an io.Reader middleware by changing the
// underlying reader and clearing any internal state.
type ReadResetter interface {
	io.Reader
	// Reset changes the receiver's underlying io.Reader to r and resets all
	// internal state. After Reset returns an err the receiver's Read method
	// must not be called (unless a subsequent, successful call to Reset is
	// made).  After Reset returns nil the receiver's Read method may be
	// called. The reader's state must be equivalent to that of a newly
	// constructed reader of the same type that proxies writes to w.
	Reset(w io.Reader) error
}
