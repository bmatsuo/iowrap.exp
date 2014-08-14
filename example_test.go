package iowrap_test

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/bmatsuo/iowrap.exp"
)

// This example shows how to perform buffered writes of gzipped data to a file.
func Example() {
	// compose a stack of writers ultimately writing to f. Wrap returns an
	// error but it can safely be ignored if explicitly given a nil error
	// value.
	w := iowrap.NewWriter(nil)
	err := w.Wrap(ioutil.TempFile("", "iowrap-example-"))
	if err != nil {
		log.Panic(err)
	}
	tmpfile := w.W(0).(*os.File).Name()
	defer os.Remove(tmpfile)
	_ = w.Wrap(bufio.NewWriter(w.W(0)), nil)
	_ = w.Wrap(gzip.NewWriter(w.W(0)), nil)

	// after writing some data the file Close closes/flushes the gzip writer,
	// writes bufferred data to the file, and finally closes the file itself.
	_, err = fmt.Fprintln(w, "hello iowrap")
	if errclose := w.Close(); err == nil {
		err = errclose
	}
	if err != nil {
		log.Panic(err)
	}

	// open the file again, but for reading. creating the iowrap.Reader with a
	// nil io.Writer can help clean up the code by providing a more logical
	// ordering of declaration.
	r := iowrap.NewReader(nil)
	err = r.Wrap(os.Open(tmpfile))
	if err != nil {
		log.Panic(err)
	}
	err = r.Wrap(gzip.NewReader(r.R(0)))
	if err != nil {
		log.Panic(err)
	}
	defer r.Close()

	// write decoded data to stdout.
	_, err = io.Copy(os.Stdout, r)
	if err != nil {
		log.Panic(err)
	}

	// Output: hello iowrap
}
