package internal

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
)

//convenience function for writing to a file.
func writeToFile(path string, b []byte) {
	var buf bytes.Buffer
	f, err := os.OpenFile(path, os.O_APPEND, 0)
	if err != nil {
		GotErr(err)
	}
	buf.Write(b)
	_, err = buf.WriteTo(f)
	if err != nil {
		GotErr(err)
	}
}

//GotErr is a convenience function for errors that will cause the program to exit
func GotErr(msg interface{}) {
	if msg != nil {
		fmt.Fprintf(os.Stderr, "got err: %v", msg)
		os.Exit(1)
	}
}

func Commit() {

}

//Init creates a directory for your repo and initializes the hidden .git directory
func Init(name string) {
	base, err := os.Getwd()
	if err = os.Mkdir(filepath.Join(base, name), 0); err != nil {
		GotErr(err)
	}
	if err = os.MkdirAll(path.Join(base, name, ".git"), 0); err != nil {
		GotErr(err)
	}
	dirs := []string{"refs", path.Join("refs", "heads"), "objects"}
	for _, dir := range dirs {
		os.MkdirAll(filepath.Join(base, name, dir), 0)
	}
	headPath := filepath.Join(base, name, ".git", "HEAD")
	if _, err = os.Create(headPath); err != nil {
		GotErr(err)
	}
	writeToFile(headPath, []byte("ref: refs/heads/master"))
	log.Printf("Done initializing repository: %s \n", name)
}
