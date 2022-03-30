package cmd

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/OLUWAMUYIWA/got/pkg"
)

type app struct {
	*log.Logger
}

func NewApp() *app {
	a := &app{
		log.New(os.Stdout, "Got Command Error", log.Ldate|log.Ltime),
	}
	return a
}

func (a *app) Run() error {
	//comeback: maybe i want to have a timeout here
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	go func() { // cancel when the channel gets notified
		select {
		case sig := <-c:
			a.Printf(fmt.Sprintf("Program stopped due to this signal: %s", sig))
			cancel()
		}
	}()

	runner, err := a.parseArgs(ctx)

	if err != nil {
		return err
	}

	if err := runner.Run(ctx); err != nil {
		return err
	}

	return nil
}

//command list
// add 	branch cat 	commit 	config 	diff 	fetch 	hash 	init 	ls-files 	ls-tree 	merge
// pull 	push 	read-tree 	remote 	rm 	status 	switch 	update-index 	verify-pack 	write-tree
//

//comeback handle exit codes and context
func (a *app) parseArgs(ctx context.Context) (Runner, error) {

	// add
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	var aall bool
	addCmd.BoolVar(&aall, "all", false, "Specify that all files starting from root directory should de added")
	addCmd.BoolVar(&aall, "A", false, "Specify that all files starting from root directory should de added (shorthand)")

	branchCmd := flag.NewFlagSet("branch", flag.ExitOnError)

	//cat
	catCmd := flag.NewFlagSet("cat-file", flag.ExitOnError)
	var size, _type, pretty bool
	catCmd.BoolVar(&size, "s", false, "specify that we only need the size")
	catCmd.BoolVar(&_type, "t", false, "specify that we only need the type")
	catCmd.BoolVar(&pretty, "p", false, "specify that we nned pretty printing")

	// commit
	// supports only the first two ways of committing as described in https://git-scm.com/docs/git-commit
	// it expexts that an `add` has already been run, or a `rm` after an `add`.
	commitCmd := flag.NewFlagSet("commit", flag.ExitOnError)
	var cmtMsg string
	var c_all bool
	commitCmd.StringVar(&cmtMsg, "m", "update", "commit after add or remove")
	commitCmd.BoolVar(&c_all, "a", false, `Tell the command to automatically stage files that have been modified and deleted, 
		but new files you have not told Git about are not affected.`)
	commitCmd.BoolVar(&c_all, "all", false, `Tell the command to automatically stage files that have been modified and deleted, 
		but new files you have not told Git about are not affected.`)

	// config
	configCmd := flag.NewFlagSet("config", flag.ExitOnError)

	// diff
	var cached bool
	var output string
	diffCmd := flag.NewFlagSet("diff", flag.ExitOnError)
	diffCmd.BoolVar(&cached, "cached", false, `Cached instructs git-diff to check for changes in the working tree
	 on files that have already  been staged in the index. if its not set, git-diff
	  checks for changes in  WT that have not been added`)
	diffCmd.StringVar(&output, "output", "", "Dumps diff in file instead of standard output")

	// fetch
	fetchCmd := flag.NewFlagSet("fetch", flag.ExitOnError)

	// hash-object
	hashObjCmd := flag.NewFlagSet("hash-object", flag.ExitOnError)

	// initializing & configuration
	// init
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)

	// ls-files
	lsFilesCmd := flag.NewFlagSet("ls-files", flag.ExitOnError)
	var lstaged, lcached, ldeleted, lmodified, lothers bool
	lsFilesCmd.BoolVar(&lstaged, "s", false, "Show staged contents' mode bits, object name and stage number in the output.")
	lsFilesCmd.BoolVar(&lcached, "c", true, "Show cached files in the output, default") //comeback
	lsFilesCmd.BoolVar(&ldeleted, "d", false, "Show deleted files in the output ")
	lsFilesCmd.BoolVar(&lmodified, "m", false, "Show modified files in the output ")
	lsFilesCmd.BoolVar(&lothers, "o", false, "Show untracked files in the output ")

	//ls-tree
	lsTreeCmd := flag.NewFlagSet("ls-tree", flag.ExitOnError)

	mergeCmd := flag.NewFlagSet("merge", flag.ExitOnError)

	// pull
	pullCmd := flag.NewFlagSet("pull", flag.ExitOnError)

	// push
	pushCmd := flag.NewFlagSet("push", flag.ExitOnError)

	//read-tree
	readTreeCmd := flag.NewFlagSet("read-tree", flag.ExitOnError)

	// remote
	rmtCmd := flag.NewFlagSet("remote", flag.ExitOnError)

	// rm
	rmvCmd := flag.NewFlagSet("rm", flag.ExitOnError)
	// comeback trick. just trying something
	rmvCmd.Usage = func() {
		os.Stdout.Write([]byte("Error parsing the rm command"))
	}
	var rcached bool
	rmvCmd.BoolVar(&rcached, "c", false, `Use this option to unstage and remove paths only from the index. 
		Working tree files, whether modified or not, will be left alone.`[1:])

	// status
	statusCmd := flag.NewFlagSet("status", flag.ExitOnError)

	//switch
	switchCmd := flag.NewFlagSet("switch", flag.ExitOnError)
	var Sname string
	switchCmd.StringVar(&Sname, "c", "", "Creates a new branch and checks it out")

	// update-index
	updIndCmd := flag.NewFlagSet("update-index", flag.ExitOnError)

	//verify-pack
	verifyPackCmd := flag.NewFlagSet("verify-pack", flag.ExitOnError)

	//write-tree
	writeTreeCmd := flag.NewFlagSet("write-tree", flag.ExitOnError)

	flag.Parse()

	args := flag.Args()

	if len(args) < 2 {
		return nil, fmt.Errorf("No subcommand provided for git to work with/n")
	}

	//parse each of the subcommands, starting from the second argument.
	//remember that args[0] will be the program binary name
	//args[1] is our subcommand.
	switch args[1] {

	case "add":
		addCmd.Parse(args[2:])
	case "branch":
		branchCmd.Parse(args[2:])
	case "cat-file":
		catCmd.Parse(args[2:])
	case "commit":
		commitCmd.Parse(args[2:])
	case "config":
		configCmd.Parse(args[2:])
	case "diff":
		diffCmd.Parse(args[2:])
	case "fetch":
		fetchCmd.Parse(args[2:])
	case "hash-object":
		hashObjCmd.Parse(args[2:])
	case "init":
		initCmd.Parse(args[2:])
	case "ls-files":
		lsFilesCmd.Parse(args[2:])
	case "ls-tree":
		lsTreeCmd.Parse(args[2:])
	case "merge":
		mergeCmd.Parse(args[2:])
	case "pull":
		pullCmd.Parse(args[2:])
	case "push":
		pushCmd.Parse(args[2:])
	case "read-tree":
		readTreeCmd.Parse(args[2:])
	case "remote":
		rmtCmd.Parse(args[2:])
	case "status":
		statusCmd.Parse(args[2:])
	case "switch":
		switchCmd.Parse(args[2:])
	case "update-index":
		updIndCmd.Parse(args[2:])
	case "verify-pack":
		verifyPackCmd.Parse(args[2:])
	case "write-tree":
		writeTreeCmd.Parse(args[2:])
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
		return &initializer{
			wkdir,
		}, nil
	}

	switch {
	//comeback
	case addCmd.Parsed():
		{
			return &add{
				aall, args,
			}, nil

		}

	case branchCmd.Parsed():
		{
			name := branchCmd.Arg(0)
			//comebac for dekete
			b := branch{
				name: name,
			}
			return &b, nil
		}

	case commitCmd.Parsed():
		{

		}

	case diffCmd.Parsed():
		{
			if len(args) > 1 {
				return nil, fmt.Errorf("Dif parse Error: We currently support only one arg for diffs")
			}

			if cached && args[0] != "" {
				return nil, fmt.Errorf("We do not support ")
			}

			return &diff{
				cached: cached,
				output: output,
				arg:    args[0],
			}, nil
		}

	case rmvCmd.Parsed():
		{
			if len(args) > 0 {
				return &rm{
					rcached,
					args,
				}, nil
			} else {
				return nil, pkg.ArgsIncomplete()
			}
		}
	case commitCmd.Parsed():
		{
			if len(args) != 0 {
				return nil, pkg.ArgsIncomplete()
			}
			return &commit{
				msg: cmtMsg,
			}, nil
		}
	case catCmd.Parsed():
		{
			if len(args) == 1 {
				if !_type && size && !pretty {
					return &cat{prefix: args[0], mode: 0}, nil
				} else if _type && !size && !pretty {
					return &cat{prefix: args[0], mode: 1}, nil
				} else if !_type && !size && pretty {
					return &cat{prefix: args[0], mode: 2}, nil
				} else {
					return nil, fmt.Errorf("Only one of the three flags must be set\n")
				}
			}
			return nil, fmt.Errorf("Only one argument is needed by command")
		}

	default:
		{
			return nil, fmt.Errorf("Error parrsing flags")
		}
	}

	return nil, fmt.Errorf("Error parrsing args")
}
