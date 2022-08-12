package cmd

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/OLUWAMUYIWA/got/pkg"
)

//comeback: handle context propagation in all runners

// Runner handles the commands after they have been parsed
// the point of it is to allow a cleaner inerface for all git commands to implement
// this would help decouple the `pkg` package from our `cmd`s so that there in the implementations
// of each `Runner`, we can abstract the calls to `cmd` `api`s, do  dome pre-processing and post-processing
// without littering the `parseArgs` method any further
type Runner interface {
	Run(ctx context.Context) error
}

type initializer struct {
	wkdir string
}

func (i *initializer) Run(ctx context.Context) error {
	return pkg.Init(ctx, (*i).wkdir)
}

type add struct {
	addFlag bool
	args    []string
}

func (a *add) Run(ctx context.Context) error {
	got := pkg.NewGot()
	if a.addFlag && len(a.args) != 0 {
		return fmt.Errorf("Error: 'A' flag is set but arguments were provided")
	}
	return got.Add(ctx, a.addFlag, a.args...)
}

type branch struct {
	name   string
	delete bool
}

func (b *branch) Run(ctx context.Context) error {
	got := pkg.NewGot()
	if b.name == "" {
		rdr, err := got.Branches()
		if err != nil {
			return err
		}
		_, err = io.Copy(os.Stdout, rdr)
		return err
	}

	if b.delete {
		err := got.DeleteBranch(ctx, b.name)
		return err
	}

	if err := got.NewBranch(ctx, b.name); err != nil {
		return err
	}

	return nil
}

type catType int

const (
	s catType = iota
	t
	p
)

// Provide content or type and size information for repository objects
type cat struct {
	prefix              string
	size, _type, pretty bool
	mode                catType
}

func (c *cat) valid() bool {
	if !c.size && !c._type && !c.pretty {
		c.pretty = true
		return true
	}

	if c.size && !c._type && !c.pretty {
		return true
	}

	if !c.size && c._type && !c.pretty {
		return true
	}

	if !c.size && !c._type && c.pretty {
		return true
	}

	return false

}
func (c *cat) Run(ctx context.Context) error {
	if !c.valid() {
		return fmt.Errorf("Problem validating arguments, more than one of -t, -s, -p is set, or none is")
	}
	got := pkg.NewGot()
	if c.size {
		c.mode = s
	} else if c._type {
		c.mode = t
	} else {
		c.mode = p
	}

	rdr, err := got.CatFile(ctx, (*c).prefix, int((*c).mode))
	if err == nil {
		err := got.Log(rdr)
		return err
	}
	return err
}

type checkout struct {
	name string
	new  bool
}

func (c *checkout) Run(ctx context.Context) error {
	got := pkg.NewGot()
	if c.new {
		err := got.NewBranch(ctx, c.name)
		if err != nil {
			return err
		}
	}
	if err := got.Checkout(ctx, c.name); err != nil {
		return err
	}
	return nil
}

type commit struct {
	all bool
	msg string
}

func (c *commit) Run(ctx context.Context) error {
	got := pkg.NewGot()
	if c.msg == "" {
		fPath := filepath.Join(os.TempDir(), "msg.txt")
		f, err := os.OpenFile(fPath, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return err
		}
		defer f.Close()
		cmd := exec.Command(os.Getenv("EDITOR"), fPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Start()
		if err != nil {
			return err
		}
		if err := cmd.Wait(); err != nil {
			return err
		}
		msg, err := io.ReadAll(f)
		if len(msg) == 0 {
			return fmt.Errorf("You wrote no message")
		}
		c.msg = string(msg)

	}
	//comeback. we're doing nothing with the string returned here
	_, err := got.Commit(ctx, (*c).msg, c.all)
	return err
}

type config struct {
	path                  []string
	value                 string
	local, global, system bool
	rmd                   int // read = 0, add/modify = 1, delete = 2
}

func (c *config) validate() (int, error) {
	var where int
	if c.global {
		where = 1
	} else if c.system {
		where = 2
	} else {
		where = 0
	}
	if (c.local && c.global) || (c.local && c.system) || (c.global && c.system) {
		return where, fmt.Errorf("only one of these options meay be set")
	}
	return where, nil
}

func (c *config) Run(ctx context.Context) error {
	got := pkg.NewGot()
	where, err := c.validate()
	if err != nil {
		return err
	}
	if c.rmd == 0 { // read
		rdr, err := got.ShowConf(c.path, where)
		if err != nil {
			return err
		}
		_, err = io.Copy(os.Stdout, rdr)
		return err
	} else if c.rmd == 1 { // modify or add
		// either add or modify property
		return got.UpdateConf(c.path, c.value, where)
	} else if c.rmd == 2 {
		return got.Delete(c.path, where)
	} else {
		return fmt.Errorf("Invalid aruments")
	}
}

// Show changes between the working tree and the index or a tree, changes between the index and a tree,
//changes between two trees,
// changes resulting from a merge, changes between two blob objects, or changes between two files on disk.

// we limit our diff command here. if an arg is provided, we find the diff between the current working tree
// and the last commit (not the index).
// if cached is set, it means we expect no argument, and we wish to find the diff between the current state
// of the index file and the files being tracked by it in the working tree. otherwise, we'll be diffing for
// files that are not yet being tracked
type diff struct {
	cached      bool
	output, arg string
}

func (d *diff) Run(ctx context.Context) error {
	got := pkg.NewGot()

	err := got.Diff(d.cached, d.output, d.arg)

	return err
}

// Fetch branches and/or tags (collectively, "refs") from one or more other repositories,
// along with the objects necessary to complete their histories.
type fetch struct {
	remote string
}

func (f *fetch) Run(ctx context.Context) error {
	got := pkg.NewGot()
	return got.Fetch(ctx, f.remote)
}

// git-hash-object - Compute object ID and optionally creates a blob from a file
type hashObj struct {
	_type string
	w     bool
	file  string
}

// comeback to check valid objects
func (h *hashObj) validate() error {
	if h._type == "" {
		h._type = "blob"
	}
	if h._type == "blob" || h._type == "tree" || h._type == "commit" {
		return nil
	}
	return fmt.Errorf("Not a valid git obect type")
}

func (h *hashObj) Run(ctx context.Context) error {
	got := pkg.NewGot()
	b, err := ioutil.ReadFile(h.file)
	if err != nil {
		return err
	}
	hash, err := got.HashObject(b, h._type, h.w)
	if err != nil {
		return err
	}
	_, err = io.WriteString(os.Stdout, hex.EncodeToString(hash[:]))
	return err
}

// git-ls-files - Show information about files in the index and the working tree

type lsFiles struct {
	lstaged, lcached, ldeleted, lmodified, lothers bool
}

func (l *lsFiles) Run(ctx context.Context) error {
	got := pkg.NewGot()
	err := got.LsFiles(ctx, l.lstaged, l.lcached, l.ldeleted, l.lmodified, l.lothers)
	return err
}

// // Lists the contents of a given tree object, like what "ls -a" does in the current working directory.
// // path is relative to the current working directory
// // output format: <mode> SP <type> SP <object> TAB <file>
// type lsTree struct {
// 	path string
// }

// func (l *lsTree) Run(ctx context.Context) error {
// 	got := pkg.NewGot()
// 	rdr, err := got.LsTree(ctx, l.path)
// 	io.Copy(os.Stdout, rdr)
// 	return err
// }

type merge struct {
	comm string
}

func (m *merge) Run(ctx context.Context) error {
	got := pkg.NewGot()
	return got.Merge(ctx, m.comm)
}

type pull struct {
	remote string
	rebase bool
}

// Incorporates changes from a remote repository into the current branch.
// If the current branch is behind the remote, then by default it will fast-forward the current branch to match the remote. If the current branch and the remote have diverged,
// the user needs to specify how to reconcile the divergent branches with --rebase
func (p *pull) Run(ctx context.Context) error {
	got := pkg.NewGot()
	return got.Pull(ctx, p.remote, p.rebase)
}

// git-push - Update remote refs along with associated objects
type push struct {
	repo string
}

func (p *push) Run(ctx context.Context) error {
	got := pkg.NewGot()
	s, err := got.Push(ctx, p.repo)
	if err != nil {
		return err
	}
	_, err = os.Stdout.WriteString(s)
	return err
}

// // git-read-tree - Reads tree information into the index
// type readTree struct {
// 	treeish string
// }

// func (r *readTree) Run(ctx context.Context) error {
// 	got := pkg.NewGot()
// 	return got.ReadTree(r.treeish)
// }

type remote struct {
	name  string
	_type int
}

func (r *remote) Run(ctx context.Context) error {
	got := pkg.NewGot()
	if r._type == 0 {
		return got.RemoteAdd(ctx, r.name)
	} else if r._type == 1 {
		return got.RemoteRm(ctx, r.name)
	} else {
		return fmt.Errorf("invalid subcommand for remote command")
	}
}

type rm struct {
	cached bool
	paths  []string
}

func (a *rm) Run(ctx context.Context) error {
	got := pkg.NewGot()
	if err := got.Rm(ctx, a.cached, a.paths); err != nil {
		return err
	}
	return nil
}

// Displays paths that have differences between the index file and the current HEAD commit,
// paths that have differences between the working tree and the index file, and paths in the working tree that are not tracked by Git
type status struct{}

func (s *status) Run(ctx context.Context) error {
	got := pkg.NewGot()
	rdr := got.Status(ctx)
	_, err := io.Copy(os.Stdout, rdr)
	return err
}

// Switch to a specified branch. The working tree and the index are updated to match the branch.
// All new commits will be added to the tip of this branch.
type _switch struct {
	name string
	new  bool
}

func (s *_switch) Run(ctx context.Context) error {
	got := pkg.NewGot()
	if s.new {
		err := got.NewBranch(ctx, s.name)
		if err != nil {
			return err
		}
	}
	if err := got.Checkout(ctx, s.name); err != nil {
		return err
	}
	return nil
}

// git-update-index - Register file contents in the working tree to the index
type updateIndex struct {
	add, remove bool
}

func (u *updateIndex) Run(ctx context.Context) error {
	got := pkg.NewGot()
	return got.UpdateIndex(ctx, u.add, u.remove)
}

// git-verify-pack - Validate packed Git archive files
// Reads given idx file for packed Git archive created with the git pack-objects command and verifies idx file and the corresponding pack file.
type verifyPack struct {
	idx string
}

// comeback
func (v *verifyPack) Run(ctx context.Context) error {
	got := pkg.NewGot()
	return got.VerifyPack(ctx, v.idx)
}

// git-write-tree - Create a tree object from the current index
// Creates a tree object using the current index. The name of the new tree object is printed to standard output.
type writeTree struct{}

func (w *writeTree) Run(ctx context.Context) error {
	got := pkg.NewGot()
	s, err := got.WriteTree(ctx)
	if err != nil {
		return err
	}
	_, err = io.Copy(os.Stdout, strings.NewReader(s))
	return err
}
