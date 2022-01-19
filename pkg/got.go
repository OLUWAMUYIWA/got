package pkg

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

type Got struct {
	baseDir string
	head    *Ref
	logger  *log.Logger
}

func (got *Got) WkDir() string {
	return got.baseDir
}

const (
	Sep   byte = 0
	Space byte = ' '
)

//NewGot takes care of initializing our git object
//comeback for refs
func NewGot() *Got {
	//NewGot checks if this is agit directory
	if is, err := IsGit(); !is || err != nil {
		log.Fatalf("This is not a git working directory: %v\n", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Could not get the working directory: %s\n", err)
	}
	baseDir:= wd
	
	//we need the head here
	head, err := RefFromSym(filepath.Join(baseDir, "HEAD"), 0)
	if err != nil {
		if errors.Is(err, NotDefinedErr) {
			log.Fatalf("Could not create head ref because this is not a working git directory")
		}
		log.Fatalf("Error while reading the HEAD file: %s\n", err)
	}

	logger := log.New(os.Stdout, "GOT library: ", log.Ldate | log.Ltime)
	logger.Println("Got Library: ")
	return &Got{baseDir: baseDir, logger: logger, head: head}
}

//All Git objects are stored the same way, just with different types – instead of the string blob, the
//header will begin with commit or tree.
//Git first constructs a header which starts by identifying the type of object
//To that first part of the header, Git adds a space followed by the size in bytes of the content, and adding a final null byte
//Git concatenates the header and the original content and then calculates the SHA-1 checksum of that new content.
// const (
// 	blob ObjectType = 1 << iota
// 	tree
// 	commit
// )

func (got *Got) writeObject(ty ObjectType) {

}

func (got *Got) GotErr(err error) {
	//comeback
	//todo
}

//great, now I can remove every damn wkdir, I think.
//TsGit checks if this is a working git directory. simply: is there a ".git" directory inside iu
func IsGit() (bool, error) {
	is := false
	err := fs.WalkDir(os.DirFS("."), ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && d.Name() == ".git" {
			is = true
			return errors.New("done")
		}
		return nil
	})
	if errors.Is(err, errors.New("'done'")) {
		err = nil
	}
	return is, err
}

//todo
func Wkdir() string {
	return ""
}
