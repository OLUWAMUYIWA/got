package internal

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"os"

	"golang.org/x/sys/unix"
)

/// |||GIT struct||| ////
type Git struct{
	logger log.Logger
}


func NewGit() *Git {
	logger := log.New(os.Stdout, "GOT: ", log.Default().Flags())
	logger.Printf("Got: ")
	return &Git{logger: *logger}
}
// func NewGit() *Git {
// 	//TODO: before init, wkdir is the enclosing dir. After init. Think about this later, prolly no point for the git struct
// 	//BUG there has to be a way to update wkdir to the git repo itself if i'm to keep it. And obviously for every cli command,
// 	//I have to init the git struct as clis are stateless.
// 	basedir, err := os.Getwd()
// 	if err != nil {
// 		GotErr(err)
// 	}
// 	return &Git{
// 		wkdir: basedir,
// 	}
// }


//TODO: replace all path strings with this. better to alias, they all look silly as strings
type FPath string


/// Errors that do not belong directly to the Git struct
//Preference is to return these errors early after adding a context.
//the caller then decides what to do. Because this app is not robust, we mostly just panic
type OpErr struct {
	ErrSTring string
}

func (e *OpErr) Error() string {
	return e.ErrSTring
}

var (
	IOWriteErr       = &OpErr{"Could not write to the specified writer: "}
	PermissionErr = &OpErr{"Permission denied: "}
	FormatErr     = &OpErr{"Bad Formatting: "}
	OpenErr = &OpErr{"Could not open file: "}
	CopyErr = &OpErr{"COuld not copy data: "}
	NotDefinedErr = &OpErr{"Value not defined" }
	IOCreateErr = &OpErr{"Could not create file/ directory:"}
	IoReadErr = &OpErr{"Could not read file:"}
	NetworkErr = &OpErr{"Network Error: "}
)

func (e *OpErr)addContext(s string) *OpErr {
	newErr := *e
	newErr.ErrSTring = fmt.Sprintf("%s: %s",newErr.ErrSTring, s)
	return &newErr
}

/// |||The Git struct error handler |||| ///
//TODO: Do better error handling here
//GotErr is a convenience function for errors that will cause the program to exit
func (git *Git)GotErr(msg interface{}) {
	if msg != nil {
		git.logger.Fatalf("got err: %v", msg)
		os.Exit(1)
	}
}

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
	//depends on the name of the file
	path []byte
}

func (git *Git) newIndex(path string) *Index {
	f, err := os.Open(string(path))
	if err != nil {
		git.GotErr(err)
	}
	blob, err := io.ReadAll(bufio.NewReader(f))
	sha1 := git.HashObject(blob, "blob", true)
	var stat unix.Stat_t
	err = unix.Stat(path, &stat)
	if err != nil {
		git.GotErr(err)
	}
	i := mapStatToIndex(&stat, sha1)
	i.flags = setUpFlags(path)
	//TODO: path
	i.path = []byte(path)
	return i
}

//TODO: This does not even make any sense yet
func mapStatToIndex(stat *unix.Stat_t, sha1 []byte) *Index {
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
	//TODO remaining
	return &i
}

//1-bit assume-valid flag (false); 1-bit extended flag (must be zero in version 2); 2-bit stage (during merge);
//12-bit name length if the length is less than 0xFFF, otherwise 0xFFF is stored in this field.
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

//we use bigendian because it is network-endian
func mapint64ToBytes(t int64) [4]byte {
	//right shift the bits by 32 or maybe not, since the lower bits will be zeroes
	var arr [4]byte
	arr[0] = byte(t >> (32 - 8))
	arr[1] = byte(t >> (32 - 16))
	arr[2] = byte(t >> (32 - 24))
	arr[3] = byte(t >> (32 - 32))
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

func (i *Index) marshall() []byte {
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
	//path is not fixed
	pathlen := len(i.path)
	//data b4 path = 64 bytes. Doing this transformation is a way of appendin the path and still getting a multiple of 8
	datalen := int(math.Ceil(float64(b.Len()+pathlen+8)/8) * 8)
	fill := datalen - (b.Len() + pathlen)
	b.Write(i.path)
	//TODO: Not sure this works
	space_fill := bytes.Repeat([]byte{byte(0)}, fill)
	//fill_bytes := make([]byte, fill, fill)
	b.Write(space_fill)
	return b.Bytes()
}
func destructureIndex(b []byte) Index {
	var i Index
	start, lim := 0, 4
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
	lim += 16
	get_next_twenty := func(b []byte) [20]byte {
		var arr [20]byte
		for i, pos := start, 0; i < lim; i, pos = i+1, pos+1 {
			arr[pos] = b[i]
		}
		start, lim = start+20, lim+16
		return arr
	}
	i.sha1_obj_id = get_next_twenty(b)
	//now start and lim are th the same number
	lim += 2
	get_next_two := func(b []byte) [2]byte {
		var arr [2]byte
		for i, pos := start, 0; i < lim; i, pos = i+1, pos+1 {
			arr[pos] = b[i]
		}
		start, lim = start+2, lim+2
		return arr
	}
	i.flags = get_next_two(b)
	//i.ver = get_next_two(b)
	//the remainder is the path
	//TODO: check again
	pos := bytes.IndexByte(b[:lim], byte(0))
	i.path = b[lim:pos]

	return i
}
func unmarshal(data []byte) []Index {
	//length of deterministic bytes = 64
	var indexEntries []Index
	bufData := bytes.NewReader(data)
	scanner := bufio.NewScanner(bufData)
	//TODO: this scan func needs to be revisited
	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		det := data[0:64]
		token = append(token, det...)
		i := bytes.IndexByte(data[64:], byte(0))
		if i >= 0 {
			path := data[64:i]
			token = append(token, path...)
		}
		advance = 64 + i
		err = nil
		return
	}
	scanner.Split(split)
	for scanner.Scan() {
		entry := scanner.Bytes()
		indexEntries = append(indexEntries, destructureIndex(entry))
	}
	return indexEntries
}


//Objeect is a composite datatype representing any of the three types in the git obects directory: blobs, trees, commits 
type Object struct {
	mode int64
	path string
	sha1 string
}


type ConfigObject struct {
	uname string `json: uname`
	email string `json: email`
}