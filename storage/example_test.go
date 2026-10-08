package storage_test

import (
	"fmt"
	"os"

	"github.com/bnggbn/vax-action-history/storage"
)

func ExampleStore() {
	root, err := os.MkdirTemp("", "vax-storage-example-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(root)
	objects, err := storage.Open(root, 1024*1024)
	if err != nil {
		panic(err)
	}
	text, err := objects.PutText("A place for words.\n")
	if err != nil {
		panic(err)
	}
	blob, err := objects.PutBlob([]byte("opaque bytes"))
	if err != nil {
		panic(err)
	}
	if err = objects.Flush(); err != nil {
		panic(err)
	}
	session := objects.Session(1024*1024, 100)
	restored, err := session.Text(text, 1024)
	if err != nil {
		panic(err)
	}
	data, err := session.Blob(blob, 1024)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s%s\n", restored, data)
	// Output:
	// A place for words.
	// opaque bytes
}
