package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/OLUWAMUYIWA/got/internal"
)

func spec() {

}

func Exec() {
	init := flag.NewFlagSet("init", flag.ExitOnError)

	add := flag.NewFlagSet("add", flag.ExitOnError)
	cat := flag.NewFlagSet("cat-file", flag.ExitOnError)
	commit := flag.NewFlagSet("commit", flag.ExitOnError)
	config := flag.NewFlagSet("config", flag.ExitOnError)
	diff := flag.NewFlagSet("diff", flag.ExitOnError)
	hashObj := flag.NewFlagSet("hash-object", flag.ExitOnError)
	push := flag.NewFlagSet("push", flag.ExitOnError)
	status := flag.NewFlagSet("status", flag.ExitOnError)
	updInd := flag.NewFlagSet("update-index", flag.ExitOnError)
	pull := flag.NewFlagSet("pull", flag.ExitOnError)

	flag.Parse()
	args := flag.Args()
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "No subcommand provided for git to work with")
		os.Exit(1)
	}

	switch args[1] {
	case "init":
		init.Parse(args[2:])
	case "add":
		add.Parse(args[2:])
	case "cat-file":
		cat.Parse(args[2:])
	case "commit":
		commit.Parse(args[2:])
	case "config":
		config.Parse(args[2:])
	case "diff":
		diff.Parse(args[2:])
	case "hash-object":
		hashObj.Parse(args[2:])
	case "push":
		push.Parse(args[2:])
	case "status":
		status.Parse(args[2:])
	case "update-index":
		updInd.Parse(args[2:])
	case "pull":
		pull.Parse(args[2:])
	}
	var got *internal.Got

	//now we have togo through each of the subcommands to know the one that was passed. we then execute our program logic
	if init.Parsed() {
		if len(args) < 3 {
			fmt.Fprintf(os.Stderr, "Init needs  the directorsy specified as an argument. \n. If this is the wkdir, put .")
			os.Exit(1)
		}
		wkdir := args[2]
		got = internal.NewGot(wkdir)
		got.Init(got.WkDir())
	} else { //workspace has already been initialized and its wkdir has been saved in our sonfig file
		got = internal.NewGot(internal.Wkdir())
	}
}
