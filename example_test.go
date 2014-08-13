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

	// compose a chain of writers using WrapFunc.
	w := iowrap.NewWriter(f)
	w.WrapFunc(func(w io.Writer) io.Writer { return bufio.NewWriter(w) })
	w.WrapFunc(func(w io.Writer) io.Writer { return gzip.NewWriter(w) })

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

	// Output:
}
