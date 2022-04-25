



|||||||INDEX !!!

type Index struct {
	entries []*IndexEntry
}

<!-- // IndexEntry holds a snapshot of the content of the working tree,
// and it is this snapshot that is taken as the contents of the next commit.

//The index stores all the info about files needed to write a tree object
//we want to ensure that the index is a multiple of 8 bytes, so we might pad with null bytes if need be
type IndexEntry struct {
	ctime_s     [4]byte
	ctime_ns    [4]byte
	mtime_s     [4]byte
	mtime_ns    [4]byte
	dev         [4]byte
	inode       [4]byte
	mode        [4]byte
	uid         [4]byte
	gid         [4]byte
	f_size      [4]byte
	sha 		[20]byte
	flags       [2]byte
	//ver         [2]byte
	path []byte
} -->
// func destructureIntoIndex(b []byte) *IndexEntry {
// 	var i *IndexEntry
<!--  -->
// 	start, lim := 0, 4
// 	//closure crunches the four-byters
// 	get_next_four := func(b []byte) [4]byte {
// 		var arr [4]byte
// 		for i, pos := start, 0; i < lim; i, pos = i+1, pos+1 {
// 			arr[pos] = b[i]
// 		}
// 		start, lim = start+4, lim+4
// 		return arr
// 	}
// 	i.ctime_s = get_next_four(b)
// 	i.ctime_ns = get_next_four(b)
// 	i.mtime_s = get_next_four(b)
// 	i.mtime_ns = get_next_four(b)
// 	i.dev = get_next_four(b)
// 	i.inode = get_next_four(b)
// 	i.mode = get_next_four(b)
// 	i.uid = get_next_four(b)
// 	i.gid = get_next_four(b)
// 	i.f_size = get_next_four(b)
// 	//now to the sha-1
// 	//lim is 4 bytes ahead of start, we want it to be 20 because of sha1
// 	lim += 16
// 	get_next_twenty := func(b []byte) [20]byte {
// 		var arr [20]byte
// 		for i, pos := start, 0; i < lim; i, pos = i+1, pos+1 {
// 			arr[pos] = b[i]
// 		}
// 		//this op levels the gap. both start and op are at the same position now
// 		start, lim = start+20, lim+16
// 		return arr
// 	}
// 	i.sha = get_next_twenty(b)
// 	//now start and lim are th the same number, but incr lim by two
// 	lim += 2
// 	get_next_two := func(b []byte) [2]byte {
// 		var arr [2]byte
// 		for i, pos := start, 0; i < lim; i, pos = i+1, pos+1 {
// 			arr[pos] = b[i]
// 		}
// 		start = start + 2
// 		return arr
// 	}
// 	i.flags = get_next_two(b)
// 	i.path = b[lim:]

// 	return i
// }

// func newIndexEntry(path string) (*IndexEntry, error) {
// 	f, err := os.OpenFile(path, os.O_RDONLY, 0)
// 	if err != nil {
// 		return nil, err
// 	}
// 	blob, err := io.ReadAll(f)
// 	sha1 := justhash(blob)
// 	var stat unix.Stat_t
// 	err = unix.Stat(path, &stat)
// 	if err != nil {
// 		return nil, err
// 	}
// 	i := mapStatToIndex(&stat, sha1, path)
// 	i.flags = setUpFlags(path)
// 	i.path = []byte(path)
// 	return i, err
// }
<!-- 
func mapStatToIndex(stat *unix.Stat_t, sha1 [20]byte, path string) *IndexEntry {
	var i IndexEntry
	i.ctime_s = mapint64ToBytes(stat.Ctim.Sec)
	i.ctime_ns = mapint64ToBytes(stat.Ctim.Nsec)
	i.mtime_s = mapint64ToBytes(stat.Mtim.Sec)
	i.mtime_ns = mapint64ToBytes(stat.Mtim.Nsec)
	//at first I worried that the conversions might alter the values of the bits.
	//turns out that isn't true. Nothing happens to the bit, except the new integer size is smaller, causing a loss
	//the difference is in the interpretation of the values by the compiler, since signed uses 2's compliment to evaluate
	//while unsigned just translates the bits
	i.dev = mapint64ToBytes(int64(stat.Dev))
	i.inode = mapint64ToBytes(int64(stat.Ino))
	i.mode = mapint64ToBytes(int64(stat.Mode))
	i.uid = mapint64ToBytes(int64(stat.Uid))
	i.gid = mapint64ToBytes(int64(stat.Gid))
	i.f_size = mapint64ToBytes(int64(stat.Size))
	i.sha = sha1
	i.flags = setUpFlags(path)
	return &i
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





-->


// func (i *IndexEntry) marshall() []byte {
// 	//I see no poit catching errors, these ops are not I/O. Only cpu failure here, I think
// 	var b bytes.Buffer
// 	b.Write(i.ctime_s[:])
// 	b.Write(i.ctime_ns[:])
// 	b.Write(i.mtime_s[:])
// 	b.Write(i.mtime_ns[:])
// 	b.Write(i.dev[:])
// 	b.Write(i.inode[:])
// 	b.Write(i.mode[:])
// 	b.Write(i.uid[:])
// 	b.Write(i.gid[:])
// 	b.Write(i.f_size[:])
// 	b.Write(i.sha[:])
// 	b.Write(i.flags[:])
// 	//b.Write(i.ver[:])
// 	pathlen := len(i.path)
// 	//data b4 path = 62 bytes. Doing this transformation is a way of appendin the path and still getting a multiple of 8
// 	//this is how we do the padding
// 	datalen := int(math.Ceil(float64(b.Len()+pathlen+8)/8) * 8)
// 	fill := datalen - (b.Len() + pathlen)
// 	b.Write(i.path)
// 	//Fill it with zero bytes
// 	space_fill := bytes.Repeat([]byte{Sep}, fill)
// 	b.Write(space_fill)
// 	return b.Bytes()
// }
