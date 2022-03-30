package pkg

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

// Tree OBJECT AND ITS GOT OBJECT IMPLEMENTATION!!!!!
// tree [content size]\0[Entries having references to other trees and blobs]
// [mode] [file/folder name]\0[SHA-1 of referencing blob or tree]
type Tree struct {
	sha     Sha1
	entries []item
	data    []byte
	cache   treeCache
	len     int
}

type treeCache struct {
	subTrees map[string]*Tree
	blobs    map[string]*item
}

//Object is a composite datatype representing any of the three types in the git obects directory: blobs, trees, commits
type item struct {
	mode uint32
	name string
	sha  Sha1
}

// Tree Structure

//     the literal string tree
//     SPACE (i.e., the byte 0x20)
//     ASCII-encoded decimal length of the uncompressed contents

// After a NUL (i.e., the byte 0x00) terminator, the tree contains one or more entries of the form

//     ASCII-encoded octal mode
//     SPACE
//     name
//     NUL
//     SHA1 hash encoded as 20 unsigned bytes
func parseTree(sha string, r io.Reader) (*Tree, error) {

	var db bytes.Buffer //db means data bytes
	rdr := io.TeeReader(r, &db)
	tree := &Tree{
		entries: []item{},
		cache: treeCache{
			subTrees: map[string]*Tree{},
			blobs:    map[string]*item{},
		},
	}

	tree.sha = strToSha(sha)
	b := bufio.NewReader(rdr)
	prefix, err := b.ReadBytes(Sep)
	if err != nil {
		return nil, err
	}

	_type, size := bytes.Split(prefix, []byte(" "))[0], string(bytes.Split(prefix, []byte(" "))[1])
	if string(_type) != "tree" {
		return nil, fmt.Errorf("Err: should be a tree, but isn't")
	}
	//comeback
	_, err = strconv.ParseInt(size, 10, 64)

	for {
		mp, err := b.ReadBytes(Sep)
		if err != nil {
			if errors.Is(err, io.EOF) { //if IOF is encountered here then we're done with the tree
				break
			}
			return nil, err
		}

		splits := bytes.Split(mp, []byte{Space})
		//filemode is 32 bits. but only 16 are useful for us. It is written as a string of octals,
		//that determines how were going to parse it
		modeStr, name := string(splits[0]), string(splits[1][:len(splits[1])-1])
		modeInt, err := strconv.ParseInt(modeStr, 8, 32)
		if err != nil {
			return nil, err
		}
		mode := uint32(modeInt)
		var s Sha1
		_, err = io.ReadFull(b, s[:])
		if err != nil {
			return nil, err
		}

		tItem := item{
			mode: mode,
			name: name,
			sha:  s,
		}
		tree.entries = append(tree.entries, tItem)
	}
	//comeback to this.
	tree.data = db.Bytes()

	return tree, nil
}

//comeback
func fullyParseTree(sha string, r io.Reader) (*Tree, error) {
	t, err := parseTree(sha, r)
	if err != nil {
		return nil, err
	}

	for _, item := range t.entries {
		if modType(item.mode) == blobfile {
			t.cache.blobs[item.name] = &item
		} else {
			path := hex.EncodeToString(item.sha[:])
			f, err := os.OpenFile(filepath.Join("", path[:2], path[2:]), os.O_RDONLY, 0)
			if err != nil {
				return nil, err
			}
			currT, err := fullyParseTree("", f)
			if err != nil {
				return nil, err
			}
			t.cache.subTrees[item.name] = currT
			f.Close()
		}

	}
	return t, nil
}

type fileType uint8

const (
	blobfile fileType = 0b00000100
	treefile fileType = 0b00001000
)

// comeback
func modType(m uint32) fileType {
	mod := m & uint32(0xffff)
	if (mod >> 8) == uint32(blobfile) {
		return blobfile
	} else if (mod >> 8) == uint32(treefile) {
		return treefile
	} else {
		return 0
	}
}

//comeback
// func (t *Tree) getItem(path string) (item, error) {

// }

func (t *Tree) Hash(wkdir string) ([]byte, error) {
	b, err := HashObj(t.Type(), t.data, wkdir)
	if err != nil {
		t.sha = b
		return b[:], nil
	}
	return nil, fmt.Errorf("Could not hash commit obect: %w", err)
}

func (c *Tree) Type() string {
	return "tree"
}

// comeback check error in caller. if error exists, reverse write
func (t *Tree) Encode(wtr io.WriteCloser) error {
	if _, err := io.WriteString(wtr, fmt.Sprintf("tree %d", t.len)); err != nil {
		return err
	}
	if _, err := wtr.Write([]byte{Sep}); err != nil {
		return err
	}

	defer wtr.Close()
	for _, e := range t.entries {
		if _, err := io.WriteString(wtr, fmt.Sprintf("%o %s", e.mode, e.name)); err != nil {
			return err
		}
		b := []byte{Sep}
		b = append(b, e.sha[:]...)
		if _, err := wtr.Write(b); err != nil {
			return err
		}
	}
	return nil
}
