package pkg

import (
	"bufio"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func shaToString(s Sha1) string {
	return hex.EncodeToString(s[:])
}

func bytesToSha(hex []byte) [20]byte {
	if len(hex) != 20 {
		//don't even try to use this method if the length of the slic is not exactly 20. Thank you!
		panic(fmt.Errorf("length not equal to 20"))
	}
	var b [20]byte
	copy(b[:], hex)
	return b
}

func strToSha(str string) Sha1 {
	h, _ := hex.DecodeString(str)
	var hash Sha1
	copy(hash[:], h)
	return hash
}

//getConfig gets into got's config and gets the username and email
//I chose to use json because reading and writing it is clean and easy. No stress
func getConfig() (string, string, error) {
	confRoot, err := os.UserConfigDir()
	if err != nil {
		return "", "", &OpErr{Context: "IO: While Getting Config file ", inner: err}
	}
	f, err := os.Open(filepath.Join(confRoot, ".git", ".config"))
	dec := json.NewDecoder(f)
	var user User
	err = dec.Decode(&user)
	if err != nil {
		return "", "", &OpErr{Context: "IO: While Getting Config file ", inner: err}
	}
	return user.Uname, user.Email, nil
}

// func diff(stra, strb string) string {
// 	dmp := diffmatchpatch.New()
// 	diffs := dmp.DiffMain(stra, strb, true)
// 	return dmp.DiffPrettyText(diffs)
// }

func writeToFile(path string, b []byte) error {
	f, err := os.OpenFile(path, os.O_APPEND, 0)
	defer f.Close()
	return err

	bf := bufio.NewWriter(f)
	_, err = bf.Write(b)
	return err
}

//zlib compress
func compress(writer io.Writer, data []byte) error {
	comp := zlib.NewWriter(writer)
	_, err := comp.Write(data)
	if err != nil {
		return &OpErr{Context: "IO: while compressing ", inner: err}
	}
	err = comp.Flush()
	if err != nil {
		return &OpErr{Context: "IO: while compressing ", inner: err}
	}
	return nil
}

//zlib decompress
func decompress(rdr io.Reader, dst io.Writer) error {
	comp, err := zlib.NewReader(rdr)
	if err != nil {
		return err
	}
	//to decompress data from zlib, Go is marvelously helpful here, you only have to read from the zlib reader
	//I use copy here because it is cool and fast
	if _, err := io.Copy(dst, comp); err != nil {
		return err
	}
	comp.Close()
	return nil
}

type Sha1 = [20]byte

//justhash is the hash function used by hashobj
func justhash(data []byte) Sha1 {
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
	var h Sha1
	copy(h[:], hahser.Sum(nil))
	return h
}

func hashWithObjFormat(data []byte, ty string) (Sha1, error) {
	var s strings.Builder
	hdr := fmt.Sprintf("%s %d", ty, len(data))
	//i see no reason to handle errors here since no I/O is happening
	//Builder only implements io.Writer.
	s.WriteString(hdr)
	s.WriteByte(Sep)
	s.Write(data)
	b := []byte(s.String())
	raw := justhash(b)
	return raw, nil
}

//TODO: not needed
//reversing the order of the bytes seem necessary in cases where I have to convert from network endian to integers in memory
//to interpret the byte stream (rightly) straightforwardly as BigEndian would give me an in-memory representation that isn't
//consistent with my machine specs.
// func reverse(b []byte) []byte {
// 	pos := len(b)-1
// 	half := len(b)/2
// 	for i :=0; i < half; i++ {
// 		b[i], b[pos] = b[pos], b[i]
// 		pos--
// 	}
// 	return b
// }

func uvarint(r io.Reader) (uint64, error) {
	b := make([]byte, 1)
	var res uint64
	d := 0
	for {
		_, err := r.Read(b)
		if err != nil {
			return 0, err
		}
		if b[0]&0x80 == 0 {
			return res | uint64(b[0])<<d, nil
		}
		res |= uint64(b[0]&0x7f) << d
		d += 7
	}
}
