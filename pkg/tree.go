package pkg

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

//!!!!TREE OBJECT AND ITS GOTOBJECT IMPLEMENTATION!!!!!
type Tree struct {
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

func parseTree(rdr io.Reader, got *Got) (*Tree, error) {
	tree := &Tree{
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
			return nil, fmt.Errorf("While Parsing tree, could not read object:%w", err)
		}
		bRdr := bufio.NewReader(rdr)
		ty, err := bRdr.ReadBytes(Space)
		if err != nil {
			return nil, err
		}
		ty = ty[:len(ty)-1]
		switch string(ty) {
		case "blob":
			tree.blobs = append(tree.blobs, item)
		case "tree":
			tree.subTrees = append(tree.subTrees, item)
		default:
			return nil, fmt.Errorf("Found non-object")
		}
	}
	return tree, nil
}


func (t *Tree) Hash(wkdir string) ([]byte, error) {
	b, err := HashObj(t.Type(), t.data, wkdir)
	if err != nil {
		t.sha = hex.EncodeToString(b)
		return b, nil
	}
	return nil, fmt.Errorf("Could not hash commit obect: %w", err)
}

func (c *Tree) Type() string {
	return "tree"
}