package cmd

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/OLUWAMUYIWA/got/pkg"
)

type app struct {
	*log.Logger
}

func NewApp() *app {
	a := &app{
		log.New(os.Stdout, "Got: ", log.LstdFlags),
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

	// subcommands as `NewFlagSet`s

	// add
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	var aall bool
	addCmd.BoolVar(&aall, "all", false, "Specify that all files starting from root directory should be added, default false")
	addCmd.BoolVar(&aall, "A", false, "Specify that all files starting from root directory should be added (shorthand). default: false")

	//branch
	branchCmd := flag.NewFlagSet("branch", flag.ExitOnError)
	var deleteBranch bool
	branchCmd.BoolVar(&deleteBranch, "d", false, "delete the named branch")
	branchCmd.BoolVar(&deleteBranch, "delete", false, "delete the named branch")

	//cat
	catCmd := flag.NewFlagSet("cat-file", flag.ExitOnError)
	var size, _type, pretty bool
	catCmd.BoolVar(&size, "s", false, "specify that we only need the size")
	catCmd.BoolVar(&_type, "t", false, "specify that we only need the type")
	catCmd.BoolVar(&pretty, "p", false, "specify that we nned pretty printing")

	//checkout
	checkoutCmd := flag.NewFlagSet("checkout", flag.ExitOnError)
	var newBranchCheckout bool
	checkoutCmd.BoolVar(&newBranchCheckout, "b", false, "Creates a new branch and checks it out")

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
	var local, global, system bool
	configCmd.BoolVar(&local, "local", false, "apply config locally")
	configCmd.BoolVar(&global, "global", false, "apply config globally for user")
	configCmd.BoolVar(&system, "system", false, "apply config system-wide for all users")

	// diff
	diffCmd := flag.NewFlagSet("diff", flag.ExitOnError)
	var cached bool
	var output string
	diffCmd.BoolVar(&cached, "cached", false,
		`Cached instructs git-diff to check for changes in the working tree
	 on files that have already  been staged in the index. if its not set, git-diff
	 checks for changes in  WT that have not been added`)
	diffCmd.StringVar(&output, "output", "", "Dumps diff in file instead of standard output")

	// fetch
	fetchCmd := flag.NewFlagSet("fetch", flag.ExitOnError)

	// hash-object
	hashObjCmd := flag.NewFlagSet("hash-object", flag.ExitOnError)
	var hashW bool
	var hashType string
	hashObjCmd.BoolVar(&hashW, "w", false, "write")
	hashObjCmd.StringVar(&hashType, "t", "blob", "specify obect type")

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

	//merge
	mergeCmd := flag.NewFlagSet("merge", flag.ExitOnError)

	// pull
	pullCmd := flag.NewFlagSet("pull", flag.ExitOnError)
	var rebase bool
	pullCmd.BoolVar(&rebase, "rebase", false, "Rebase if conflict exist")

	// push
	pushCmd := flag.NewFlagSet("push", flag.ExitOnError)

	//read-tree
	readTreeCmd := flag.NewFlagSet("read-tree", flag.ExitOnError)

	// remote
	rmtCmd := flag.NewFlagSet("remote", flag.ExitOnError)

	// rm
	rmvCmd := flag.NewFlagSet("rm", flag.ExitOnError)
	// comeback. just trying something out
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
	var newBranchSwitch bool
	switchCmd.BoolVar(&newBranchSwitch, "c", false, "Creates a new branch and checks it out")
	switchCmd.BoolVar(&newBranchSwitch, "C", false, "Creates a new branch and checks it out")

	// update-index
	updIndCmd := flag.NewFlagSet("update-index", flag.ExitOnError)
	var addInd, rmvInd bool
	updIndCmd.BoolVar(&addInd, "add", false, `If a specified file isn’t in the index already then it’s added. 
		Default behaviour is to ignore new files.`)
	updIndCmd.BoolVar(&rmvInd, "add", false, `If a specified file isn’t in the index already then it’s removed. 
		Default behaviour is to ignore removed files.`)

	//verify-pack
	verifyPackCmd := flag.NewFlagSet("verify-pack", flag.ExitOnError)

	//write-tree
	writeTreeCmd := flag.NewFlagSet("write-tree", flag.ExitOnError)

	// parse flags. populates flag.Args()
	flag.Parse()

	// args represent the arguments provided to the main command. unlike os.Args, it does not include the main command
	args := flag.Args()

	if len(args) < 1 {
		return nil, fmt.Errorf("No subcommand provided for git to work with/n")
	}

	//parse each of the subcommands, starting from the second argument.
	// of course only one is parsed per execution of the program
	// args[0] is our subcommand in this case (wouldve been args[1] if we had used os.Args)
	switch args[0] {
	// parse the flags as defined by the flag sets above
	case "add":
		addCmd.Parse(args[1:])
	case "branch":
		branchCmd.Parse(args[1:])
	case "cat-file":
		catCmd.Parse(args[1:])
	case "checkout":
		checkoutCmd.Parse(args[1:])
	case "commit":
		commitCmd.Parse(args[1:])
	case "config":
		configCmd.Parse(args[1:])
	case "diff":
		diffCmd.Parse(args[1:])
	case "fetch":
		fetchCmd.Parse(args[1:])
	case "hash-object":
		hashObjCmd.Parse(args[1:])
	case "init":
		initCmd.Parse(args[1:])
	case "ls-files":
		lsFilesCmd.Parse(args[1:])
	case "ls-tree":
		lsTreeCmd.Parse(args[1:])
	case "merge":
		mergeCmd.Parse(args[1:])
	case "pull":
		pullCmd.Parse(args[1:])
	case "push":
		pushCmd.Parse(args[1:])
	case "read-tree":
		readTreeCmd.Parse(args[1:])
	case "remote":
		rmtCmd.Parse(args[1:])
	case "rm":
		rmvCmd.Parse(args[1:])
	case "status":
		statusCmd.Parse(args[1:])
	case "switch":
		switchCmd.Parse(args[1:])
	case "update-index":
		updIndCmd.Parse(args[1:])
	case "verify-pack":
		verifyPackCmd.Parse(args[1:])
	case "write-tree":
		writeTreeCmd.Parse(args[1:])
	default:
		return nil, fmt.Errorf("Error parrsing flags and args")
	}

	//now we have to go through each of the subcommands to know the one that was passed. we then execute our program logic
	if initCmd.Parsed() {

		initArgs := initCmd.Args()
		if len(initArgs) < 1 { //no path argument provided
			return nil, fmt.Errorf("Init needs  the directory specified as an argument. \n. If this is the wkdir, put .")
		}
		return &initializer{
			initArgs[0],
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
			if len(branchCmd.Args()) < 1 { //no path argument provided
				return nil, fmt.Errorf("branch needs  the name specified as an argument. \n.")
			}
			name := branchCmd.Arg(0)
			b := branch{
				name:   name,
				delete: deleteBranch,
			}
			return &b, nil
		}

	case catCmd.Parsed():
		{
			catArgs := catCmd.Args()
			if len(catArgs) == 1 {
				if !_type && size && !pretty {
					return &cat{prefix: catArgs[0], mode: 0}, nil
				} else if _type && !size && !pretty {
					return &cat{prefix: catArgs[0], mode: 1}, nil
				} else if !_type && !size && pretty {
					return &cat{prefix: catArgs[0], mode: 2}, nil
				} else {
					return nil, fmt.Errorf("Only one of the three flags must be set\n")
				}
			}
			return nil, fmt.Errorf("Only one argument is needed by command")
		}

	case checkoutCmd.Parsed():
		{

			name := checkoutCmd.Arg(0)
			if name == "" {
				return nil, fmt.Errorf("Error parsing flags")
			}

			return &checkout{
				name: name,
				new:  newBranchCheckout,
			}, nil

		}

	case commitCmd.Parsed():
		{
			if len(commitCmd.Args()) > 0 {
				return nil, fmt.Errorf("Commit args parse Error: we do not support having arguments with commit")
			}
			return &commit{
				msg: cmtMsg,
				all: c_all,
			}, nil
		}

	case configCmd.Parsed():
		{
			cargs := configCmd.Args()
			if len(cargs) < 1 {
				return nil, fmt.Errorf("Confing suports min of one argument")
			}
			sekKey := strings.Split(cargs[0], ".")
			val := ""
			if len(cargs) == 2 {
				val = cargs[1]
			}
			if len(cargs) > 2 {
				return nil, fmt.Errorf("Confing suports max of two arguments")
			}
			return &config{
				section: sekKey[0],
				key:     sekKey[1],
				value:   val,
				local:   local,
				global:  global,
				system:  system,
			}, nil
		}

	case diffCmd.Parsed():
		{
			// cached and diff arg provided == error
			if cached && diffCmd.Arg(0) != "" {
				return nil, fmt.Errorf("We do not support ")
			}

			return &diff{
				cached: cached,
				output: output,
				arg:    diffCmd.Arg(0),
			}, nil
		}

	case fetchCmd.Parsed():
		{
			fetchArgs := fetchCmd.Args()
			if len(fetchArgs) > 1 {
				return nil, fmt.Errorf("Fetch expects zero or one arguments, namely the remote")
			}
			return &fetch{remote: fetchArgs[0]}, nil
		}

	case hashObjCmd.Parsed():
		{
			if len(hashObjCmd.Args()) != 1 {
				return nil, fmt.Errorf("Fetch expects one argument, namely the object filename")
			}

			return &hashObj{
				_type: hashType,
				w:     hashW,
				file:  hashObjCmd.Arg(0),
			}, nil
		}

	case lsFilesCmd.Parsed():
		{
			return &lsFiles{
				lstaged:   lstaged,
				lcached:   lcached,
				ldeleted:  ldeleted,
				lmodified: lmodified,
				lothers:   lothers,
			}, nil
		}

	case lsTreeCmd.Parsed():
		{
			if len(lsTreeCmd.Args()) != 1 {
				return nil, fmt.Errorf("Error parrsing args, expected one argument: path")
			}
			return &lsTree{
				path: lsTreeCmd.Arg(0),
			}, nil
		}

	case mergeCmd.Parsed():
		{
			if len(mergeCmd.Args()) != 1 {
				return nil, fmt.Errorf("Error parrsing args. expected commit to merge")
			}
			return &merge{
				comm: mergeCmd.Arg(0),
			}, nil
		}

	case pullCmd.Parsed():
		{
			pullArgs := pullCmd.Args()
			if len(pullArgs) != 1 {
				return nil, fmt.Errorf("Error parrsing args")
			}
			return &pull{
				remote: pullArgs[0],
				rebase: rebase,
			}, nil
		}

	case pushCmd.Parsed():
		{
			pushArgs := pushCmd.Args()
			if len(pushArgs) != 1 {
				return nil, fmt.Errorf("Error parrsing args")
			}

			return &push{
				repo: pushArgs[0],
			}, nil
		}

	case readTreeCmd.Parsed():
		{
			if len(readTreeCmd.Args()) != 1 {
				return nil, fmt.Errorf("Error parsing flags")
			}
			return &readTree{
				treeish: readTreeCmd.Arg(0),
			}, nil
		}

	case rmtCmd.Parsed():
		{
			rmtArgs := rmtCmd.Args()
			var rmt remote
			if len(rmtArgs) != 2 {
				return nil, fmt.Errorf("Remote command expects two arguments")
			}
			if rmtArgs[0] == "add" {
				rmt._type = 0
			} else if rmtArgs[0] == "remove" {
				rmt._type = 1
			} else {
				return nil, fmt.Errorf("Remote command expects one of two subcommands: `add` or `remove`")
			}
			rmt.name = rmtArgs[1]
			return &rmt, nil
		}

	case rmvCmd.Parsed():
		{
			rmvArgs := rmvCmd.Args()
			if len(rmvArgs) > 0 {
				return &rm{
					rcached,
					rmvArgs,
				}, nil
			} else {
				return nil, pkg.ArgsIncomplete()
			}
		}

	case statusCmd.Parsed():
		{
			return &status{}, nil
		}

	case switchCmd.Parsed():
		{
			name := switchCmd.Arg(0)
			if name == "" {
				return nil, fmt.Errorf("Error parsing flags")
			}

			return &_switch{
				name: name,
				new:  newBranchSwitch,
			}, nil
		}

	case updIndCmd.Parsed():
		{
			return &updateIndex{
				add:    addInd,
				remove: rmvInd,
			}, nil
		}

	case verifyPackCmd.Parsed():
		{
			if len(verifyPackCmd.Args()) != 1 {
				return nil, fmt.Errorf("Error parsing flags")
			}

			return &verifyPack{idx: verifyPackCmd.Arg(0)}, nil
		}

	case writeTreeCmd.Parsed():
		{
			return &writeTree{}, nil
		}

	default:
		{
			return nil, fmt.Errorf("Error parsing flags")
		}
	}
}
