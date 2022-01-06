package internal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

type RefType int

var HeadNotINitErr = fmt.Errorf("Head not yet initialized")

const (
	heads RefType = 0
	remotes
	tags
)

type Ref struct {
	path  string
	_type RefType
}

//initialize a ref
//this initializer does nothing except storing the values
//it expects a full path because it does not have access to the working directory
func InitRef(fullpath string, _type RefType) *Ref {
	return &Ref{fullpath, _type}

}
func RefFromSym(path string, _type RefType) (*Ref, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, HeadNotINitErr
		}
		return nil, fmt.Errorf("Could not open HEAD file: %w\n", err)
	}
	pre := []byte("ref: ")
	if len(b) == 0 {
		return nil, HeadNotINitErr
	}
	if bytes.HasPrefix(b, pre) {
		b = bytes.TrimPrefix(b, pre)
	} else {
		return nil, fmt.Errorf("HEAD file is not well formatted")
	}
	p := string(b)

	if utf8.ValidString(p) {
		return &Ref{p, _type}, nil
	} else {
		return nil, fmt.Errorf("Invalid utf-8 in ref symlink")
	}
}

//contents of refs are commits
func (r *Ref) ReadCont() ([]byte, error) {
	f, err := os.OpenFile(r.path, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("Could not open ref: %s because: %w\n", r.path, err)
	}

	defer f.Close()
	if b, err := io.ReadAll(f); err != nil {
		return nil, fmt.Errorf("Could not read from ref: %s because: %w\n", r.path, err)
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

//comeback
func (r *Ref) Update(path string) {
	pre := "ref: "
	r.path = strings.Join([]string{pre, path}, "")
}
