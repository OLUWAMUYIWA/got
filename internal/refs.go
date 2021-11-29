package internal

import (
	"fmt"
	"io"
	"os"
)


type Refs struct {
	head string
}

//supply basedir with full path to refs/HEAD
func InitRef(path string) *Refs {
	return &Refs{head: path}
}

func (r *Refs) ReadCont() ([]byte, error) {
	f, err := os.OpenFile(r.head, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("Could not open HEAD file: %w\n", err)
	}

	defer f.Close()
	if b, err := io.ReadAll(f); err != nil {
		return nil, fmt.Errorf("Could not read from HEAD file: %w\n", err)
	} else {
		return b, nil
	}
}

// //sha-1 of the last commit (or shall we say latest?)
// //git alays chech the HEAD file for the last commit.
// func (got *Got) parentSha() string {
// 	path := filepath.Join(".git", "refs", "head", "master")
// 	f, err := os.Open(path)
// 	got.GotErr(err)

// 	//Hahaha. I always feel good anytime I'm able to explit Go's interface semantics
// 	var s strings.Builder
// 	_, err = io.Copy(&s, f)
// 	got.GotErr(err)
// 	return strings.Trim(s.String(), "\n")
// }

func (r *Refs) Update(path string) {
	r.head = path
}