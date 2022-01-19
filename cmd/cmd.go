package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"github.com/OLUWAMUYIWA/got/pkg"
)

type app struct {
	logger *log.Logger
}

func (a *app) new() {
	a.logger = log.New(os.Stdout, "Got Command Error", log.Ldate | log.Ltime)
}


func (a app) run() {

}

//comeback handle exit codes
func (a *app) parseArgs() (runner, error) {

	// initializing & configuration
	// init
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	
	//config
	configCmd := flag.NewFlagSet("config", flag.ExitOnError)
	

	// snapshotting

	// add
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	var addFlag bool
	addCmd.BoolVar(&addFlag, "all", false, "Specify that all files starting from root directory should de added")
	addCmd.BoolVar(&addFlag, "A", false, "Specify that all files starting from root directory should de added (shorthand)")

	// status
	statusCmd := flag.NewFlagSet("status", flag.ExitOnError)

	// rm
	rmvCmd := flag.NewFlagSet("rm", flag.ExitOnError)

	// commit
	// supports only the first two ways of committing as described in https://git-scm.com/docs/git-commit
	// it expexts that an `add` has already been run, or a `rm` after an `add`.
	commitCmd := flag.NewFlagSet("commit", flag.ExitOnError)
	var cmtMsg string 
	commitCmd.StringVar(&cmtMsg, "m", "update", "commit after add or remove" )



	// inspection

	//diff
	diffCmd := flag.NewFlagSet("diff", flag.ExitOnError)
	


	// sharing
	
	// fetch
	fetchCmd := flag.NewFlagSet("fetch", flag.ExitOnError)

	// pull
	pullCmd := flag.NewFlagSet("pull", flag.ExitOnError)
	
	// push
	pushCmd := flag.NewFlagSet("push", flag.ExitOnError)

	// remote
	rmtCmd := flag.NewFlagSet("remote", flag.ExitOnError)

	


	//plumbing

	//cat
	catCmd := flag.NewFlagSet("cat-file", flag.ExitOnError)
	var _type, size, pretty bool
	catCmd.BoolVar(&_type, "t", false, "specify that we only need the type" )
	catCmd.BoolVar(&size, "s", false, "specify that we only need the size" )
	catCmd.BoolVar(&pretty, "p", false, "specify that we nned pretty printing" )

	//ls-files
	lsFilesCmd := flag.NewFlagSet("ls-files", flag.ExitOnError)

	//ls-tree 
	lsTreeCmd := flag.NewFlagSet("ls-tree", flag.ExitOnError)

	//read-tree
	readTreeCmd := flag.NewFlagSet("read-tree", flag.ExitOnError)

	//verify-pack
	verifyPackCmd := flag.NewFlagSet("verify-pack", flag.ExitOnError)

	//write-tree
	writeTreeCmd := flag.NewFlagSet("write-tree", flag.ExitOnError)

	//hash-object
	hashObjCmd := flag.NewFlagSet("hash-object", flag.ExitOnError)

	//update-index
	updIndCmd := flag.NewFlagSet("update-index", flag.ExitOnError)

	flag.Parse()

	args := flag.Args()
	
	if len(args) < 2 {
		return nil, fmt.Errorf("No subcommand provided for git to work with/n")
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
		case "ls-tree":
			lsTreeCmd.Parse(args[2:])
		case "push":
			pushCmd.Parse(args[2:])
		case "remote":
			rmtCmd.Parse(args[2:])
		case "verify-pack":
			verifyPackCmd.Parse(args[2:])
		case "read-tree":
			readTreeCmd.Parse(args[2:])
		case "write-tree":
			writeTreeCmd.Parse(args[2:])
		case "status":
			statusCmd.Parse(args[2:])
		case "update-index":
			updIndCmd.Parse(args[2:])
		case "pull":
			pullCmd.Parse(args[2:])
		default:
			return nil, fmt.Errorf("Error parrsing flags and args")
	}

	args = args[2:] //update args to no longer having the app name and the command name

	//now we have to go through each of the subcommands to know the one that was passed. we then execute our program logic
	if initCmd.Parsed() {
		if len(args) < 1 { //no path argument
			return nil, fmt.Errorf("Init needs  the directory specified as an argument. \n. If this is the wkdir, put .")
			
		}
		wkdir := args[0]
		return &ini{
			wkdir,
		}, nil
	}

	switch  {
		//comeback
	case addCmd.Parsed(): {
		return &add{
			addFlag, args,
		}, nil

	} 
	case rmvCmd.Parsed(): {
		if len(args) > 0 {
			return &rm {
				args,
			}, nil
		} else {
			return nil, pkg.ArgsIncomplete()
		}
	}
	case commitCmd.Parsed(): {
		if len(args) != 0 {
			return nil, pkg.ArgsIncomplete()
		}
		return &commit{
			msg: cmtMsg,
		}, nil
	} 
	case catCmd.Parsed(): {
		if len(args) == 1 {
			if (_type && !size && !pretty) {
				return &cat{args[0], 0}, nil
			} else if (!_type && size && !pretty) {
				return &cat{args[0], 1}, nil
			} else if (!_type && !size && pretty) {
				return &cat{args[0], 2}, nil
			} else {
				return nil, fmt.Errorf("Only one of the three flags must be set\n")
			}
		}
		return nil, fmt.Errorf("Only one argument is needed by command")
	}

	default: { 
		return nil, fmt.Errorf("Error parrsing flags and args")
	}
	} 

}


func Exec() int {
	return 0
}