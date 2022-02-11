package pkg

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ObjectType int

const (
	blob ObjectType = 1 << iota
	tree
	commit
	tag
)

type ObjectErr struct {
	ErrSTring string
	inner error
}

func (e *ObjectErr) Error() string {
	return e.ErrSTring
}

func (e *ObjectErr) Unwrap() error {
	return e.inner
}

func (e *ObjectErr) SetInner(err error) {
	e.inner = err
}

type Hasher interface {
	Hash(wkdir string) ([]byte, error)
}

type GotObject interface {
	Hasher
	Type() string
}


//an object could be a commit, tree, blob, or tag
func parseObject(sha string, got *Got) (*GotObject, error) {
	// first we check the packfiles to see if the object is among the parsed

	//then we check the git object directory
}


//general hashfunction
func HashObj(ty string, data []byte, base string) ([]byte, error) {
	//use a string builder because it minimized memory allocation, which is expensive
	//each write appends to the builder
	//IGNORING errors here, too many writes, error handling will bloat the code.
	var s strings.Builder
	hdr := fmt.Sprintf("%s %d", ty, len(data))
	//i see no reason to handle errors here since no I/O is happening
	//Builder only implements io.Writer.
	s.WriteString(hdr)
	s.WriteByte(Sep)
	s.Write(data)
	b := []byte(s.String())
	raw := justhash(b)
	//the byte result must be converted to hex string as that is how it is useful to us
	//we could either use fmt or hex.EncodeString here. Both works fine
	//TODO: explain how hex.Decode() does its job. Cool stuff. Also explain how EncodeToString() does it.

	hash_str := hex.EncodeToString(raw)
	//first two characters (1 byte) are the name of the directory. The remaining 38 (19 bytes) are the  name of the file
	//that contains the compressed version of the blob.
	//remember that sha1 produces a 20-byte hash (160 bits, or 40 hex characters)
	path := filepath.Join(string(base), ".git/objects/", hash_str[:2])
	err := os.MkdirAll(path, 0777)
	if err != nil {
		return nil, &ObjectErr{}
	}
	fPath := filepath.Join(path, hash_str[2:])
	f, err := os.Create(fPath)
	if err != nil {
		return nil, &ObjectErr{}
	}
	defer f.Close()
	//the actual file is then compressed and stored in the file created
	err = compress(f, b)
	if err != nil {
		return nil, &ObjectErr{}
	}
	return raw, nil
}

///READING and WRITING Objects
func (g *Got) OpenRead(id string) (io.Reader, error) {
	if len(id) < 6 {
		return nil, errors.New("id not long enough. Use > 6")
	}
	f, err := os.OpenFile(filepath.Join(g.baseDir, ".git", "objects", id[:2], id[2:]), os.O_RDONLY, 0)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("File does not exist: %w", err)
		}
		return nil, fmt.Errorf("While opening Object: %w", err)
	}
	defer f.Close()
	if b, err := io.ReadAll(f); err == nil {
		var writeB bytes.Buffer
		if err = decompress(bytes.NewReader(b), &writeB); err != nil {
			return nil, fmt.Errorf("Error while decompressing object: %w", err)
		}
		return &writeB, nil
	} else {
		return nil, fmt.Errorf("Error while reading from Object directory: %w", err)
	}
}

func (g *Got) OpenWrite() error {
	return fmt.Errorf("")
}

func (got *Got) Object(sha string, ty ObjectType) (GotObject, error){
	// try pack

	//try object directory
	// b, err := fs.ReadFile(os.DirFS(filepath.Join(got.WkDir(), ".git", sha[0:2])), shap[2:])
	f, err := os.Open(filepath.Join(got.WkDir(), ".git", sha[0:2], sha[2:]))

	if err != nil {
		return nil, err
	} 
	objRdr, err := zlib.NewReader(f)
	if err != nil {
		return nil, err
	}
	switch ty {
		case blob: {
			return parseBlob(objRdr, got)
		}

		case tree: {
			return parseTree(objRdr, got)
		}
		
		case commit: {
			return parseCommit(objRdr, got)
		}	

		case tag: {
			parseTag(objRdr, got)
		}

		default: {
			return nil, fmt.Errorf("invalid object type")
		}
	}
	f.Close()

	return nil, err
}

type gotObject struct {
	ty  ObjectType
	obj GotObject
}


func (o *gotObject) parse() {
	
}


