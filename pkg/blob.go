package pkg

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type Blob struct {
	sha  Sha1
	size int64
	//uncompressed data
	data []byte
}

func (blob *Blob) Hash(wkdir string) ([]byte, error) {
	b, err := HashObj(blob.Type(), blob.data, wkdir)
	return b[:], err
}

func (c *Blob) Type() string {
	return "blob"
}

func (c *Blob) String() string {
	return "blob"
}

func parseBlob(rdr io.Reader) (*Blob, error) {
	b := bufio.NewReader(rdr)
	var d bytes.Buffer
	blob := &Blob{}
	if ty, err := b.ReadBytes(' '); err != nil {
		if bytes.Compare(ty, []byte("blob")) != 0 {
			return nil, fmt.Errorf("Expected blob, found: %s", ty)
		}
		if len, err := b.ReadBytes(Sep); err != nil {
			blob.size = int64(binary.BigEndian.Uint32(len)) //comeback
			if _, err := io.CopyN(&d, b, int64(blob.size)); err != nil {
				blob.data = d.Bytes()
				return blob, nil
			}
		}
	}
	return nil, fmt.Errorf("Error reading blob")
}
