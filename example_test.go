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
	f, err := ioutil.TempFile("", "iowrap-example-")
	if err != nil {
		log.Panic(err)
	}
	defer os.Remove(f.Name())

	// compose a stack of writers using WrapFunc.
	w := iowrap.NewWriter(f)
	_ = w.Wrap(bufio.NewWriter(w.W(0)), nil)
	_ = w.Wrap(gzip.NewWriter(w.W(0)), nil)

	// data will first be gzipped and then buffered before being written to the
	// file.
	_, err = fmt.Fprintln(w, "hello iowrap")
	if err != nil {
		log.Panic(err)
	}

	// close/flush the gzip writer, write bufferred data to the file, and close
	// the file.
	err = w.Close()
	if err != nil {
		log.Panic(err)
	}

	// open the file for reading
	r := iowrap.NewReader(nil)
	err = r.Wrap(os.Open(f.Name()))
	if err != nil {
		log.Panic(err)
	}
	err = r.Wrap(gzip.NewReader(r.R(0)))
	if err != nil {
		log.Panic(err)
	}
	defer r.Close()

	_, err = io.Copy(os.Stdout, r)
	if err != nil {
		log.Panic(err)
	}

	// Output: hello iowrap
}
