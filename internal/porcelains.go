package internal

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//we use sep in place of all zero-byte separators
var sep byte = 0
//Time format used in formatting commit time
const TIME_FORMAT = "Mon Jan 2 15:04:05 2006 -0700"

//#####PORCELAINS#####
//Porcelains are generally small and neat methods that rely heavily on plumbers to do the dirty work, and sometimes helprs too.
//Its an unfair world, but it is what it is
//Check: https://stackoverflow.com/questions/35894613/how-to-disallow-access-to-a-file-for-one-user/35895436#35895436 on file permissions
//Init creates a directory for your repo and initializes the hidden .git directory
func (got *Got) Init(name string) {

	if err := os.Mkdir(name, 0777); err != nil {
		got.GotErr(err)
	}
	//MkdirAll is just perfect because it creates directories on all the paths
	l := filepath.Join(name, ".git")
	fmt.Println(l)
	if err := os.MkdirAll(filepath.Join(name, ".git"), 0777); err != nil {
		got.GotErr(err)
	}
	if err := os.MkdirAll(filepath.Join(name, ".git", "objects"), 0777); err != nil {
		got.GotErr(err)
	}
	if err := os.MkdirAll(filepath.Join(name, ".git", "refs", "heads"), 0777); err != nil {
		got.GotErr(err)
	}
	//we create the HEAD file, a pointer to the current branch
	headPath := filepath.Join(name, ".git", "HEAD")
	if _, err := os.Create(headPath); err != nil {
		got.GotErr(err)
	}
	//init with the ref at  master
	got.writeToFile(headPath, []byte("ref: refs/heads/master"))
	got.logger.Printf("Initialized Empty Repository: %s \n", name)
}

func (got *Got) Status() {
	if is, _ := IsGit(); !is {
		got.logger.Fatalf("Not a valid git directory\n")
	}
	added, modified, deleted := got.get_status()
	if len(added) != 0 {
		_, err := fmt.Fprintf(got.logger.Writer(), "New files (untracked): %v\n To track run 'git add <filenale>'\n", added)
			got.GotErr(err)
	}
	if len(modified) != 0 {
		_, err := fmt.Fprintf(got.logger.Writer(), "Modified files: %v\n To stage changes 'git add <filenale>'\n", modified)
			got.GotErr(err)
	}
	if len(deleted) != 0 {
		_, err := fmt.Fprintf(got.logger.Writer(), "Deleted files: %v\n", deleted)
			got.GotErr(err)
	}
}


func (got *Got) Add(paths []string) {
	if is, _ := IsGit(); !is {
		got.logger.Fatalf("Not a valid git directory\n")
	}
	r := strings.NewReplacer("\\", "/")	
	for i, path := range paths {
		paths[i] = r.Replace(path)
	}
	indexes := got.readIndexFile()
	//Go, unlike Rust does not have iterators. Shit!
	var news []FPath
	//couldn't quickly of amuch better way of comparing these two
	var index_paths []string
	for _, ind := range indexes {
		index_paths = append(index_paths, string(ind.path))
	}
	index_paths_string := strings.Join(index_paths, " ")
	for _, p := range paths {
		if !strings.Contains(index_paths_string, p) {
			news = append(news, FPath(p))
		}
	}
	var new_inds []*Index
	for _, new := range news {
		ind := got.newIndex(string(new))
		new_inds = append(new_inds, ind)
	}
	if len(new_inds) != 0 {
		got.UpdateIndex(new_inds)
	} else {
		fmt.Fprintf(os.Stdout, "No file to be added")
	}
}

//to write a tree, we need to stage the files first, then from the indexed files, we write the tree
//TODO: for now, we support only root-level files
func (got *Got) WriteTree() string {
	if is, _ := IsGit(); !is {
		got.logger.Fatalf("Not a valid git directory\n")
	}
	// we need the mode, the path from root, and the sha1
	indexes := got.readIndexFile()
	var b bytes.Buffer
	for _, ind := range indexes {
		mode := binary.LittleEndian.Uint32(ind.mode[:])
		s := fmt.Sprintf("%o %s%v%x", mode, ind.path, sep, ind.sha1_obj_id)
		b.Write([]byte(s))
	}
	hash := got.HashObject(b.Bytes(), "tree")
	hash_s := hex.EncodeToString(hash)
	return hash_s
}

//https://github.com/git/git/blob/master/Documentation/technical/http-protocol.txt
//https://github.com/git/git/blob/master/Documentation/technical/pack-protocol.txt

func (got *Got) Commit(msg string) string {
	if is, _ := IsGit(); !is {
		got.logger.Fatalf("Not a valid git directory\n")
	}
	tree := got.WriteTree()
	parent := got.getLocalMAsterHash()
	uname, email, err := getConfig()
	if err != nil {
		got.logger.Fatalf("during commit: %s \n", err)
	}
	author := fmt.Sprintf("%s <%s>", uname, email)
	//TODO check if formatting is correct
	t := time.Now()
	commmit_time := t.Format(TIME_FORMAT)
	var s strings.Builder
	s.WriteString(fmt.Sprintf("tree %s\n", tree))
	s.WriteString(fmt.Sprintf("parent %s\n", parent))
	s.WriteString(fmt.Sprintf("author %s %s\n", author, commmit_time))
	s.WriteString(fmt.Sprintf("committer %s %s\n", author, commmit_time))
	s.WriteString(fmt.Sprintln())
	s.WriteString(msg)
	s.WriteString(fmt.Sprintln())
	sha1 := got.HashObject([]byte(s.String()), "commit")
	sha1_s := fmt.Sprintf("%x", sha1)
	path := filepath.Join(".git", "refs", "head", "master")
	//write the commit to refs/master
	f, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND, 0)
		got.GotErr(err)
	buf_f := bufio.NewWriter(f)
	_, err = buf_f.WriteString(sha1_s)
		got.GotErr(err)
	os.Stdout.WriteString("Commit succeded")
	return sha1_s
}

func (got *Got) Push(url string) {
if is, _ := IsGit(); !is {
		got.logger.Fatalf("Not a valid git directory\n")
	}
}
