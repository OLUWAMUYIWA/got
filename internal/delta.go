package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)



func applyDelta(delta []byte, base io.ReadSeeker) (io.Reader, error) {
	d:= bufio.NewReader(bytes.NewBuffer(delta))
	_, err :=  uvarint(d) //baseSize
	if err != nil {
		return nil, fmt.Errorf("While applying delta: %w",err)
	}
	finalSize, err := uvarint(d)
	if err != nil {
		return nil, fmt.Errorf("While applying delta: %w",err)
	}
	var ret bytes.Buffer


	for {
		first, err := d.ReadByte()
		if err != nil {
			return nil, err
		}
		if first == 0 {
			// ==== Reserved instruction

			// +----------+============
			// | 00000000 |
			// +----------+============
			return nil, fmt.Errorf("Reserved instruction. cannot process")
		} else if first&0x80 == 0 { //seventh bit is set. starting from zeroth
			  // ==== Instruction to add new data

			  // +----------+============+
			  // | 0xxxxxxx |    data    |
			  // +----------+============+
			  size := int(first)
			  buf := make([]byte, size)
			  _, err := d.Read(buf)
			  if err != nil {
			  	return nil, err
			  }
			  ret.Write(buf)
		} else {
			  // ==== Instruction to copy from base object

			  // +----------+---------+---------+---------+---------+-------+-------+-------+
			  // | 1xxxxxxx | offset1 | offset2 | offset3 | offset4 | size1 | size2 | size3 |
			  // +----------+---------+---------+---------+---------+-------+-------+-------+


			var off, size uint32
			shift := 0
			for i := 0; i < 4; i++ {
				curr, err := d.ReadByte()
				if err != nil {
					//TODO
				}
				if isKthSet(first, i) {
					off |= uint32(curr) << shift
				}
				shift += 8
			}
			shift = 0
			for i := 4; i < 7; i ++ {
				curr, err := d.ReadByte()
				if err != nil {
					//TODO
				}
				if isKthSet(first, i) {
					size |= uint32(curr) << shift
				}
				shift += 8
			}

			if size == 0 {
				//TODO: check
				size = 0x10000
			}

			_, err = base.Seek(int64(off), io.SeekStart)
			if err != nil {
				return nil, err
			}
			data := make([]byte, size)
			n, err := base.Read(data)
			if n != int(size) {
				return nil, fmt.Errorf("Did not read fully")
			}
			if err != nil {
				return nil,err
			}
			ret.Write(data)
		}
	}
	if finalSize != uint64(ret.Len()) {
		return nil, fmt.Errorf("expected final size not matched")
	}
	return &ret, nil
}