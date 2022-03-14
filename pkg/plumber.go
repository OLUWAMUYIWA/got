package pkg

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
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

//FindObject takes a sha1 prefix. It returns the path to the object file. It doesn't care t open it
//another method does that
//we want to have findObject such that even though we do not have
//a full 20 byte string, we can still find the path to the file
//it is unlikely to find two blobs with the same sha1 prefix, but if such happens, we return
//an error and expect the user to provide a longer string
func (got *Got) FindObject(prefix string) (string, error) {
	
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
		if entry.Type().IsRegular() && strings.HasPrefix(entry.Name(), prefix[2:]) {
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
	decompress(f, &b)
	if err != nil {
		return "", "", nil, err
	}
	//remember that when writing the file, we put in a separator sep to demarcate the header from the body
	//sep is equal to byte(0)
	hdr, err := b.ReadBytes(Sep)
	if err != nil {
		return "", "", nil, err
	}
	//the remainder after reading out the header from the buffer, just like were flushing th buffer
	//we expect nothing else to be in the buffer right now other than the data itself
	data := b.Bytes()
	dType, dLen := "", 0
	_, err = fmt.Sscanf(string(hdr), "%s %d", &dType, &dLen)
	if err != nil {
		return "", "", nil, err
	}
	if len(data) != int(dLen) {
		return "", "", nil, fmt.Errorf("the data is corrupt, specified length does not match length of data")
	}
	return f_name, dType, data, nil
}

//CatFile displays the file info using the git logger (set as os.Stdout). It uses flags to determine what it displays
func (got *Got) CatFile(prefix string, mode int) (io.Reader, error){
	
	f_name, dType, data, err := got.ReadObject(prefix)
	//this error should just cause the program to exit.
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	switch mode {
	case 0: //size
		_, err := io.WriteString(&b, fmt.Sprintf("File %s: Size: %d\n", f_name, len(data)))
		return &b, err
	case 1: //type
		_, err := io.WriteString(&b, fmt.Sprintf("File %s Type: %s\n", f_name, dType))
		return &b, err
	case 2: //pretty
		if dType == "commit" || dType == "blob" {
			_, err := io.WriteString(&b, fmt.Sprintf("Content %s: \n%s", f_name, string(data)))
			return &b, err
		} else if dType == "tree" {
			//is a directory. ewe need to read th tree before printing
			_, err := io.WriteString(&b, fmt.Sprintf("Directory: \n"))
			if err != nil {
				return nil, err
			}
			objs := got.deserTree(prefix)
			for _, obj := range objs {
				rdr, err := got.CatFile(obj.sha1, 2)
				if err != nil {
					return nil, err
				}
				io.Copy(&b, rdr)
			}
			return &b, nil
		}
	default:
		_, err := io.WriteString(&b, fmt.Sprintf("Bad flag mode. Check again, must be either size, type, pretty \n"))
		return &b, err
	}

}

//LsFiles prints to stdOut the state of staged files, i.e. the index files
//After a Commit, it is clean
// comeback
func (got *Got) LsFiles(stage, cached, deleted, modified, others bool ) error {

	indexes, err := readIndexFile(got)
	if err != nil {
		return err
	}

	if len(indexes) != 0 {
		got.logger.Printf("No staged files\n")
		return fmt.Errorf("No files staged")
	}

	for _, ind := range indexes {
		path := string(ind.path)
		if stage {
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
			got.logger.Printf("%s\n", path)
		}
	}

	return nil
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
	//TODO check if DrFS should be this `.`
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
	fmt.Println(files)
	//sort the file paths
	sort.Slice(files, func(i, j int) bool {
		return files[i] < files[j]
	})
	//from the index we know files that are currently staged
	index, err := readIndexFile(got)
	if err != nil {
		got.FatalErr(err)
	}
	var index_map map[string]Index
	for _, ind := range index {
		index_map[string(ind.path)] = ind
	}
	//stored hash in index differs from the new hash, these ones have been modified. They are being tracked
	modified := func(files []string) {
		for _, f_path := range files {
			if ind, ok := index_map[f_path]; ok {
				f, err := os.Open(f_path)
				got.GotErr(err)
				cont, err := io.ReadAll(f)
				got.GotErr(err)
				raw, err := hashWithObjFormat(cont, "blob")
				if err != nil {
					return
				}
				if hex.EncodeToString(raw) != hex.EncodeToString(ind.sha1_obj_id[:]) {
					mod[string(ind.path)] = f_path
				}
			}
		}
	}

	//not being tracked yet, since they haven't been staged
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
		d, err := diff([]byte(k), []byte(v))
		if err != nil {
			break
		}
		s.WriteString(fmt.Sprintf("%s\n", d))
	}
	_, err := fmt.Fprintf(os.Stdout, "%s\n", s.String())
	got.GotErr(err)

}

//to write a tree, we need to stage the files first i.e. index them, then from the indexed files, we write the tree
//TODO: for now, we support only root-level files
//WriteTree just takes the current values in the index (i.e. the staged files) and writes as tree object
func (got *Got) WriteTree() (string, error) {
	// we need the mode, the path from root, and the sha1
	indexes, err := readIndexFile(got)
	if err != nil {
		got.FatalErr(err)
	}
	if len(indexes) == 0 {
		return "", fmt.Errorf("No files staged \n")
	}
	var b bytes.Buffer
	for _, ind := range indexes {
		mode := binary.BigEndian.Uint32(ind.mode[:])
		s := fmt.Sprintf("%o %s%v%v", mode, ind.path, Sep, ind.sha1_obj_id)
		b.Write([]byte(s))
	}
	tree := Tree{
		data: b.Bytes(),
	}
	hash, err := tree.Hash(got.baseDir)
	if err != nil {
		return "", fmt.Errorf("Writing Tree: %w", err)
	}
	hash_s := hex.EncodeToString(hash)
	return hash_s, nil
}

//return all the objects (subtrees and blobs) inside a tree
func (got *Got) deserTree(sha string) []treeItem {
	if is, _ := IsGit(); !is {
		got.logger.Fatalf("Not a valid git directory\n")
	}
	_, ty, data, err := got.ReadObject(sha)
	got.GotErr(err)
	if ty != "tree" {
		got.FatalErr(fmt.Errorf("should be tree object"))
	}
	if len(data) == 0 {
		got.FatalErr(fmt.Errorf("should be tree object"))
	}
	objs := make([]treeItem, 0)
	var path, sha1 string
	start := 0
	for {
		d := data[start:]
		if len(data) == 0 {
			break
		}
		split := bytes.SplitN(d, []byte(" "), 2)
		md := string(split[0])
		mode, err := strconv.ParseUint(md, 8, 32)
		got.GotErr(err)
		sep_pos := bytes.IndexByte(split[1], Sep)
		path = string(split[1][:sep_pos])
		sha1 = string(split[1][sep_pos+1 : sep_pos+21])
		start += len(split[0]) + 1 + sep_pos + 1 + 20
		objs = append(objs, treeItem{mode, path, sha1})
	}

	return objs
}

//TODO: readTree() -> read the contents of a tree into the staging area

func (got *Got) findTreeObjs(sha1 string) []string {
	var objs []string
	for _, obj := range got.deserTree(sha1) {
		if fs.FileMode(obj.mode) == os.ModeDir {
			objs = append(objs, got.findTreeObjs(obj.sha1)...)
		} else {
			objs = append(objs, obj.sha1)
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

//TODO
//func (got *Got) Log() {}
//output the commit info starting from the latest commit
