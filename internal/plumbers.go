package internal

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/sha1"
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

//TODO:Check all the endianness in this code


func Config(conf ConfigObject) error {
	confRoot, err := os.UserConfigDir()
	if err != nil {
		return NotDefinedErr.addContext(err.Error())
	}
	err = os.Mkdir(filepath.Join(confRoot, ".got"), 0)
	if err != nil {
		return IOWriteErr.addContext(err.Error())
	}
	f, err := os.Create(filepath.Join(confRoot, ".got", ".config"))
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
func (git *Git) HashObject(data []byte, ty string, shouldwrite bool) []byte {
	base, err := os.Getwd()
	if err != nil {
		git.GotErr(err)
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
	//the byte result must be converted to hex string as that is how it is useful to us
	//we could either use fmt or hex.EncodeString here. Both works fine
	hash_str := fmt.Sprintf("%x", raw)
	//first two characters (1 byte) are the name of the directory. The remaining 38 (19 bytes) are the  name of the file
	//that contains the compressed version of the blob.
	//remember that sha1 produces a 20-byte hash (160 bits, or 40 hex characters)
	if shouldwrite {
		path := filepath.Join(base, ".git/objects/", hash_str[:2])
		err = os.MkdirAll(path, 0)
		git.GotErr(err)
		fPath := filepath.Join(path, hash_str[2:])
		f, err := os.Create(fPath)
		git.GotErr(err)
		defer f.Close()
		//the actual file is then compressed and stored in the file created
		err = compress(f, b)
		git.GotErr(err)
	}
	return raw
}

//FindObject takes a sha1 prefix. It returns the path to the object file. It doesn't care t open it 
//another method does that
//we want to have findObject such that even though we do not have
//a full 20 byte string, we can still find the path to the file
//it is unlikely to find two blobs with the same sha1 prefix, but if such happens, we return
//an error and expect the user to provide a longer string
func (git *Git) FindObject(prefix string) (string, error) {
	//first check for length
	switch l := len(prefix); {
	case l <= 2:
		return "", fmt.Errorf("the prefix provided is not sufficient for a search, ensure it's more than two")
	case l > 40:
		return "", fmt.Errorf("way too long. How did you come about a string longer than 40? darn it! This is sha-1")
	}
	path := filepath.Join(".git/objects", prefix[:2])
	entries, err := os.ReadDir(path)
	git.GotErr(err)
	location := ""
	//we also need to be sure that our given prefix is unique too. If it isn't we may be return ing the wrong file
	num := 0
	for _, entry := range entries {
		if entry.Type().IsRegular() && entry.Name() == prefix[:2] {
			num += 1
			location = entry.Name()
			if num > 1 {
				return "", fmt.Errorf("%s matches more than one blob", prefix)
			}
		}
	}
	if len(location) != 0 {
		return filepath.Join(".git/object", location), nil
	} else {
		return "", fmt.Errorf("%s matches no blob", prefix)
	}
}


//ReadObject bulds on findObject. First, it finds the object,
//but it does more, it attempts to read it and assert that it contains valid git object files
//apart from that, it tries to understand what kind of git object it is
func (git *Git) readObject(prefix string) (string, []byte) {
	location, err := git.FindObject(prefix)
	git.GotErr(err)
	f, err := os.Open(location)
	git.GotErr(err)
	defer f.Close()
	rdr, err := zlib.NewReader(f)
	git.GotErr(err)
	//create a buffer to hold the bytes to be read from zlib
	//buffer implements io.Writer
	var b bytes.Buffer
	//to decompress data from zlib, Go is marvelously helpful here, you only have to read from the zlib reader
	//I use copy here because it is cool and fast
	_, err = io.Copy(&b, rdr)
	git.GotErr(err)
	//remember that when writing the file, we put in a separator sep to demarcate the header from the body
	//sep is equal to byte(0)
	hdr, err := b.ReadBytes(sep)
	git.GotErr(err)
	//the remainder after reading out the header from the buffer, just like were flushing th buffer
	//we expect nothing else to be in the buffer right now other than the data itself
	data := b.Bytes()
	dType, dLen := "", 0
	_, err = fmt.Sscanf(string(hdr),"%s %d", &dType, &dLen)
	git.GotErr(err)
	if len(data) != int(dLen) {
		git.GotErr(fmt.Errorf("the data is corrupt, specified length foess not match length of data"))
	}
	return dType, data
}
//CatFile displays the file info on os.Stdout. It uses flags to determine what it displays
func (git *Git) CatFile(prefix, mode string) {
	dType, data := git.readObject(prefix)
	switch mode {
	case "size":
		git.logger.Printf("File Size: %d\n",len(data))
	case "type":
		git.logger.Printf("File Type: %s\n", dType)
	default:
		git.logger.Printf("File Content: \n%s", string(data))
	}
}

//write the index file, given a slice of index
//the gt version of updataindex
func (git *Git) UpdateIndex(entries []*Index) {

	var hdr []byte
	hdr = append(hdr, []byte("DIRC")...)
	//buf is apparently reusable
	var buf []byte
	//TODO: comeback here, this encoding is prolly wrong
	binary.BigEndian.PutUint32(buf, 2)
	hdr = append(hdr, buf[:4]...)
	//is this conversion safe?
	binary.BigEndian.PutUint32(buf, uint32(len(entries)))
	hdr = append(hdr, buf[:4]...)

	var data []byte
	for _, entry := range entries {
		data = append(data, entry.marshall()...)
	}
	allData := bytes.Join([][]byte{hdr, data}, nil)
	hash := sha1.New()
	hash.Write(allData)
	checksum := hash.Sum(nil)
	index := bytes.Join([][]byte{allData, checksum}, nil)
	err := git.writeToFile(filepath.Join(".git", "index"), index)
	if err != nil {
		git.GotErr(err)
	}
}

// read: https://mincong.io/2018/04/28/git-index/
//The index file contains:
// 12-byte header.
// A number of sorted index entries.
// Extensions. They are identified by signature.
// 160-bit SHA-1 over the content of the index file before this checksum.

func (git *Git) readIndexFile() []Index {
	f, err := os.Open(filepath.Join(".git/index"))
	data, err := io.ReadAll(f)
	if err != nil {
		git.GotErr(err)
	}
	hasher := sha1.New()
	//TODO check this boundary

	_, err = hasher.Write(data[:len(data)-20])
	if err != nil {
		git.GotErr(err)
	}
	hash := hasher.Sum(nil)
	//the index file has the lst 160 bits (i.e. 20 bytes) as the sha-1 checksum of all the bits tat come before it
	//we need to ensure that it matches before considering the data valid
	if bytes.Compare(hash, data[:(len(data)-20)]) != 0 {
		git.GotErr(errors.New("Checksum is not equal to file digest. File is corrupt"))
	}
	hdr := data[:12]
	//TODO: look for better ways to do this
	sign, version, numEntries := string(hdr[:4]), binary.BigEndian.Uint32(hdr[4:8]), binary.BigEndian.Uint32(hdr[8:])
	if strings.Compare(sign, "DIRC") != 0 {
		git.GotErr(fmt.Errorf("signature %s is not valid", sign))
	}
	if version != 2 {
		git.GotErr(fmt.Errorf("Version number must be 2, got %d", version))
	}
	//now for the index entries :
	//we need to use the unix fstat
	//the index files are listed between the 12-byte header and the 20-byte checksum
	indEntries := data[12:(len(data) - 20)]
	indexes := unmarshal(indEntries)
	if len(indexes) != int(numEntries) {
		git.GotErr(fmt.Errorf("Number of enteries does not equal to what the head specified"))
	}
	return indexes
}

func (git *Git) LsFiles(detail bool) {
	for _, ind := range git.readIndexFile() {
		path := string(ind.path)
		if detail {
			//stage number is the 3rd and 4th bit in the 16-bit flag
			//so we >> by 12 top pos the first three bits on the rightmost
			//our mask will be 11, i.e. 3, we & against 11 (or 0000000000000011) so that we'll keep only
			//the values of the last two bits intact, and cancel the third to the last
			stage := (binary.BigEndian.Uint16(ind.flags[:]) >> 12) & 3
			mode := binary.BigEndian.Uint32(ind.mode[:])
			sha1 := string(ind.sha1_obj_id[:])
			s := fmt.Sprintf("Mode: %o sha1: %x  stage: %d path: %s\n", mode, sha1, stage, path)
			_, err := os.Stdout.WriteString(s)
			git.GotErr(err)
		} else {
			_, err := os.Stdout.WriteString(path)
			git.GotErr(err)
		}
	}
}

func (git *Git) get_status() ([]string, []string, map[string]string) {
	var add, del []string
	mod := make(map[string]string)
	//we want to know the file that were changes, the ones that were deleted, and the ones that were added
	entries, err := os.ReadDir(".")
		git.GotErr(err)
	
	//remove .git
	for i, entry := range entries {
		if entry.IsDir() && !(entry.Name() == ".git") {
			entries = append(entries[:i], entries[i+1:]...)
		}
	}
	var files []string
	err = fs.WalkDir(os.DirFS("."), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fs.SkipDir
		}
		if !d.IsDir() {
			files = append(files, filepath.Clean(path))
		}
		return nil
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.Compare(files[i], files[j]) == -1
	})
	git.GotErr(err)
	
	index := git.readIndexFile()
	var index_map map[string]Index
	for _, ind := range index {
		index_map[string(ind.path)] = ind
	}
	//stored hash in index differs from the new hash
	modified := func(files []string) {
		for _, f_path := range files {
			if ind, ok := index_map[f_path]; ok {
				f, err := os.Open(f_path)
				git.GotErr(err)
				cont, err := io.ReadAll(f)
					git.GotErr(err)
			
				if hex.EncodeToString(git.HashObject(cont, "blob", true)) != string(ind.path) {
					mod[string(ind.path)] = f_path
				}
			}
		}
	}
	//no longer exists in the index file
	added := func(files []string) {
		for _, file := range files {
			if _, ok := index_map[file]; !ok {
				add = append(add, file)
			}
		}
	}
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
	//did not exist in the index file

	return add, del, mod
}
func (git *Git) status() {
	add, del, mod := git.get_status()
	var s strings.Builder
	if len(add) != 0 {
		s.WriteString(fmt.Sprintf("Added Files: \n%v\n", add))
	}
	if len(del) != 0 {
		s.WriteString(fmt.Sprintf("Deleted Files: \n%v\n", del))
	}
	s.WriteString("Modified files\n")
	for k, v := range mod {
		d := diff(k, v)
		s.WriteString(fmt.Sprintf("%s\n", d))
	}
	_, err := fmt.Fprintf(os.Stdout, "%s\n", s.String())
		git.GotErr(err)
	
}
func (git *Git) writeToFile(path string, b []byte) error {
	f, err := os.OpenFile(path, os.O_APPEND, 0)
	defer f.Close()
		git.GotErr(err)
	
	bufWriter := bufio.NewWriter(f)
	_, err = bufWriter.Write(b)

		git.GotErr(err)
	return err
}

func (git *Git) getLocalMAsterHash() string {
	path := filepath.Join(".git", "refs", "head", "master")
	f, err := os.Open(path)
		git.GotErr(err)
	
	//Hahaha. I always feel good anytime I'm able to explit Go's interface semantics
	var s strings.Builder
	_, err = io.Copy(&s, f)
		git.GotErr(err)
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

//TODO: Of course I need to read this from a file
func getEnvs() (name, mail string) {
	return "", ""
}


func (git *Git) readTree(sha string) ([]Object) {
	objs := make([]Object, 0)
	ty, data := git.readObject(sha)
	if ty != "tree" {
		git.GotErr("should be tree object")
	}
	if len(data) == 0 {
		git.GotErr("data is empty")
	}
	var path, sha1 string
	start := 0
	for {
		//TODO: is this the outher slice reassigned or just the slice i just finished working with in th last iter?
		data = data[start:]
		if len(data) == 0 {
			break
		}
		split := bytes.SplitN(data, []byte(" "), 2)
		md :=  string(split[0])
		mode, err := strconv.ParseInt(md, 8, 32)
		git.GotErr(err)
		sep_pos := bytes.IndexByte(split[1], byte(0))
		path = string(split[1][:sep_pos])
		sha1 = string(split[1][sep_pos+1: sep_pos+21])
		start += len(md) + len(path) + len(sha1)
		objs = append(objs, Object{mode, path, sha1})
	}
	
	return objs
}


func (git *Git) findTreeObjs(sha1 string) []string {
	var objs []string
	for _, obj := range git.readTree(sha1) {
		if fs.FileMode(obj.mode) == os.ModeDir {
			objs = append(objs, git.findTreeObjs(obj.sha1)...)
		} else {
			objs =append(objs, obj.sha1)
		}
	}

	return objs
}


func (git *Git) findCommitObjs(sha1 string) []string {
	var objs []string
	ty, data := git.readObject(sha1)
	if ty != "commit" {
		git.GotErr(fmt.Errorf("object is not a commit object"))
	}
	//one tree and 0 or more parents
	//find tree
	data_s := strings.Split(string(data), "\n")
	tree := ""
	parents := []string{}
	for _, item := range data_s {
		if strings.HasPrefix(item, "tree") {
			tree = strings.TrimPrefix(item, "tree ")
			objs = append(objs, git.findTreeObjs(tree)...)
		} else if strings.HasPrefix(item, "parent") {
			parents = append(parents, strings.TrimPrefix(item, "parent "))
		}
	}
	if len(parents) != 0 {
		for _, par := range parents {
			objs = append(objs, git.findCommitObjs(par)...)
		}
	}
	return objs
}

func (git *Git) missingObjs(localSha string, remoteSha string) []string {
	lo := git.findCommitObjs(localSha)
	if remoteSha == "" {
		return lo
	}
	ro := git.findCommitObjs(remoteSha)
	var ret []string
	ro_all := strings.Join(ro, "")
	for _, obj := range lo {
		if !strings.Contains(ro_all, obj) {
			ret = append(ret, obj)
		}
	}
	return ret
}

func (git *Git) encodePackObjects(sha1 string) []byte {
	ty, data := git.readObject(sha1)
	var b bytes.Buffer
	comp := zlib.NewWriter(&b)
	_, err := comp.Write(data)
	git.GotErr(err)
	data_compressed := b.Bytes()

	ty_num := 0
	switch ty {
	case "commit": ty_num = 1
	case "tree": ty_num = 2
	case "blob": ty_num = 3
	default: 
	}
	if ty_num == 0 {
		git.GotErr(errors.New("wrong boject type"))
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


func (git *Git) createPack(objs []string) []byte {
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
		b = append(b, git.encodePackObjects(obj)...)
	} 
	sha1 := justhash(b)
	b = append(b, sha1...)
	return b

}