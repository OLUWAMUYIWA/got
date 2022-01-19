package pkg

import (
	"bufio"
	"bytes"
	"encoding/binary"
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
)

type ObjectErr struct {
	ErrSTring string
}

func (e *ObjectErr) Error() string {
	return e.ErrSTring
}

type Hasher interface {
	Hash(wkdir string) ([]byte, error)
}

type GotObject interface {
	Hasher
	Type() string
}

// func parseObject() GotObject {

// }

func parseTree(rdr io.Reader, got *Got) (treeObj, error) {
	tree := treeObj{
		subTrees: []treeItem{},
		blobs:    []treeItem{},
	}
	item := treeItem{}
	all := []treeItem{}
	b := bufio.NewReader(rdr)
	for {
		//s := fmt.Sprintf("%o %s%v%v", mode, ind.path, Sep, ind.sha1_obj_id)
		mp, err := b.ReadBytes(Sep)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			//handle error
		}

		splits := bytes.Split(mp, []byte{Space})
		item.mode, item.path = uint64(binary.BigEndian.Uint32(splits[0])), string(splits[1][:len(splits[1])-1])
		sha := make([]byte, 20)
		n, err := io.ReadFull(b, sha)
		if n != 20 {
			return tree, fmt.Errorf("Last byte stream not up to 20 bytes expected")
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			//handle error
		}
		item.sha1 = hex.EncodeToString(sha)
		all = append(all, item)
	}

	//Now we categorize them based on which kind of object they are and put them inside the tree

	for _, item := range all {
		rdr, err := got.OpenRead(item.sha1)
		if err != nil {
			return treeObj{}, fmt.Errorf("While Parsing tree, could not read object:%w", err)
		}
		bRdr := bufio.NewReader(rdr)
		ty, err := bRdr.ReadBytes(Space)
		if err != nil {
			return treeObj{}, err
		}
		ty = ty[:len(ty)-1]
		switch string(ty) {
		case "blob":
			tree.blobs = append(tree.blobs, item)
		case "tree":
			tree.subTrees = append(tree.subTrees, item)
		default:
			return treeObj{}, fmt.Errorf("Found non-object")
		}
	}
	return tree, nil
}

//!!!!COMMIT OBJECT AND ITS GOTOBJECT IMPLEMENTATION!!!!!
type commitObj struct {
	sha  string
	data []byte
}

func (c *commitObj) Hash(wkdir string) ([]byte, error) {
	b, err := HashObj(c.Type(), c.data, wkdir)
	if err != nil {
		c.sha = hex.EncodeToString(b)
		return b, nil
	}
	return nil, fmt.Errorf("Could not hash commit obect: %w", err)
}

func (c *commitObj) Type() string {
	return "commit"
}

func parseCommit(rdr io.Reader, got *Got) (commitObj, error) {
	//TODO
	_ = bufio.NewReader(rdr)
	return commitObj{}, nil
}

//!!!!TREE OBJECT AND ITS GOTOBJECT IMPLEMENTATION!!!!!
type treeObj struct {
	sha      string
	subTrees []treeItem
	blobs    []treeItem
	data     []byte
}

//Object is a composite datatype representing any of the three types in the git obects directory: blobs, trees, commits
type treeItem struct {
	mode uint64
	path string
	sha1 string
}

func (t *treeObj) Hash(wkdir string) ([]byte, error) {
	b, err := HashObj(t.Type(), t.data, wkdir)
	if err != nil {
		t.sha = hex.EncodeToString(b)
		return b, nil
	}
	return nil, fmt.Errorf("Could not hash commit obect: %w", err)
}

func (c *treeObj) Type() string {
	return "tree"
}

//!!!!BLOB OBJECT AND ITS GOTOBJECT IMPLEMENTATION!!!!!
type blobObj struct {
	sha  string
	data []byte
}

func (blob *blobObj) Hash(wkdir string) ([]byte, error) {
	return HashObj(blob.Type(), blob.data, wkdir)
}

func (c *blobObj) Type() string {
	return "blob"
}

func parseBlob(rdr io.Reader, got *Got) (blobObj, error) {
	//TODO
	return blobObj{}, nil
}

//general hashfunction
func HashObj(ty string, data []byte, base string) ([]byte, error) {
	//use a string builder because it minimizzed memory allocation, which is expensive
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

func (got *Got) Object(sha string) {

}

type gotObject struct {
	ty  ObjectType
	obj GotObject
}
