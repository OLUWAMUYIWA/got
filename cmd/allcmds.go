package cmd

import "github.com/OLUWAMUYIWA/got/pkg"


type runner interface {
	run() error
}

type ini struct{
	wkdir string
}

func (i *ini) run() error {
	return  pkg.Init((*i).wkdir)
}


type add struct {
	addFlag bool
	args []string
}

func (a *add) run() error {
	got := pkg.NewGot()
	return got.Add(a.addFlag, a.args...)
}


type rm struct {
	paths []string
}

func (a *rm) run() error {
	got := pkg.NewGot()
	if err := got.Rm(a.paths...); err != nil {
			return err
	}
	return nil
}

type commit struct {
	msg string
} 

func (c *commit) run() error {
	got := pkg.NewGot()
	//comeback. we're doing nothing with the string returned here
	_, err := got.Commit((*c).msg)
	return err
}

type cat struct {
	prefix string
	mode int
}

func(c *cat) run() error {
	got := pkg.NewGot()
	return got.CatFile((*c).prefix, (*c).mode)
}