package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/OLUWAMUYIWA/got/pkg"
)

type app struct {

}

func (a app) parseArgs() {
	

}

func (a app) run() {

}

//comeback handle exit codes
func Exec() int {


	//each subcommand has its own flagset
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)


	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	var addFlag bool
	addCmd.BoolVar(&addFlag, "all", false, "Specify that all files starting from root directory should de added")
	addCmd.BoolVar(&addFlag, "A", false, "Specify that all files starting from root directory should de added (shorthand)")

	catCmd := flag.NewFlagSet("cat-file", flag.ExitOnError)
	

	commitCmd := flag.NewFlagSet("commit", flag.ExitOnError)
	

	configCmd := flag.NewFlagSet("config", flag.ExitOnError)
	

	diffCmd := flag.NewFlagSet("diff", flag.ExitOnError)
	

	hashObjCmd := flag.NewFlagSet("hash-object", flag.ExitOnError)
	

	fetchCmd := flag.NewFlagSet("fetch", flag.ExitOnError)
	

	lsFilesCmd := flag.NewFlagSet("ls-files", flag.ExitOnError)
	

	pushCmd := flag.NewFlagSet("push", flag.ExitOnError)
	

	statusCmd := flag.NewFlagSet("status", flag.ExitOnError)
	

	updIndCmd := flag.NewFlagSet("update-index", flag.ExitOnError)
	

	pullCmd := flag.NewFlagSet("pull", flag.ExitOnError)

	flag.Parse()

	args := flag.Args()
	
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "No subcommand provided for git to work with")
		return 1
	}


	//parse each of the subcommands, starting from the second argument.
	//remember that args[0] will be the program binary name
	//args[1] is our subcommand.
	switch args[1] {
	case "init":
		initCmd.Parse(args[2:])
	case "add":
		addCmd.Parse(args[2:])
	case "cat-file":
		catCmd.Parse(args[2:])
	case "commit":
		commitCmd.Parse(args[2:])
	case "config":
		configCmd.Parse(args[2:])
	case "diff":
		diffCmd.Parse(args[2:])
	case "hash-object":
		hashObjCmd.Parse(args[2:])
	case "fetch":
		fetchCmd.Parse(args[2:])
	case "ls-files":
		lsFilesCmd.Parse(args[2:])
	case "push":
		pushCmd.Parse(args[2:])
	case "status":
		statusCmd.Parse(args[2:])
	case "update-index":
		updIndCmd.Parse(args[2:])
	case "pull":
		pullCmd.Parse(args[2:])
	default:
		return 1
	}

	args = args[2:] //update args to no longer having the app name and the command name
	
	

	//now we have to go through each of the subcommands to know the one that was passed. we then execute our program logic
	if initCmd.Parsed() {
		if len(args) < 1 { //no path argument
			fmt.Fprintf(os.Stderr, "Init needs  the directory specified as an argument. \n. If this is the wkdir, put .")
			return 1
		}
		wkdir := args[0]
		if err := pkg.Init(wkdir); err != nil {
			return 1
		}
	}

	//not an initialization, so NewGot() shound be valid
	got := pkg.NewGot()
	if addCmd.Parsed() {
		if err := got.Add(addFlag, args...); err != nil {
			return 1
		}

	} else { //workspace has already been initialized and its wkdir has been saved in our sonfig file
		
	}

	return 0
}
