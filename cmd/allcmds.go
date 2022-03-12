package cmd

import (
	"fmt"

	"github.com/OLUWAMUYIWA/got/pkg"
)

// Runner handles the commands after they have been parsed
// the point of it is to allow a cleaner inerface for all git commands to implement
// this would help decouple the `pkg` package from our `cmd`s so that there in the implementations
// of each `Runner`, we can abstract the calls to `cmd` `api`s, do  dome pre-processing and post-processing
// without littering the `parseArgs` method any further
type Runner interface {
	Run() error
}

type initializer struct{
	wkdir string
}

func (i *initializer) Run() error {
	return  pkg.Init((*i).wkdir)
}


type add struct {
	addFlag bool
	args []string
}

func (a *add) Run() error {
	got := pkg.NewGot()
	if a.addFlag && len(a.args) != 0 {
		return fmt.Errorf("Error: 'A' flag is set but arguments were provided")
	}
	return got.Add(a.addFlag, a.args...)
}


type branch struct {
	name string
	delete bool
}

func (b *branch) Run() error {
	return nil
}


type cat struct {
	prefix string
	mode int
}

func(c *cat) Run() error {
	got := pkg.NewGot()
	return got.CatFile((*c).prefix, (*c).mode)
}



type commit struct {
	all bool
	msg string
} 

func (c *commit) Run() error {
	got := pkg.NewGot()
	if c.msg == "" {
		return fmt.Errorf("message should not be empty") 
	}
	//comeback. we're doing nothing with the string returned here
	_, err := got.Commit((*c).msg, c.all)
	return err
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
	cached bool
	output, arg string
}

func (d *diff) Run() error {
	got := pkg.NewGot()

	err := got.Diff(d.cached, d.output, d.arg)

	return err
}

type rm struct {
	paths []string
}

func (a *rm) Run() error {
	got := pkg.NewGot()
	if err := got.Rm(a.paths...); err != nil {
			return err
	}
	return nil
}


type lsFiles struct {
	lcached, ldeleted, lmodified, lothers bool
}

func (l *lsFiles) Run() error {
	got := pkg.NewGot()

	got.LsFiles()
}

type _switch struct {
	name string
	new bool
}

func (s *_switch) Run() error {
	return nil
}

