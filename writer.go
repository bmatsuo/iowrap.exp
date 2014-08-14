package iowrap

import "io"

// Writer is a stack of io.Writer implementations, each wrapping the underlying
// one until the base which may be a file, network connection, byte buffer,
// etc.
type Writer struct {
	w []io.Writer
}

// NewWriter allocates and returns a Writer with w on the stack.  If w is nil
// the returned Writer has an empty stack and any calls to Read return an error
// without a preceding call to Wrap.
func NewWriter(w io.Writer) *Writer {
	if w == nil {
		return new(Writer)
	}
	return &Writer{[]io.Writer{w}}
}

// Top returns the writer on top of the stack. It is equivalent to w.W(0).
func (w *Writer) Top() io.Writer {
	return w.W(0)
}

// NumW returns the number of writers in the stack.
func (w *Writer) NumW() int {
	return len(w.w)
}

// W returns the writer at index i in the stack. The top of the stack is index
// zero.
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

// Close terminates all writers on the stack from top to bottom.  Writers in
// the stack implementing io.WriteCloser will have their Close method calls.
// Writers implementing WriteFlusher will have Flush called.  Other writers
// will be ignored.
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
// when all writes have completed.
type WriteFlusher interface {
	// Flush writes any remaining buffered data out to an underlying writer.
	// When interal buffers are empty Flush should be a noop.
	Flush() error
}
