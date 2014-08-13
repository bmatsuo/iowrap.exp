package iowrap

import (
	"fmt"
	"io"
)

// Writer represents a stack of io.Writer implementations, each wrapping the
// underlying one until the base which may be a file, network connection, byte
// buffer, etc.
type Writer struct {
	err error
	w   io.Writer
	ws  []io.Writer
}

// NewWriter allocates and returns a Writer.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w: w,
	}
}

// WrapFunc is like Wrap but sets the underlying writer to the result of fn.
func (w *Writer) WrapFunc(fn func(io.Writer) io.Writer) {
	w.Wrap(fn(w.w), nil)
}

// WrapFuncErr is like WrapFunc but allows for fn to return an error fn.
func (w *Writer) WrapFuncErr(fn func(io.Writer) (io.Writer, error)) error {
	return w.Wrap(fn(w.w))
}

// Wrapper sets wrapper as the underlying io.Writer written to by w.
func (w *Writer) Wrap(wrapper io.Writer, err error) error {
	if err != nil {
		return err
	}
	w.ws = append(w.ws, w.w)
	w.w = wrapper
	return nil
}

// Write passes b to the Write method of the immediately underlying writer and
// returns the result.
func (w *Writer) Write(b []byte) (int, error) {
	return w.w.Write(b)
}

// Close terminates all underlying writers from least to most wrapped.  The
// first error encountered, if any, is returned.  All writers are terminated
// regardless of errors in wrapping writers.
func (w *Writer) Close() error {
	err := termWriter(w.w)()
	for i := len(w.ws); i >= 0; i-- {
		_err := termWriter(w.w)()
		if err == nil && _err != nil {
			err = _err
		}
	}
	return err
}

// Reset resets the chain of writers making wbase the base writer.  If the base
// writer is a ResetWriter its Reset method is called, otherwise it is replaced
// with wbase.  Reset returns any error encountered resetting writers including
// encountering any non-ResetWriter writers in the chain.
func (w *Writer) Reset(wbase io.Writer) error {
	var _w io.Writer = wbase
	for i := range w.ws {
		if rw, ok := w.ws[i].(ResetWriter); ok {
			err := rw.Reset(_w)
			if err != nil {
				return err
			}
			_w = w.ws[i]
			continue
		}
		if i > 0 {
			return fmt.Errorf("not a ResetWriter")
		}
	}
	if rw, ok := w.w.(ResetWriter); ok {
		err := rw.Reset(_w)
		if err != nil {
			return err
		}
		return nil
	}
	if len(w.ws) > 0 {
		return fmt.Errorf("not a ResetWriter")
	}
	return nil
}

func termWriter(w io.Writer) func() error {
	switch w := w.(type) {
	case io.Closer:
		return w.Close
	case Flusher:
		return w.Flush
	default:
		return nil
	}
}

type Flusher interface {
	Flush() error
}

type ResetWriter interface {
	io.Writer
	Reset(io.Writer) error
}
