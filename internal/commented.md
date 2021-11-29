//HashObject returns the hash of the file it hashes
//plumber + helper function
//needed for blobs, trees, and commit hashes
// func (got *Got) HashObject(data []byte, ty string, w bool) []byte {
// 	base, err := os.Getwd()
// 	if err != nil {
// 		got.GotErr(err)
// 	}
// 	//use a string builder because it minimizzed memory allocation, which is expensive
// 	//each write appends to the builder
// 	//IGNORING errors here, too many writes, error handling will bloat the code.
// 	var s strings.Builder
// 	hdr := fmt.Sprintf("%s %d", ty, len(data))
// 	//i see no reason to handle errors here since no I/O is happening
// 	//Builder only implements io.Writer.
// 	s.WriteString(hdr)
// 	s.WriteByte(Sep)
// 	s.Write(data)
// 	b := []byte(s.String())
// 	raw := justhash(b)
// 	if w {
// 		//the byte result must be converted to hex string as that is how it is useful to us
// 		//we could either use fmt or hex.EncodeString here. Both works fine
// 		hash_str := fmt.Sprintf("%x", raw)
// 		//first two characters (1 byte) are the name of the directory. The remaining 38 (19 bytes) are the  name of the file
// 		//that contains the compressed version of the blob.
// 		//remember that sha1 produces a 20-byte hash (160 bits, or 40 hex characters)
// 		path := filepath.Join(base, ".git/objects/", hash_str[:2])
// 		err = os.MkdirAll(path, 0777)
// 		got.GotErr(err)
// 		fPath := filepath.Join(path, hash_str[2:])
// 		f, err := os.Create(fPath)
// 		got.GotErr(err)
// 		defer f.Close()
// 		//the actual file is then compressed and stored in the file created
// 		err = compress(f, b)
// 		got.GotErr(err)
// 	}

// 	return raw
// }