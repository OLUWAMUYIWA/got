package proto

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var sep byte = 0
var nullstr = hex.EncodeToString([]byte{sep})
var zeroId = hex.EncodeToString(make([]byte, 20))
var flushPacket = fmt.Sprintf("%.4x", 0)

//https://github.com/git/git/blob/master/Documentation/technical/protocol-common.txt
// ----
//   pkt-line     =  data-pkt / flush-pkt

//   data-pkt     =  pkt-len pkt-payload
//   pkt-len      =  4*(HEXDIG)
//   pkt-payload  =  (pkt-len - 4)*(OCTET)

//   flush-pkt    = "0000"
// ----

type Pack struct {
	Sha string
}
type PktLine struct {
	len     int16
	id      string
	refname string
	payload []byte
}

//pkt-line stream describing each ref and its current value
type PktStream struct {
	stream []PktLine
}

func DecodePkts(stream []byte) PktStream {
	return PktStream{}
}

func (streamm PktStream) EncodePkts() io.Reader {
	var b bytes.Buffer

	return &b
}

// func createPack(objs []string) []byte {
// 	//var b bytes.Buffer
// 	var b []byte
// 	b = append(b, []byte("PACK")...)
// 	var buf []byte
// 	binary.BigEndian.PutUint32(buf, 2)
// 	b = append(b, buf...)
// 	binary.BigEndian.PutUint32(buf, uint32(len(objs)))
// 	b = append(b, buf...)
// 	sort.Slice(objs, func(i, j int) bool { return strings.Compare(objs[i], objs[j]) == -1 })
// 	for _, obj := range objs {
// 		b = append(b, encodePackObjects(obj)...)
// 	}
// 	sha1 := hash(b)
// 	b = append(b, sha1...)
// 	return b

// }

// func encodePackObjects(sha1 string) []byte {
// 	_, ty, data, err := got.ReadObject(sha1)
// 	if err != nil {
// 		//TODO
// 	}
// 	var b bytes.Buffer
// 	err = compress(&b, data)
// 	if err != nil {
// 		//TODO
// 	}
// 	data_compressed := b.Bytes()

// 	ty_num := 0
// 	switch ty {
// 	case "commit":
// 		ty_num = 1
// 	case "tree":
// 		ty_num = 2
// 	case "blob":
// 		ty_num = 3
// 	default:
// 	}
// 	if ty_num == 0 {
// 		//TODO
// 	}
// 	size := len(data)
// 	by := (ty_num << 4) | (size & 0x0f)
// 	size >>= 4
// 	var ret bytes.Buffer
// 	for i := size; i > 0; i++ {
// 		var b []byte
// 		binary.BigEndian.PutUint64(b, uint64((by | 0x80)))
// 		ret.Write(b)
// 		by = size & 0x7f
// 		size >>= 7
// 	}
// 	var buff []byte
// 	binary.BigEndian.PutUint64(buff, uint64(by))
// 	ret.Write(buff)
// 	ret.Write(data_compressed)
// 	return ret.Bytes()
// }

func verifyPack(p *Pack) error {
	pack, err := os.Open(filepath.Join(".git", "objects", "pack", strings.Join([]string{p.Sha, ".pack"}, "")))
	if err != nil {
		return err
	}
	defer pack.Close()
	idx, err := os.Open(filepath.Join(".git", "objects", "pack", strings.Join([]string{p.Sha, ".idx"}, "")))
	defer idx.Close()
	if err != nil {
		return err
	}
	return nil
}

func hash(data []byte) []byte {
	//create a  new hasher
	hahser := sha1.New()
	//write the bytes to the hasher
	//the hasher write implementation automatically computes the hash written to it
	_, err := hahser.Write(data)
	if err != nil {
		log.Fatalln("could not write data to hasher")
	}
	//to get the hash, we only need to sum a nill byte, which appends a write of its argument(byte slice) to the currently written hash
	//and returns the result
	return hahser.Sum(nil)
}

func compress(writer io.Writer, data []byte) error {
	comp := zlib.NewWriter(writer)
	_, err := comp.Write(data)
	if err != nil {
		return fmt.Errorf("Error whilr compressing: %w", err)
	}
	err = comp.Flush()
	if err != nil {
		return fmt.Errorf("Error whilr compressing: %w", err)
	}
	return nil
}
