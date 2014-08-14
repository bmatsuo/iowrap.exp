package iowrap

import "io"

// Reader represents a stack of io.Reader implementations, each wrapping the
// underlying one until the base io.Reader which may be a file, network
// connection, byte buffer, etc.
type Reader struct {
	r []io.Reader
}

// NewReader allocates and returns a Reader with r on the stack.  If r is nil
// the returned Reader has an empty stack and any calls to Read return an error
// without a preceding call to Wrap.
func NewReader(r io.Reader) *Reader {
	if r == nil {
		return new(Reader)
	}
	return &Reader{
		r: []io.Reader{r},
	}
}

// NumR returns the number of readers in the stack.
func (r *Reader) NumR() int {
	return len(r.r)
}

// R returns the reader at index i in the stack. The top if the stack is index
// zero.
func (r *Reader) R(i int) io.Reader {
	return r.r[len(r.r)-1-i]
}

// Wrap pushes wrapper onto the stack of writers.  Wrap itself should read
// bytes from the reader previously at the top of the stack but this behavior
// is not enforced.
func (r *Reader) Wrap(wrapper io.Reader, err error) error {
	if err != nil {
		return err
	}
	r.r = append(r.r, wrapper)
	return nil
}

// Read bytes from the top of the stack.
func (r *Reader) Read(bs []byte) (int, error) {
	return r.R(0).Read(bs)
}

// Close all readers in the stack from top to bottom.  Any reader that does not
// implement io.ReadCloser is ignored.  The first error encountered is
// returned.  All readers in the stack are closed regardless of the errors.
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

func termReader(w io.Reader) func() error {
	switch w := w.(type) {
	case io.Closer:
		return w.Close
	default:
		return nil
	}
}
