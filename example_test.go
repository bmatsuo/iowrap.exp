package iowrap_test

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/bmatsuo/iowrap.exp"
)

// This example shows how to write gzipped data to a file and read it back out.
func Example() {
	// compose a stack of writers ultimately writing to f.
	w := iowrap.NewWriter(nil)
	err := w.Wrap(ioutil.TempFile("", "iowrap-example-"))
	if err != nil {
		log.Panic(err)
	}
	tmpfile := w.W(0).(*os.File).Name()
	defer os.Remove(tmpfile)
	_ = w.Wrap(gzip.NewWriter(w.W(0)), nil)

	// after writing some data, close the file (flushing the gzip stream in the
	// process).
	_, err = fmt.Fprintln(w, "hello iowrap")
	errclose := w.Close()
	if err == nil {
		err = errclose
	}
	if err != nil {
		log.Panic(err)
	}

	// open the file again, but for reading.
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
