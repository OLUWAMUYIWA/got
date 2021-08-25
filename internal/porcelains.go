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

//TODO: use sep in place of all zero-yte separators
var sep byte = 0

const TIME_FORMAT = "Mon Jan 2 15:04:05 2006 -0700"

//#####PORCELAINS#####
//Porcelains are generally small and neat methods that rely heavily on plumbers to do the dirty work.
//Its an unfair world, but it is what it is

//Init creates a directory for your repo and initializes the hidden .git directory
func (git *Git) Init(name string) {

	if err := os.Mkdir(name, 0); err != nil {
		git.GotErr(err)
	}
	//MkdirAll is just perfect because it creates directories on all the paths
	if err := os.MkdirAll(filepath.Join(name, ".git"), 0); err != nil {
		git.GotErr(err)
	}
	if err := os.MkdirAll(filepath.Join(name, "objects"), 0); err != nil {
		git.GotErr(err)
	}
	if err := os.MkdirAll(filepath.Join(name, "refs", "heads"), 0); err != nil {
		git.GotErr(err)
	}
	//we create the HEAD file, a pointer to the current branch
	headPath := filepath.Join(name, ".git", "HEAD")
	if _, err := os.Create(headPath); err != nil {
		git.GotErr(err)
	}
	//init with te ref at  master
	git.writeToFile(headPath, []byte("ref: refs/heads/master"))
	git.logger.Printf("Initialized Empty Repository: %s \n", name)
}

func (git *Git) Status() {
	added, modified, deleted := git.get_status()
	out := os.Stdout
	if len(added) != 0 {
		_, err := fmt.Fprintf(out, "New files (untracked): %v\n To track run 'git add <filenale>'\n", added)
			git.GotErr(err)
	}
	if len(modified) != 0 {
		_, err := fmt.Fprintf(out, "Modified files: %v\n To stage changes 'git add <filenale>'\n", modified)
			git.GotErr(err)
	}
	if len(deleted) != 0 {
		_, err := fmt.Fprintf(out, "Deleted files: %v\n", deleted)
			git.GotErr(err)
	}
}

//i must use a library to pretty_print/highlight the changed parts of files here
//should I write a simple
func (git *Git) Diff() {

}

func (git *Git) Add(paths []string) {
	r := strings.NewReplacer("\\", "/")
	for i, path := range paths {
		paths[i] = r.Replace(path)
	}
	indexes := git.readIndexFile()
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
		ind := git.newIndex(string(new))
		new_inds = append(new_inds, ind)
	}
	if len(new_inds) != 0 {
		git.UpdateIndex(new_inds)
	} else {
		fmt.Fprintf(os.Stdout, "No file to be added")
	}
}

//to write a tree, we need to stage the files first, then from the indexed files, we write the tree
//TODO: for now, we support only root-level files
func (git *Git) WriteTree() string {
	// we need the mode, the path from root, and the sha1
	indexes := git.readIndexFile()
	var b bytes.Buffer
	for _, ind := range indexes {
		mode := binary.LittleEndian.Uint32(ind.mode[:])
		s := fmt.Sprintf("%o %s%v%x", mode, ind.path, sep, ind.sha1_obj_id)
		b.Write([]byte(s))
	}
	hash := git.HashObject(b.Bytes(), "tree", true)
	hash_s := hex.EncodeToString(hash)
	return hash_s
}

//https://github.com/git/git/blob/master/Documentation/technical/http-protocol.txt
//https://github.com/git/git/blob/master/Documentation/technical/pack-protocol.txt

func (git *Git) Commit(msg string) string {
	tree := git.WriteTree()
	parent := git.getLocalMAsterHash()
	uname, email := getEnvs()
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
	sha1 := git.HashObject([]byte(s.String()), "commit", true)
	sha1_s := fmt.Sprintf("%x", sha1)
	path := filepath.Join(".git", "refs", "head", "master")
	//write the commit to refs/master
	f, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND, 0)
		git.GotErr(err)
	buf_f := bufio.NewWriter(f)
	_, err = buf_f.WriteString(sha1_s)
		git.GotErr(err)
	os.Stdout.WriteString("Commit succeded")
	return sha1_s
}

func (git *Git) Push(url string) {

}
