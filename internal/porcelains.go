package internal

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/OLUWAMUYIWA/got/internal/proto"
)

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
	if len(added) != 0 || len(modified) != 0 || len(deleted) != 0 {
		got.logger.Printf("Changes to be committed:\n\n")
	}
	if len(modified) != 0 {
		_, err := fmt.Fprintf(got.logger.Writer(), "Modified files: %v\n To stage changes 'git add <filenale>'\n", modified)
		got.GotErr(err)
	}
	if len(deleted) != 0 {
		_, err := fmt.Fprintf(got.logger.Writer(), "Deleted files: %v\n", deleted)
		got.GotErr(err)
	}

	if len(added) != 0 {
		_, err := fmt.Fprintf(got.logger.Writer(), "Untracked files. Use `git add` to start tracking: %v\n To track run 'git add <filenale>'\n", added)
		got.GotErr(err)
	}
}

//To add files to the staging area/cache,
//first read the index file, and compare the paths of its contents to the set of new paths you want to add
func (got *Got) Add(paths []string) {
	if is, _ := IsGit(); !is {
		got.logger.Fatalf("Not a valid git directory\n")
	}
	indexes, err := readIndexFile(got)
	if err != nil {
		got.FatalErr(err)
	}
	//Go, unlike Rust does not have iterators. Shit!
	var news []string
	//couldn't quicklyfind a much better way of comparing these two
	var index_paths []string
	for _, ind := range indexes {
		index_paths = append(index_paths, string(ind.path))
	}
	index_paths_string := strings.Join(index_paths, " ")
	for _, p := range paths {
		if !strings.Contains(index_paths_string, p) {
			news = append(news, p)
		}
	}
	var new_inds []*Index
	for _, new := range news {
		ind := got.newIndex(string(new))
		new_inds = append(new_inds, ind)
	}

	if len(new_inds) != 0 {
		//sort the indexes by pathname
		sort.Slice(new_inds, func(i, j int) bool {
			return string(new_inds[i].path) < string(new_inds[j].path)
		})
		got.UpdateIndex(new_inds)
	} else {
		fmt.Fprintf(os.Stdout, "No file to be added")
	}
}

//https://github.com/git/git/blob/master/Documentation/technical/http-protocol.txt
//https://github.com/git/git/blob/master/Documentation/technical/pack-protocol.txt
//Commit first writes the tree from the set of staged objects
//we have no  commit-tree method
func (got *Got) Commit(msg string) (string, error) {
	if is, _ := IsGit(); !is {
		got.logger.Fatalf("Not a valid git directory\n")
	}
	//write the tree object first
	tree, err := got.WriteTree()
	if err != nil {
		return "", fmt.Errorf("Could not commit because: %w", err)
	}
	//get the parent sha from the HEAD file
	parent, err := got.head.ReadCont()
	if err != nil {
		return "", fmt.Errorf("Commit error: %w", err)
	}
	//get uname and email from configuration file
	uname, email, err := getConfig()
	if err != nil {
		got.logger.Fatalf("during commit: %s \n", err)
	}
	author := fmt.Sprintf("%s <%s>", uname, email)
	//TODO check if formatting is correct
	//format the time
	t := time.Now()
	commmit_time := t.Format(TIME_FORMAT)
	var s strings.Builder
	//write it in a commit format. specified in the progit book
	s.WriteString(fmt.Sprintf("tree %s\n", tree))
	s.WriteString(fmt.Sprintf("parent %s\n", parent))
	s.WriteString(fmt.Sprintf("author %s %s\n", author, commmit_time))
	s.WriteString(fmt.Sprintf("committer %s %s\n", author, commmit_time))
	s.WriteString(fmt.Sprintln())
	s.WriteString(msg)
	s.WriteString(fmt.Sprintln())
	//write the commit object
	commit := &commitObj{data: []byte(s.String())}
	_, err = commit.Hash(got.baseDir)
	if err != nil {
		//TODO: handle error
	}
	path := filepath.Join(".git", "refs", "head", "master")
	//write the commit to refs/head/master. replace, no append
	//this becomes the latest commit in the master branch.
	//the refs/head/master is a symbolic to the
	f, err := os.OpenFile(path, os.O_RDWR, 0777)
	got.GotErr(err)
	buf_f := bufio.NewWriter(f)
	_, err = buf_f.WriteString(commit.sha)
	got.GotErr(err)
	os.Stdout.WriteString("Commit succeded")
	return commit.sha, nil
}

//after refs and capabilities discovery, client may send the flush packet to tell the server it has ended
func (got *Got) LsRemote() {

}

//discover references available in the remote repo first
//the remote repo has no workspace. It basically contains what is in the .git directory.
func (got *Got) Push(url string) {
	if is, _ := IsGit(); !is {
		got.logger.Fatalf("Not a valid git directory\n")
	}
	uname, passwd, err := getConfig()
	if err != nil {
		fmt.Println("supply a space-separated username and password")
		_, err = fmt.Scanln(&uname, &passwd)
		if err != nil {
			got.logger.Fatalln("error reading username and password")
		}
	}
	localSha, err := got.head.ReadCont()
	remoteSha, err := proto.GetRemoteMasterHash(url, uname, passwd)
	got.GotErr(err)
	missings := got.missingObjs(string(localSha), remoteSha)
	//TODO: inform the user
	var s strings.Builder
	s.WriteString(fmt.Sprintf("%s %s refs/heads/master%x report-status", remoteSha, localSha, 0))
	//remove this
	fmt.Println(missings)
}
