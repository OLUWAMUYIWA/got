package internal

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

//###### PLUMBERS!!!!! #######
//The Hitokiri Battousais. They do the killing, and they have no shame in doing it
//Most Plumbers should send their errors upstream. Porcelains will handle them. Except when it is not a fit directory were in. 
//This exception is because plumbers can be used directly by the user too
//TODO:Check all the endianness in this code


func Config(conf ConfigObject) error {
	confRoot, err := os.UserConfigDir()
	if err != nil {
		return NotDefinedErr.addContext(err.Error())
	}
	err = os.Mkdir(filepath.Join(confRoot, ".git"), os.ModeDir)
	if err != nil {
		return IOWriteErr.addContext(err.Error())
	}
	f, err := os.Create(filepath.Join(confRoot, ".git", ".config"))
	if err != nil {
		return IOCreateErr.addContext(err.Error())
	}
	enc := json.NewEncoder(f)
	err = enc.Encode(conf)
	if err != nil {
		return IOWriteErr.addContext(err.Error())
	}
	return nil
}

//HashObject returns the hash of the file it hashes
//plumber + helper function
//needed for blobs, trees, and commit hashes
func (got *Got) HashObject(data []byte, ty string, w bool) []byte {
	base, err := os.Getwd()
	if err != nil {
		got.GotErr(err)
	}
	//use a string builder because it minimizzed memory allocation, which is expensive
	//each write appends to the builder
	//IGNORING errors here, too many writes, error handling will bloat the code.
	var s strings.Builder
	hdr := fmt.Sprintf("%s %d", ty, len(data))
	//i see no reason to handle errors here since no I/O is happening
	//Builder only implements io.Writer.
	s.WriteString(hdr)																				
	s.WriteByte(sep)
	s.Write(data)
	b := []byte(s.String())
	raw := justhash(b)
	if w {
		//the byte result must be converted to hex string as that is how it is useful to us
		//we could either use fmt or hex.EncodeString here. Both works fine
		hash_str := fmt.Sprintf("%x", raw)
		//first two characters (1 byte) are the name of the directory. The remaining 38 (19 bytes) are the  name of the file
		//that contains the compressed version of the blob.
		//remember that sha1 produces a 20-byte hash (160 bits, or 40 hex characters)
		path := filepath.Join(base, ".git/objects/", hash_str[:2])
		err = os.MkdirAll(path, 0777)
		got.GotErr(err)
		fPath := filepath.Join(path, hash_str[2:])
		f, err := os.Create(fPath)
		got.GotErr(err)
		defer f.Close()
		//the actual file is then compressed and stored in the file created
		err = compress(f, b)
		got.GotErr(err)
	}

	return raw
}

//FindObject takes a sha1 prefix. It returns the path to the object file. It doesn't care t open it 
//another method does that
//we want to have findObject such that even though we do not have
//a full 20 byte string, we can still find the path to the file
//it is unlikely to find two blobs with the same sha1 prefix, but if such happens, we return
//an error and expect the user to provide a longer string
func (got *Got) FindObject(prefix string) (string, error) {
	if is, _ := IsGit(); !is {
		got.logger.Fatalf("Not a valid git directory\n")
	}
	//first check for length
	switch l := len(prefix); {
	case l <= 2:
		return "", fmt.Errorf("the prefix provided is not sufficient for a search, ensure it's more than two")
	case l > 40:
		return "", fmt.Errorf("way too long. How did you come about a string longer than 40? darn it! This is sha-1")
	}
	path := filepath.Join(".git/objects", prefix[:2])
	entries, err := os.ReadDir(path)
	got.GotErr(err)
	location := ""
	//we also need to be sure that our given prefix is unique too. If it isn't we may be return ing the wrong file
	num := 0

	for _, entry := range entries {
		if entry.Type().IsRegular() && strings.HasPrefix(entry.Name(), prefix[2:]){
			num += 1
			location = entry.Name()
			if num > 1 {
				return "", fmt.Errorf("%s matches more than one blob", prefix)
			}
		}
	}
	if len(location) != 0 {
		return filepath.Join(".git/objects", prefix[:2], location), nil
	} else {
		return "", fmt.Errorf("%s matches no blob", prefix)
	}
}


//ReadObject bulds on findObject. First, it finds the object,
//but it does more, it attempts to read it and assert that it contains valid git object files
//apart from that, it tries to understand what kind of git object it is
func (got *Got) ReadObject(prefix string) (string, string, []byte, error) {
	if is, _ := IsGit(); !is {
		got.logger.Fatalf("Not a valid git directory\n")
	}
	location, err := got.FindObject(prefix)
	if err != nil {
		return "", "", nil, err
	}
		//get the file name
	splits := strings.Split(location, "/")
	f_name := splits[len(splits)-1]

	f, err := os.Open(location)
	if err != nil {
		return "", "", nil, err
	}
	defer f.Close()
	//create a buffer to hold the bytes to be read from zlib
	//buffer implements io.Writer
	var b bytes.Buffer
	uncompress(f, &b)
	if err != nil {
		return "", "", nil, err
	}
	//remember that when writing the file, we put in a separator sep to demarcate the header from the body
	//sep is equal to byte(0)
	hdr, err := b.ReadBytes(sep)
	if err != nil {
		return "", "", nil, err
	}
	//the remainder after reading out the header from the buffer, just like were flushing th buffer
	//we expect nothing else to be in the buffer right now other than the data itself
	data := b.Bytes()
	dType, dLen := "", 0
	_, err = fmt.Sscanf(string(hdr),"%s %d", &dType, &dLen)
	if err != nil {
		return "", "", nil, err
	}
	if len(data) != int(dLen) {
		return "", "", nil, fmt.Errorf("the data is corrupt, specified length does not match length of data")
	}
	return f_name, dType, data, nil
}
//CatFile displays the file info using the git logger (set as os.Stdout). It uses flags to determine what it displays
func (got *Got) CatFile(prefix, mode string) {
	//normalize
	mode = strings.ToLower(mode)
	if is, _ := IsGit(); !is {
		got.logger.Fatalf("Not a valid git directory\n")
	}
	f_name, dType, data, err := got.ReadObject(prefix)
	//this error should just cause the program to exit.
	got.GotErr(err)
	switch mode {
	case "size":
		got.logger.Printf("File %s: Size: %d\n",f_name, len(data))
	case "type":
		got.logger.Printf("File %s Type: %s\n",f_name, dType)
	case "pretty":
		if dType == "commit" || dType == "blob" {
			got.logger.Printf("Content %s: \n%s", f_name, string(data))	
		} else if dType == "tree" {
			//is a directory. ewe need to read th tree before printing
			got.logger.Printf("Directory: \n")
			objs := got.deserTree(prefix)
			for _, obj := range objs {
				got.CatFile(obj.sha1, "pretty")
			}
		}
	default: 
		got.logger.Fatalf("Bad flag mode. Check again, must be either size, type, pretty \n")	
	}
}

//write the index file, given a slice of index
//the gt version of updataindex
//Index file integers in git are written in NE.
func (got *Got) UpdateIndex(entries []*Index) {
	if is, _ := IsGit(); !is {
		got.logger.Fatalf("Not a valid git directory\n")
	}
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
	err := got.writeToFile(filepath.Join(".git", "index"), index)
	if err != nil {
		got.GotErr(err)
	}
}

// read: https://mincong.io/2018/04/28/git-index/
//The index file contains:
// 12-byte header.
// A number of sorted index entries.
// Extensions. They are identified by signature.
// 160-bit SHA-1 over the content of the index file before this checksum.

func (got *Got) readIndexFile() []Index {
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
		got.GotErr(err)
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
	if strings.Compare(sign, "DIRC") != 0 {
		got.GotErr(fmt.Errorf("signature %s is not valid", sign))
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
	return indexes
}

func (got *Got) LsFiles(detail bool) {
	if is, _ := IsGit(); !is {
		got.logger.Fatalf("Not a valid git directory\n")
	}
	for _, ind := range got.readIndexFile() {
		path := string(ind.path)
		if detail {
			//eating the mode and sha1 is easy. As for the mode, we'll octal-format it later with fmt, and the sha too
			mode := binary.BigEndian.Uint32(ind.mode[:])
			sha1 := ind.sha1_obj_id[:]
			//stage number is the 3rd and 4th bit in the 16-bit flag
			//so first, we >> by 12 top put the first four bits on the rightmost
			//then, our mask will be 0b00000011, i.e. 3, we & against (or 0000000000000011) so that we'll keep only
			//the values of the last two bits intact, and cancel every othher before it, namely third and fourth to the last
			stage := (binary.BigEndian.Uint16(ind.flags[:]) >> 12) & 3
			s := fmt.Sprintf("Mode: %o sha1: %x  stage: %d path: %s\n", mode, sha1, stage, path)
			got.logger.Printf(s)
		} else {
			got.logger.Printf(path)
		}
	}
}

//we want to know the file that were changes, the ones that were deleted, and the ones that were added
func (got *Got) get_status() ([]string, []string, map[string]string) {
	if is, _ := IsGit(); !is {
		got.logger.Fatalf("Not a valid git directory\n")
	}
	var add, del []string
	mod := make(map[string]string)

	// Walk directory to fill up entries, but remove .git
	//TODO: check if this works fine
	entries := make([]fs.DirEntry, 0)
	var files []string
	//do two things at the same time: ensure .git is not inncludesd in files, and fill up files with all cleaned paths to non-dir files
	fs.WalkDir(os.DirFS("."), ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && d.Name() == ".git" {
			return fs.SkipDir
		}
		entries = append(entries, d)
		if !d.IsDir() {
			files = append(files, filepath.Clean(path))
		}
		return nil
	})
	//sort the file paths
	sort.Slice(files, func(i, j int) bool {
		return strings.Compare(files[i], files[j]) == -1
	})
	index := got.readIndexFile()
	var index_map map[string]Index
	for _, ind := range index {
		index_map[string(ind.path)] = ind
	}
	//stored hash in index differs from the new hash
	modified := func(files []string) {
		for _, f_path := range files {
			if ind, ok := index_map[f_path]; ok {
				f, err := os.Open(f_path)
				got.GotErr(err)
				cont, err := io.ReadAll(f)
				got.GotErr(err)
				
				if hex.EncodeToString(got.HashObject(cont, "blob", false)) != hex.EncodeToString(ind.sha1_obj_id[:]) {
					mod[string(ind.path)] = f_path
				}
			}
		}
	}

	added := func(files []string) {
		for _, file := range files {
			if _, ok := index_map[file]; !ok {
				add = append(add, file)
			}
		}
	}

	//no longer exists in the index file
	deleted := func(files []string) {
		for _, f_path := range files {
			if ind, ok := index_map[f_path]; ok {
				delete(index_map, string(ind.path))
			}
		}
		for p := range index_map {
			del = append(del, p)
		}
	}
	modified(files)
	added(files)
	deleted(files)

	return add, del, mod
}
func (got *Got) status() {
	if is, _ := IsGit(); !is {
		got.logger.Fatalf("Not a valid git directory\n")
	}
	add, del, mod := got.get_status()
	var s strings.Builder
	if len(mod) != 0 {
		s.WriteString("Modified files\n")
	}
	if len(add) != 0 {
		s.WriteString(fmt.Sprintf("Untracked  Files: \n%v\n", add))
	}
	if len(del) != 0 {
		s.WriteString(fmt.Sprintf("Deleted Files: \n%v\n", del))
	}

	for k, v := range mod {
		d := diff(k, v)
		s.WriteString(fmt.Sprintf("%s\n", d))
	}
	_, err := fmt.Fprintf(os.Stdout, "%s\n", s.String())
		got.GotErr(err)
	
}


func (got *Got) getLocalMAsterHash() string {
	if is, _ := IsGit(); !is {
		got.logger.Fatalf("Not a valid git directory\n")
	}
	path := filepath.Join(".git", "refs", "head", "master")
	f, err := os.Open(path)
		got.GotErr(err)
	
	//Hahaha. I always feel good anytime I'm able to explit Go's interface semantics
	var s strings.Builder
	_, err = io.Copy(&s, f)
		got.GotErr(err)
	return strings.TrimSpace(s.String())
}

//TODO: This function isn't what I want
//TODO: I do not know if os.setenv sets the environment variable permanently, like saving it in a .bash_profile, .bash_rc, /etc/profile, .zsh_rc etc.
//or maybe it is better to store it in a file in the git installation folder. I should think that is better
func SetEnvVars(name, email string) error {
	if err := os.Setenv("GOT_USERNAME", name); err != nil {
		return err
	}
	if err := os.Setenv("GOT_USER_MAIL", email); err != nil {
		return err
	}
	return nil
}



func (got *Got) deserTree(sha string) ([]Object) {
	if is, _ := IsGit(); !is {
		got.logger.Fatalf("Not a valid git directory\n")
	}
	objs := make([]Object, 0)
	_, ty, data, err := got.ReadObject(sha)
	got.GotErr(err)
	if ty != "tree" {
		got.GotErr("should be tree object")
	}
	if len(data) == 0 {
		got.GotErr("data is empty")
	}
	var path, sha1 string
	start := 0
	for {
		d := data[start:]
		if len(data) == 0 {
			break
		}
		split := bytes.SplitN(d, []byte(" "), 2)
		md :=  string(split[0])
		mode, err := strconv.ParseUint(md, 8, 32)
		got.GotErr(err)
		sep_pos := bytes.IndexByte(split[1], sep)
		path = string(split[1][:sep_pos])
		sha1 = string(split[1][sep_pos+1: sep_pos+21])
		start += len(split[0]) + 1 + sep_pos+1 + 20
		objs = append(objs, Object{mode, path, sha1})
	}
	
	return objs
}


func (got *Got) findTreeObjs(sha1 string) []string {
	var objs []string
	for _, obj := range got.deserTree(sha1) {
		if fs.FileMode(obj.mode) == os.ModeDir {
			objs = append(objs, got.findTreeObjs(obj.sha1)...)
		} else {
			objs =append(objs, obj.sha1)
		}
	}

	return objs
}


func (got *Got) findCommitObjs(sha1 string) []string {
	var objs []string
	_, ty, data, err := got.ReadObject(sha1)
	got.GotErr(err)
	if ty != "commit" {
		got.GotErr(fmt.Errorf("object is not a commit object"))
	}
	//one tree and 0 or more parents
	//find tree
	data_s := strings.Split(string(data), "\n")
	tree := ""
	parents := []string{}
	for _, item := range data_s {
		if strings.HasPrefix(item, "tree") {
			tree = strings.TrimPrefix(item, "tree ")
			objs = append(objs, got.findTreeObjs(tree)...)
		} else if strings.HasPrefix(item, "parent") {
			parents = append(parents, strings.TrimPrefix(item, "parent "))
		}
	}
	if len(parents) != 0 {
		for _, par := range parents {
			objs = append(objs, got.findCommitObjs(par)...)
		}
	}
	return objs
}

func (got *Got) missingObjs(localSha string, remoteSha string) []string {
	lo := got.findCommitObjs(localSha)
	if remoteSha == "" {
		return lo
	}
	ro := got.findCommitObjs(remoteSha)
	var ret []string
	ro_all := strings.Join(ro, "")
	for _, obj := range lo {
		if !strings.Contains(ro_all, obj) {
			ret = append(ret, obj)
		}
	}
	return ret
}

func (got *Got) encodePackObjects(sha1 string) []byte {
	_, ty, data, err := got.ReadObject(sha1)
	got.GotErr(err)
	var b bytes.Buffer
	err = compress(&b, data)
	if err != nil {
		got.GotErr(err)
	}
	data_compressed := b.Bytes()

	ty_num := 0
	switch ty {
	case "commit": ty_num = 1
	case "tree": ty_num = 2
	case "blob": ty_num = 3
	default: 
	}
	if ty_num == 0 {
		got.GotErr(errors.New("wrong boject type"))
	}
	size := len(data)
	by := (ty_num << 4) | (size & 0x0f)
	size >>= 4
	var ret bytes.Buffer
	for i:=size; i > 0; i++ {
		var b []byte
		binary.BigEndian.PutUint64(b, uint64((by | 0x80)))
		ret.Write(b)
		by = size & 0x7f
		size >>= 7
	}
	var buff []byte
	binary.BigEndian.PutUint64(buff, uint64(by))
	ret.Write(buff)
	ret.Write(data_compressed)
	return ret.Bytes()
}


func (got *Got) createPack(objs []string) []byte {
	//var b bytes.Buffer
	var b []byte
	b =  append(b, []byte("PACK")...)
	var buf []byte
	binary.BigEndian.PutUint32(buf, 2)
	b = append(b, buf...)
	binary.BigEndian.PutUint32(buf, uint32(len(objs)))
	b = append(b, buf...)
	sort.Slice(objs, func(i, j int) bool {return strings.Compare(objs[i], objs[j]) == -1})
	for _, obj := range objs {
		b = append(b, got.encodePackObjects(obj)...)
	} 
	sha1 := justhash(b)
	b = append(b, sha1...)
	return b

}