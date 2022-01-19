package proto

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
)

//IMPORTANT:
//source: https://github.com/git/git/blob/master/Documentation/technical/pack-format.txt

//TODO:check all the binary encoding functions in this project to see if they donot fail the size test. i.e _=b[3] for Uint32() and _=b[7] for Uint64()
// Valid object types are:

// - OBJ_COMMIT (1)
// - OBJ_TREE (2)
// - OBJ_BLOB (3)
// - OBJ_TAG (4)
// - OBJ_OFS_DELTA (6)
// - OBJ_REF_DELTA (7)

type pkObjectType = uint8

const (
	_ pkObjectType = iota // Type 0 is invalid
	OBJ_COMMIT
	OBJ_TREE
	OBJ_BLOB
	OBJ_TAG
	_ //reserved for future expansion
	OBJ_OFS_DELTA
	OBJ_REF_DELTA
)

//TODO: TRY USING SOMETHING LIKE THE ENCODER INTERFACE TO DESCRIBE THE FUNCTIONALITY OF THE PACK. CHECK HEX ENCODER
type Pack struct {
	Sha     string
	Objects map[string]*pkObject
}

type pkObject struct {
	idx        idx
	ty         pkObjectType
	data       []byte //would either be the uncompressed data, of the uncompressed delta
	sizeUncomp uint64
	sizeComp   uint64
	baseObj    string
	baseOffset uint64
	depth      uint32
}

type idx struct {
	sha    []byte
	offset uint64
}

func (p *pkObject) Type() string {
	switch p.ty {
	case OBJ_COMMIT:
		return "commit"
	case OBJ_TREE:
		return "tree"
	case OBJ_BLOB:
		return "blob"
	case OBJ_OFS_DELTA:
		return "OBJ_OFS_DELTA"
	case OBJ_REF_DELTA:
		return "OBJ_REF_DELTA"
	default:
		return ""
	}
}

func parsePackFile(pack, idx io.ReadSeekCloser) (Pack, error) {
	defer pack.Close()

	//first parse the idx file
	//TODO: can be made concurrent
	idxes, packSha, err := parseIdxFile(idx)
	if err != nil {
		return Pack{}, fmt.Errorf("Error while parsing idx file: %w", err)
	}
	//we need just 12 bytes for the header
	hdr := make([]byte, 12)
	n, err := io.ReadFull(pack, hdr)
	if err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return Pack{}, fmt.Errorf("Could not read whole header: %w", err)
		}
		return Pack{}, fmt.Errorf("Failed reading: %w", err)
	}

	sig := "PACK"
	if bytes.Compare([]byte(sig), hdr[0:4]) != 0 {
		return Pack{}, &ProtoErr{Context: "not a valid pack file, Signature is not PACK as was expected"}
	}
	//chose uint64 to prevent overflow
	version := binary.BigEndian.Uint64(hdr[4:8])
	if version < 2 || version > 3 {
		//TODO: handle error better
		return Pack{}, fmt.Errorf("Bad formatting")
	}
	objNum := binary.BigEndian.Uint32(hdr[8:12])
	if len(idxes) != int(objNum) {
		return Pack{}, fmt.Errorf("idx file and pack file disagree on number of objects")
	}
	if n != 12 {
		return Pack{}, &ProtoErr{Context: "header not up to 12 bytes, check"}
	}
	packs := make(map[string]*pkObject)
	for _, idx := range idxes {
		obj := pkObject{}

		obj.idx = idx
		pack.Seek(int64(idx.offset), io.SeekStart)
		var meta []byte
		//get meta, i.e. type + length
		buf := make([]byte, 1)
		for {
			_, err := pack.Read(buf)
			if err != nil {
				return Pack{}, &ProtoErr{Context: "Error reading pack_file", Inner: err}
			}
			meta = append(meta, buf[0])
			if !isKthSet(buf[0], 7) {
				break
			}
		}
		ty, rem := packTypeAndRem(meta[0])
		var size uint64
		var d int
		size |= uint64(rem)
		d += 4
		for b := range meta {
			if b < 0x80 { //last byte
				size |= uint64(b) << d
			}
			size |= uint64(b&0x7f) << d
			d += 7
		}

		switch ty {
		case OBJ_COMMIT, OBJ_TREE, OBJ_BLOB:
			{
				buf := make([]byte, size)
				r := io.LimitReader(pack, int64(size))
				z, _ := zlib.NewReader(r)
				_, err = z.Read(buf)
				if err != nil {
					return Pack{}, err
				}
				obj.data = buf
				obj.ty = ty
				z.Close()
			}
		case OBJ_REF_DELTA:
			{
				//we have the name of the base object and the delta data
				sha := make([]byte, 20)
				_, err := pack.Read(sha)
				if err != nil {
					//TODO
				}
				obj.baseObj = hex.EncodeToString(sha)
				buf := make([]byte, size)
				z, _ := zlib.NewReader(pack)
				_, err = z.Read(buf)
				if err != nil {
					//TODO
				}
				obj.data = buf
			}
		case OBJ_OFS_DELTA:
			{
				var MSB uint8 = 0x80
				var negOffset uint64
				d := 0
				for {
					b := make([]byte, 1)
					_, err := pack.Read(b)
					if err != nil {
						//TODO
					}
					negOffset |= uint64(b[0]&0x7f) << d
					d += 7
					if MSB&0x80 == 0 {
						break
					}
				}
				obj.baseOffset = uint64(obj.idx.offset) - negOffset
				obj.data = make([]byte, size)
				z, _ := zlib.NewReader(pack)
				_, err := z.Read(obj.data)
				if err != nil {
					//TODO
					continue
				}
				z.Close()

			}
		default:
			return Pack{}, fmt.Errorf("Bad format")
		}
		packs[hex.EncodeToString(idx.sha)] = &obj
	}

	p := Pack{
		Sha:     hex.EncodeToString(packSha),
		Objects: packs,
	}

	return p, nil
}

//checks if the kth bit is set. contract: n <= 255. indexed beginning from zero
func isKthSet(i uint8, k int) bool {
	mask := uint8(1) << k
	return (i & mask) != 0
}

//TODO: come back
func getBitRange(i uint8, l int, r int) uint8 {
	return (i >> uint8(r)) & (^uint8(0) >> (8 - (l - r + 1)))
}

func packTypeAndRem(i uint8) (pkObjectType, uint8) {
	ty := pkObjectType(getBitRange(i, 6, 4))
	rem := getBitRange(i, 3, 0)
	return ty, rem
}

func parseIdxFile(r io.ReadCloser) ([]idx, []byte, error) {
	defer r.Close()
	//idRdr := bufio.NewReader(idx)
	//this is needed so we can do the checksumming later
	wtr := new(bytes.Buffer)
	idRdr := io.TeeReader(r, wtr)
	hdr := make([]byte, 8)
	_, err := io.ReadFull(idRdr, hdr)
	if err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, nil, errors.New("could not read header completely")
		}
		return nil, nil, err
	}
	//check firstpart of hdr
	//A 4-byte magic number '\377tOc' which is an unreasonable fanout[0] value.
	if bytes.Compare(hdr[0:4], []byte{255, 116, 79, 99}) != 0 {
		return nil, nil, errors.New("invalid header")
	}

	//check version
	//A 4-byte version number (= 2)
	if ver := binary.BigEndian.Uint32(hdr[4:8]); ver != 2 {
		return nil, nil, errors.New("wrong version included. expected 2")
	}
	buf := make([]byte, 1026)
	_, err = io.ReadFull(idRdr, buf)
	if err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, nil, errors.New("could not read fan-out table completely")
		}
		return nil, nil, err
	}

	//A 256-entry fan-out table just like v1.
	fanout := make([][]byte, 256)
	//let's make the table a seuence of 4 bytes each
	//TODO:check
	for i := 0; i < len(buf); {
		fanout[i/4] = buf[i : i+4]
		i += 4
	}
	//from the last element of the fan-out table we get the total number of objects in the pack
	objNum := binary.BigEndian.Uint32(fanout[len(fanout)-1])

	//A table of sorted object names
	shas := make([]byte, objNum*20)
	_, err = io.ReadFull(idRdr, shas)
	if err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, nil, errors.New("could not read names completely")
		}
		return nil, nil, err
	}
	idxes := []idx{}
	//TODO: check
	for i, j := 0, 0; i < int(objNum); i++ {
		idxes[i].sha = shas[j : j+20] //hex.EncodeToString(shas[j:j+20]) i'm trying to avoid allocation
		j += 20
	}

	valuesCRC32Flat := make([]byte, objNum*4)
	_, err = io.ReadFull(idRdr, valuesCRC32Flat)
	if err != nil {
		return nil, nil, err
	}
	if err := checkCRC(valuesCRC32Flat); err != nil {
		return nil, nil, fmt.Errorf("Cyclic Redundancy Check failed")
	}
	offFlat := make([]byte, objNum*4)
	_, err = io.ReadFull(idRdr, offFlat)
	if err != nil {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, nil, fmt.Errorf("Could not read offsets completely")
		}
	}

	//chunk the offests
	offsets := make([][4]byte, objNum)
	j, k := 0, 0
	for i := 0; i < len(offFlat); i++ {
		offsets[j][k] = offFlat[i]
		k++
		if k == 4 {
			k = 0
			j += 1
		}
	}

	// These are usually 31-bit pack file offsets
	offs := []uint64{}
	m := make(map[int]uint64)
	isLarge := 0
	for i := range offsets {
		ui32 := binary.BigEndian.Uint32(offsets[i][:])
		ui64 := uint64(ui32)
		if ui64&0x80000000 != 0 {
			ui64 -= 0x80000000
			m[i] = ui64
			offs = append(offs, ui64)
			isLarge++
		} else {
			offs = append(offs, ui64)
		}
	}

	//dealing with the larges and the trailer
	rem, err := io.ReadAll(idRdr)
	pkCheck, idxCheck := make([]byte, 20), make([]byte, 20)
	remB := bytes.NewReader(rem)
	rChecks := io.NewSectionReader(remB, int64(len(rem)-40), 40)
	_, err = rChecks.Read(pkCheck)
	_, err = rChecks.Read(idxCheck)
	if isLarge != 0 {
		rLarges := io.NewSectionReader(remB, 0, int64(len(rem)-40))
		off := make([]byte, 8)
		for k, v := range m {
			_, _ = rLarges.ReadAt(off, int64(v))
			offs[k] = binary.BigEndian.Uint64(off)
		}
	}

	for i, off := range offs {
		idxes[i].offset = off
	}

	hasher := sha1.New()
	_, err = wtr.WriteTo(hasher)
	if err != nil {
		return nil, nil, fmt.Errorf("error writing to hasher")
	}
	hash := hasher.Sum(nil)
	if same := bytes.Compare(hash, idxCheck); same != 0 {
		return nil, nil, fmt.Errorf("Bad Checksum")
	}

	return idxes, pkCheck, nil
}

func writePackFile() {

}

func decodeSize(r bufio.Reader) uint64 {
	return 0
}
