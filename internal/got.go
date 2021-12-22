package internal

import (
	"log"
	"os"
)

type Got struct {
	baseDir string
	logger  log.Logger
}

const (
	Sep   byte = 0
	Space byte = ' '
)

func NewGot(dir ...string) *Got {
	baseDir := ""
	if len(dir) == 0 {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Could not get the working directory: %s\n", err)
		}
		baseDir = wd
	} else if len(dir) == 1 {
		baseDir = dir[0]
	} else {
		log.Fatalln("Argument should not be more than one")
	}
	logger := log.New(os.Stdout, "GOT: ", log.Default().Flags())
	logger.Printf("Got: ")
	return &Got{baseDir: baseDir, logger: *logger}
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
