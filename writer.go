package iowrap

import (
	"bufio"
	"fmt"
	"io"
	"log"
)

// Buffer wraps v with bufio.Reader if v is a *Reader or bufio.Writer if v is a
// *Writer.  Buffer issues log.Panic if v's type is neither *Writer nor
// *Reader.
func Buffer(v interface{}) {
	switch x := v.(type) {
	case *Writer:
		// this does not return an error. Wrap just forwards input errors.
		x.Wrap(bufio.NewWriter(x.W(0)), nil)
	case *Reader:
		// this does not return an error. Wrap just forwards input errors.
		x.Wrap(bufio.NewReader(x.R(0)), nil)
	default:
		log.Panicf("buffer: invalid type %T not *iowrap.Writer or *iowrap.Reader", v)
	}
}

// Writer represents a stack of io.Writer implementations, each wrapping the
// underlying one until the base which may be a file, network connection, byte
// buffer, etc.
type Writer struct {
	err error
	w   []io.Writer
}

// NewWriter allocates and returns a Writer using w as the base writer.
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w: []io.Writer{w},
	}
}

func (w *Writer) NumW() int {
	return len(w.w)
}

func (w *Writer) W(i int) io.Writer {
	return w.w[len(w.w)-1-i]
}

// Wrap pushes wrapper onto the stack of writers.
func (w *Writer) Wrap(wrapper io.Writer, err error) error {
	if err != nil {
		return err
	}
	w.w = append(w.w, wrapper)
	return nil
}

// Write writes b to the writer on the top of the stack.
func (w *Writer) Write(b []byte) (int, error) {
	return w.W(0).Write(b)
}

// Close terminates all writers on the stack.  Writers are terminated from top
// of the stack downward.
func (w *Writer) Close() error {
	var err error
	for i, n := 0, w.NumW(); i < n; i++ {
		var _err error
		term := termWriter(w.W(i))
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
// bottom upward, making wbase the new base writer.  An error is returned if a
// writer is encountered that does not implement WriteResetter.
func (w *Writer) Reset(wbase io.Writer) error {
	var _w io.Writer = wbase
	// range over w.w goes from the bottom of the stack upward.
	for i := range w.w {
		if wr, ok := w.w[i].(WriteResetter); ok {
			err := wr.Reset(_w)
			if err != nil {
				return err
			}
			_w = w.w[i]
			continue
		}
		if i == 0 {
			w.w[i] = wbase
			continue
		}
		return fmt.Errorf("not a WriteResetter")
	}
	return nil
}

func termWriter(w io.Writer) func() error {
	switch w := w.(type) {
	case io.Closer:
		return w.Close
	case WriteFlusher:
		return w.Flush
	default:
		return nil
	}
}

// WriteFlusher is an interface for io.Writer middleware that needs to buffer
// data internally before flushing to an underlying writer.  In such situations
// it is often the case that the buffer is only partially full at the point
// where all writes have completed.
type WriteFlusher interface {
	// Flush writes any remaining buffered data out to an underlying writer.
	// When interal buffers Flush should be a noop.
	Flush() error
}

// WriteResetter allows reuse of an io.Writer middleware by changing the
// underlying writer and clearing any internal state.
type WriteResetter interface {
	io.Writer
	// Reset changes the receiver's underlying io.Writer to w and resets all
	// internal state. After Reset returns an err the receiver's Write method
	// must not be called (unless a subsequent, successful call to Reset is
	// made).  After Reset returns nil the receiver's Write method may be
	// called. The writer's state must be equivalent to that of a newly
	// constructed object of the same type that proxies writes to w.
	Reset(w io.Writer) error
}
