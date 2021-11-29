package internal

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func (got *Got) createPack(objs []string) []byte {
	//var b bytes.Buffer
	var b []byte
	b = append(b, []byte("PACK")...)
	var buf []byte
	binary.BigEndian.PutUint32(buf, 2)
	b = append(b, buf...)
	binary.BigEndian.PutUint32(buf, uint32(len(objs)))
	b = append(b, buf...)
	sort.Slice(objs, func(i, j int) bool { return strings.Compare(objs[i], objs[j]) == -1 })
	for _, obj := range objs {
		b = append(b, got.encodePackObjects(obj)...)
	}
	sha1 := justhash(b)
	b = append(b, sha1...)
	return b

}


func (got *Got) encodePackObjects(sha1 string) []byte {
	_, ty, data, err := got.ReadObject(sha1)
	got.GotErr(err)
	var b bytes.Buffer
	err = compress(&b, data)
	if err != nil {
		got.GotErr(err)
	}
	data_compressed := b.Bytes()

	ty_num := 0
	switch ty {
	case "commit":
		ty_num = 1
	case "tree":
		ty_num = 2
	case "blob":
		ty_num = 3
	default:
	}
	if ty_num == 0 {
		got.GotErr(errors.New("wrong boject type"))
	}
	size := len(data)
	by := (ty_num << 4) | (size & 0x0f)
	size >>= 4
	var ret bytes.Buffer
	for i := size; i > 0; i++ {
		var b []byte
		binary.BigEndian.PutUint64(b, uint64((by | 0x80)))
		ret.Write(b)
		by = size & 0x7f
		size >>= 7
	}
	var buff []byte
	binary.BigEndian.PutUint64(buff, uint64(by))
	ret.Write(buff)
	ret.Write(data_compressed)
	return ret.Bytes()
}


func verifyPack(p *Pack) error {
	if is, err := IsGit(); err != nil || !is {
		return errors.New("not git girectory")
	}
	pack, err := os.Open(filepath.Join(".git", "objects", "pack", strings.Join([]string{p.sha, ".pack"}, "")))
	if err != nil {
		return err
	}
	defer pack.Close()
	idx, err := os.Open(filepath.Join(".git", "objects", "pack",  strings.Join([]string{p.sha, ".idx"}, "")))
	defer idx.Close()
	if err != nil {
		return err
	}
	return nil
}
