package cmd

import (
	"context"
	"fmt"

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

type initializer struct{
	wkdir string
}

func (i *initializer) Run(ctx context.Context) error {
	return  pkg.Init((*i).wkdir)
}


type add struct {
	addFlag bool
	args []string
}

func (a *add) Run(ctx context.Context) error {
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

func (b *branch) Run(ctx context.Context) error {
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
	prefix string
	size, _type, pretty bool
	mode catType
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
func(c *cat) Run(ctx context.Context) error {
	if !c.valid() {
		return fmt.Errorf("Problem validating arguments, more than one of -t, -s, -p is set, or none is")
	}
	got := pkg.NewGot()
	if c.size  {
		c.mode = s
	} else if c._type  {
		c.mode = t
	} else {
		c.mode = p
	}

	rdr, err := got.CatFile((*c).prefix, int((*c).mode))
	if err == nil {
		err := got.Log(rdr)
		return err
	}
	return err
}



type commit struct {
	all bool
	msg string
} 

func (c *commit) Run(ctx context.Context) error {
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

func (d *diff) Run(ctx context.Context) error {
	got := pkg.NewGot()

	err := got.Diff(d.cached, d.output, d.arg)

	return err
}

type rm struct {
	cached bool
	paths []string
}

func (a *rm) Run(ctx context.Context) error {
	got := pkg.NewGot()
	if err := got.Rm(a.cached, a.paths); err != nil {
			return err
	}
	return nil
}


type lsFiles struct {
	lstaged, lcached, ldeleted, lmodified, lothers bool
}

func (l *lsFiles) Run(ctx context.Context) error {
	got := pkg.NewGot()
	err := got.LsFiles(l.lstaged, l.lcached, l.ldeleted, l.lmodified, l.lothers)
	return err
}

type _switch struct {
	name string
	new bool
}

func (s *_switch) Run(ctx context.Context) error {
	return nil
}

