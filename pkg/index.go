package pkg

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

//The index stores all the info about files needed to write a tree object
//we want to ensure that the index is a multiple of 8 bytes, so we might pad with null bytes if need be
type Index struct {
	ctime_s     [4]byte
	ctime_ns    [4]byte
	mtime_s     [4]byte
	mtime_ns    [4]byte
	dev         [4]byte
	ino         [4]byte
	mode        [4]byte
	uid         [4]byte
	gid         [4]byte
	f_size      [4]byte
	sha1_obj_id [20]byte
	flags       [2]byte
	//ver         [2]byte
	path []byte
}

type Sha1 = [20]byte

func (got *Got) newIndex(path string) *Index {
	f, err := os.Open(string(path))
	if err != nil {
		got.GotErr(err)
	}
	blob, err := io.ReadAll(bufio.NewReader(f))
	sha1 := justhash(blob)
	var stat unix.Stat_t
	err = unix.Stat(path, &stat)
	if err != nil {
		got.GotErr(err)
	}
	i := mapStatToIndex(&stat, sha1, path)
	i.flags = setUpFlags(path)
	i.path = []byte(path)
	return i
}

func mapStatToIndex(stat *unix.Stat_t, sha1 []byte, path string) *Index {
	var i Index
	i.ctime_s = mapint64ToBytes(stat.Ctim.Sec)
	i.ctime_ns = mapint64ToBytes(stat.Ctim.Nsec)
	i.mtime_s = mapint64ToBytes(stat.Mtim.Sec)
	i.mtime_ns = mapint64ToBytes(stat.Mtim.Nsec)
	//at first I worried that the conversions might alter the values of the bits.
	//turns out that isn't true. Nothing happens to the bit, except the new integer size is smaller, causing a loss
	//the difference is in the interpretation of the values by the compiler, since signed uses 2's compliment to evaluate
	//while unsigned just translates the bits
	i.dev = mapint64ToBytes(int64(stat.Dev))
	i.ino = mapint64ToBytes(int64(stat.Ino))
	i.mode = mapint64ToBytes(int64(stat.Mode))
	i.uid = mapint64ToBytes(int64(stat.Uid))
	i.gid = mapint64ToBytes(int64(stat.Gid))
	i.f_size = mapint64ToBytes(int64(stat.Size))
	i.sha1_obj_id = shaToBytes(sha1)
	i.flags = setUpFlags(path)
	return &i
}

//1-bit assume-valid flag (false); 1-bit extended flag (must be zero in version 2); 2-bit stage (during merge);
//12-bit name length if the length is less than 0xFFF, otherwise 0xFFF is stored in this field.
//TODO: check
//The value is 9 in decimal, or 0x9.
func setUpFlags(name string) [2]byte {
	i := int16(0)
	l := len(name)
	if l > 0xFFF {
		i = 0xFFF
	}
	var b []byte
	binary.BigEndian.PutUint16(b, uint16(i))
	var ret [2]byte
	ret[0] = b[0]
	ret[1] = b[1]
	return ret
}

//TODO: should prolly use little-endian since that is what intel porocessors use
//we use bigendian because it is network-endian
func mapint64ToBytes(t int64) [4]byte {
	//right shift the bits by 32 or maybe not, since the lower bits will be zeroes
	var arr [4]byte
	arr[0] = byte(t >> (32 - 8))
	arr[1] = byte(t >> (32 - 16))
	arr[2] = byte(t >> (32 - 24))
	arr[3] = byte(t >> (32 - 32)) //or t & 0xff
	return arr
}

//TODO: is there an idiomatic way to do this?
func shaToBytes(hex []byte) [20]byte {
	if len(hex) != 20 {
		//don't even try to use this method id the length of the slic is not exactly 20. Thank you!
		panic(fmt.Errorf("length not equal to 20"))
	}
	var b [20]byte
	for i := range hex {
		b[i] = hex[i]
	}
	return b
}

//futzing around, thinking I might have need for doing this.
//Index is implementing the io.Reader interface
func (i *Index) Read(b []byte) (int, error) {
	res := i.marshall()
	n := copy(b, res)
	return n, nil
}

//marshall an index into bytes
func (i *Index) marshall() []byte {
	//I see no poit catching errors, these ops are not I/O. Only cpu failure here, I think
	var b bytes.Buffer
	b.Write(i.ctime_s[:])
	b.Write(i.ctime_ns[:])
	b.Write(i.mtime_s[:])
	b.Write(i.mtime_ns[:])
	b.Write(i.dev[:])
	b.Write(i.ino[:])
	b.Write(i.mode[:])
	b.Write(i.uid[:])
	b.Write(i.gid[:])
	b.Write(i.f_size[:])
	b.Write(i.sha1_obj_id[:])
	b.Write(i.flags[:])
	//b.Write(i.ver[:])
	pathlen := len(i.path)
	//data b4 path = 62 bytes. Doing this transformation is a way of appendin the path and still getting a multiple of 8
	//this is how we do the padding
	datalen := int(math.Ceil(float64(b.Len()+pathlen+8)/8) * 8)
	fill := datalen - (b.Len() + pathlen)
	b.Write(i.path)
	//Fill it with zero bytes
	space_fill := bytes.Repeat([]byte{Sep}, fill)
	b.Write(space_fill)
	return b.Bytes()
}
func destructureIntoIndex(b []byte) Index {
	var i Index
	start, lim := 0, 4
	//closure crunches the four-byters
	get_next_four := func(b []byte) [4]byte {
		var arr [4]byte
		for i, pos := start, 0; i < lim; i, pos = i+1, pos+1 {
			arr[pos] = b[i]
		}
		start, lim = start+4, lim+4
		return arr
	}
	i.ctime_s = get_next_four(b)
	i.ctime_ns = get_next_four(b)
	i.mtime_s = get_next_four(b)
	i.mtime_ns = get_next_four(b)
	i.dev = get_next_four(b)
	i.ino = get_next_four(b)
	i.mode = get_next_four(b)
	i.uid = get_next_four(b)
	i.gid = get_next_four(b)
	i.f_size = get_next_four(b)
	//now to the sha-1
	//lim is 4 bytes ahead of start, we want it to be 20 because of sha1
	lim += 16
	get_next_twenty := func(b []byte) [20]byte {
		var arr [20]byte
		for i, pos := start, 0; i < lim; i, pos = i+1, pos+1 {
			arr[pos] = b[i]
		}
		//this op levels the gap. both start and op are at the same position now
		start, lim = start+20, lim+16
		return arr
	}
	i.sha1_obj_id = get_next_twenty(b)
	//now start and lim are th the same number, but incr lim by two
	lim += 2
	get_next_two := func(b []byte) [2]byte {
		var arr [2]byte
		for i, pos := start, 0; i < lim; i, pos = i+1, pos+1 {
			arr[pos] = b[i]
		}
		start = start + 2
		return arr
	}
	i.flags = get_next_two(b)
	i.path = b[lim:]

	return i
}
func unmarshal(data []byte) []Index {
	//length of deterministic bytes = 64
	var indexEntries []Index
	//here, I chose to use a bufio Scanner, because it makes reading the bytes easier. I could just set a custom scanner split func
	bufData := bytes.NewReader(data)
	scanner := bufio.NewScanner(bufData)
	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		//the pre-path length is 64 bytes
		det := data[0:62]
		//put that in first
		token = append(token, det...)
		//now we deal with the path
		//remeber we filled the remaining bytes with zero bytes just after the path
		//and remember that paths are string files, text, more precicely. They could never have zero bytes
		//so it is safe to assume that the first instance of byte(0) signifies the end of the path
		i := bytes.IndexByte(data[62:], Sep)
		if i >= 0 {
			//clean
			path := data[64:i]
			token = append(token, path...)
		}
		//TODO: what if I use: len(token) + 7 /8
		advance = int(math.Ceil(float64(len(token)+8)/8) * 8)
		err = nil
		return
	}
	scanner.Split(split)
	for scanner.Scan() {
		entry := scanner.Bytes()
		indexEntries = append(indexEntries, destructureIntoIndex(entry))
	}
	return indexEntries
}

// read: https://mincong.io/2018/04/28/git-index/
//The index file contains:
// 12-byte header.
// A number of sorted index entries.
// Extensions. They are identified by signature.
// 160-bit SHA-1 over the content of the index file before this checksum.

func readIndexFile(got *Got) ([]Index, error) {
	if is, _ := IsGit(); !is {
		got.logger.Fatalf("Not a valid git directory\n")
	}
	f, err := os.Open(filepath.Join(".git/index"))
	p_err, ok := err.(*os.PathError)
	if ok {
		temp_err := errors.New("no such file or directory")
		if p_err.Unwrap() == temp_err {
			got.logger.Fatalf("You have not indexed any file\n")
		}
	} else {
		got.UpErr(err)
		return nil, err
	}
	data, err := io.ReadAll(f)
	got.GotErr(err)
	hash := justhash(data[:len(data)-20])
	//the index file has the lst 160 bits (i.e. 20 bytes) as the sha-1 checksum of all the bits tat come before it
	//we need to ensure that it matches before considering the data valid
	if bytes.Compare(hash, data[:(len(data)-20)]) != 0 {
		got.GotErr(errors.New("Checksum is not equal to file digest. File has been tampered with"))
	}
	hdr := data[:12]
	sign := string(hdr[:4])
	version := binary.BigEndian.Uint32(hdr[4:8])
	numEntries := binary.BigEndian.Uint32(hdr[8:])
	//we need to check what the header says.
	if sign != "DIRC" {
		got.GotErr(fmt.Errorf("bad index file sha1 signature: %s", sign))
	}
	if version != 2 {
		got.GotErr(fmt.Errorf("Version number must be 2, got %d", version))
	}
	//now for the index entries :
	//we need to use the unix fstat
	//the index files are listed between the 12-byte header and the 20-byte checksum
	indEntries := data[12:(len(data) - 20)]
	indexes := unmarshal(indEntries)
	if len(indexes) != int(numEntries) {
		got.GotErr(fmt.Errorf("Number of enteries does not equal to what the head specified"))
	}
	return indexes, nil
}

//write the index file, given a slice of index
//this is the function that stages files
//Index file integers in git are written in NE.
func (got *Got) UpdateIndex(entries []*Index) error {
	var hdr []byte
	hdr = append(hdr, []byte("DIRC")...)
	//buf is apparently reusable
	var buf []byte
	binary.BigEndian.PutUint32(buf, 2)
	hdr = append(hdr, buf[:4]...)
	//use the same buffer, since the buffer does not keep its state. It starts over
	binary.BigEndian.PutUint32(buf, uint32(len(entries)))
	hdr = append(hdr, buf[:4]...)
	var data []byte
	for _, entry := range entries {
		data = append(data, entry.marshall()...)
	}
	allData := bytes.Join([][]byte{hdr, data}, nil)
	checksum := justhash(allData)
	index := bytes.Join([][]byte{allData, checksum}, nil)
	err := writeToFile(filepath.Join(".git", "index"), index)
	if err != nil {
		return err
	}
	return nil
}
