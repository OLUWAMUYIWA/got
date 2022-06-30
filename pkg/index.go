package pkg

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/unix"
)

// we work only with version 2
// https://github.com/git/git/blob/master/Documentation/technical/index-format.txt
type Idx struct {
	entries []*IdxEntry
	cache   map[string]*IdxEntry
}

type IdxEntry struct {
	cTime                             time.Time
	mTime                             time.Time
	dev, inode, mode, uid, gid, fsize uint32
	sha                               [20]byte
	flags                             uint16
	path                              []byte
}

// read: https://mincong.io/2018/04/28/git-index/
//The index file contains:
// 12-byte header.
// A number of sorted index entries.
// Extensions. They are identified by signature.
// 160-bit SHA-1 over the content of the index file before this checksum.

func readIndexFile() (*Idx, error) {
	i := &Idx{}
	if is, _ := IsGit(); !is {
		return nil, errors.New("Not a valid git directory\n")
	}
	f, err := os.Open(".git/index")
	p_err, ok := err.(*os.PathError)
	if ok {
		temp_err := errors.New("no such file or directory")
		if p_err.Unwrap() == temp_err {
			return nil, fmt.Errorf("You have not indexed any file\n")
		}
	} else {
		return nil, err
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	hash := justhash(data[:len(data)-20])
	//the index file has the lst 160 bits (i.e. 20 bytes) as the sha-1 checksum of all the bits tat come before it
	//we need to ensure that it matches before considering the data valid
	if bytes.Compare(hash[:], data[:(len(data)-20)]) != 0 {
		return nil, errors.New("Checksum is not equal to file digest. File has been tampered with")
	}
	hdr := data[:12]
	sign := hdr[:4]
	version := binary.BigEndian.Uint32(hdr[4:8])
	numEntries := binary.BigEndian.Uint32(hdr[8:])
	//we need to check what the header says.
	if !bytes.Equal(sign, []byte{'D', 'I', 'R', 'C'}) {
		return nil, fmt.Errorf("bad index file sha1 signature: %s", sign)
	}
	if version != 2 {
		return nil, fmt.Errorf("Version number must be at least 2, got %d", version)
	}
	//now for the index entries :
	//we need to use the unix fstat
	//the index files are listed between the 12-byte header and the 20-byte checksum
	indEntries := data[12:(len(data) - 20)]
	indexes, err := unmarshal(indEntries)
	if err != nil {
		return nil, err
	}
	if len(indexes) != int(numEntries) {
		return nil, fmt.Errorf("Number of enteries does not equal to what the head specified")
	}
	i.entries = indexes
	return i, nil
}

func unmarshal(data []byte) ([]*IdxEntry, error) {
	//length of deterministic bytes = 64
	var indexEntries []*IdxEntry
	b := bufio.NewReader(bytes.NewReader(data))
	for {
		//the pre-path length is 62 bytes
		buf := make([]byte, 62)
		_, err := io.ReadFull(b, buf)
		if err == io.ErrUnexpectedEOF {
			return nil, err
		}
		if err == io.EOF {
			break
		}
		if p, err := b.ReadBytes(Sep); err != nil {
			return nil, err
		} else {
			buf = append(buf, p[:len(p)-1]...) //exclude the delemiter since `ReadBytes` includes the delimiter
		}
		indexEntries = append(indexEntries, destructure(buf))

		// now we need to discard a specific number of bytes that were used to pad the entry
		// since in writing the index, we paded the entry to a multiple of eight bytes while keeping the name NUL-terminated
		// len(buf)+1 is the length we've read so far, since we had to discard the null byte ending the name string earlier
		adv := int(math.Ceil(float64(len(buf)+1)/8)*8) - (len(buf) + 1)
		_, err = io.CopyN(io.Discard, b, int64(adv))
		if err != nil {
			return nil, err
		}
	}

	return indexEntries, nil
}

func destructure(b []byte) *IdxEntry {
	var e *IdxEntry
	e.cTime = time.Unix(int64(binary.BigEndian.Uint32(b[0:4])), int64(binary.BigEndian.Uint32(b[4:8])))
	e.mTime = time.Unix(int64(binary.BigEndian.Uint32(b[8:12])), int64(binary.BigEndian.Uint32(b[12:16])))
	e.dev = binary.BigEndian.Uint32(b[16:20])
	e.inode = binary.BigEndian.Uint32(b[20:24])
	e.mode = binary.BigEndian.Uint32(b[24:28])
	e.uid = binary.BigEndian.Uint32(b[28:32])
	e.gid = binary.BigEndian.Uint32(b[32:36])
	e.fsize = binary.BigEndian.Uint32(b[36:40])
	e.sha = *(*[20]byte)(b[40:60]) // trick converting slice to array pointer
	e.flags = binary.BigEndian.Uint16(b[60:62])
	e.path = b[62:]
	return e
}

func encodeNewIdxEntry(path string) (io.Reader, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	blob, err := io.ReadAll(f)
	sha1 := justhash(blob)
	var stat *unix.Stat_t
	err = unix.Stat(path, stat)
	if err != nil {
		return nil, err
	}
	entry := mapStatToEntry(stat, path, sha1)
	return entry.marshall(), nil
}

func newIdxEntry(path string) (*IdxEntry, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	blob, err := io.ReadAll(f)
	sha1 := justhash(blob)
	var stat *unix.Stat_t
	err = unix.Stat(path, stat)
	if err != nil {
		return nil, err
	}
	return mapStatToEntry(stat, path, sha1), nil
}

func mapStatToEntry(stat *unix.Stat_t, path string, sha1 Sha1) *IdxEntry {
	e := IdxEntry{
		cTime: time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec)),
		mTime: time.Unix(int64(stat.Mtim.Sec), int64(stat.Mtim.Nsec)),
		dev:   uint32(stat.Dev),
		inode: uint32(stat.Ino),
		mode:  stat.Mode,
		uid:   stat.Uid,
		gid:   stat.Gid,
		fsize: uint32(stat.Size),
		sha:   sha1,
		path:  []byte(path),

		// comeback for flag
		flags: setFlags(path),
	}

	return &e
}

func (e *IdxEntry) marshall() io.Reader {
	var b bytes.Buffer
	b.Grow(70) // i expect the buffer t be greater than 62. the extra 8 on top is for the path string
	slice := make([]byte, 4)
	binary.BigEndian.PutUint32(slice, uint32(e.cTime.Unix()))
	b.Write(slice)
	binary.BigEndian.PutUint32(slice, uint32(e.cTime.Nanosecond()))
	b.Write(slice)
	binary.BigEndian.PutUint32(slice, uint32(e.mTime.Unix()))
	b.Write(slice)
	binary.BigEndian.PutUint32(slice, uint32(e.mTime.Nanosecond()))
	b.Write(slice)
	binary.BigEndian.PutUint32(slice, e.dev)
	b.Write(slice)
	binary.BigEndian.PutUint32(slice, e.inode)
	b.Write(slice)
	binary.BigEndian.PutUint32(slice, e.mode)
	b.Write(slice)
	binary.BigEndian.PutUint32(slice, e.uid)
	b.Write(slice)
	binary.BigEndian.PutUint32(slice, e.gid)
	b.Write(slice)
	binary.BigEndian.PutUint32(slice, e.fsize)
	b.Write(slice)
	b.Write(e.sha[:])
	flags := encodeFlags(string(e.path))
	b.Write(flags[:])
	buf := b.Bytes()
	pad := 8 - (len(buf) % 8)
	buf = append(buf, bytes.Repeat([]byte{'\x00'}, pad)...)
	return bytes.NewReader(buf)
}

//1-bit assume-valid flag (false); 1-bit extended flag (must be zero in version 2); 2-bit stage (during merge);
//12-bit name length if the length is less than 0xFFF, otherwise 0xFFF is stored in this field.
//comeback: check
//The value is 9 in decimal, or 0x9.
func encodeFlags(name string) [2]byte {
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

func setFlags(name string) uint16 {
	i := uint16(0)
	l := len(name)
	if l > 0xFFF {
		i = 0xFFF
	}
	return i
}

//write the index file, given a slice of index
//this is the function that stages files
//IndexEntry file integers in git are written in NE.
func (got *Got) UpIndexEntries(entries []*IdxEntry) error {
	var hdr []byte
	hdr = append(hdr, []byte("DIRC")...)
	//buf is apparently reusable
	var buf []byte
	binary.BigEndian.PutUint32(buf, 2)
	hdr = append(hdr, buf[:4]...)
	//use the same buffer, since the buffer does not keep its state. It starts over
	binary.BigEndian.PutUint32(buf, uint32(len(entries)))
	hdr = append(hdr, buf[:4]...)
	var b bytes.Buffer
	hasher := sha1.New()
	w := io.MultiWriter(&b, hasher)
	w.Write(hdr) // write the header
	// now write the entries
	for _, entry := range entries {
		io.Copy(w, entry.marshall())
	}
	// version 2 has empty extensions, so the only thing we need to write here is the 160-bit Sha1 checksum
	checksum := hasher.Sum(nil)
	f, err := os.OpenFile(filepath.Join(".git", "index"), os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	defer f.Close()
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, &b); err != nil {
		return err
	}
	if _, err := f.Write(checksum); err != nil {
		return err
	}
	return nil
}

// comeback
func (got *Got) UpdateIndex(ctx context.Context, all, remove bool) error {
	return nil
}

func (i *Idx) Append(path string) {

}

func (i *Idx) Remove(path string) error {
	return nil
}
