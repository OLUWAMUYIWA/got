package pkg

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//NB: this is a much simplified version of Git's configuration
//read: https://git-scm.com/docs/git-config
//I used regexp here. read: https://github.com/google/re2/wiki/Syntax . it's quite interesting and straightforward
//This is actually one of the hardest parts of the project, the third or fourth hardest perhaps. I had to parse the whole thing by hand
//we parse the git config line by line because it is human readable, editable as well as machine-readable and editable

type User struct {
	Uname string `json: uname`
	Email string `json: email`
}

func Config(conf User) error {
	confRoot, err := os.UserConfigDir()
	if err != nil {
		return &OpErr{Context: "IO: While getting the config directory", inner: err}
	}
	err = os.Mkdir(filepath.Join(confRoot, ".git"), os.ModeDir)
	if err != nil {
		return &OpErr{Context: "IO: While creating direcctory In Config", inner: err}
	}
	f, err := os.Create(filepath.Join(confRoot, ".git", ".config"))
	if err != nil {
		return &OpErr{Context: "IO: While creating .config file", inner: err}
	}
	enc := json.NewEncoder(f)
	err = enc.Encode(conf)
	if err != nil {
		return &OpErr{Context: "While encoding json", inner: err}
	}

	return nil
}

type lineType uint8

const (
	comment lineType = 1 << iota //used for comment lines or empty lines
	section                      //section title
	kv                           //key-value
	emt                          //empty line
)

//default RefSpecs
const (
	FetchRefSpec = "+refs/heads/*:refs/remotes/%s/*"
	PushRefSpec  = "refs/heads/*:refs/heads/*"
)

var (
	rCmt   = regexp.MustCompile(`(?m)^\s*(?P<cmt>[#;]\w+)$`)
	rEmt   = regexp.MustCompile(`(?m)^\s*$`)
	rNumb  = regexp.MustCompile(`\A\s*(?P<num>-?\d+)\z`)                                                      //might not need this
	rSectn = regexp.MustCompile(`(?im)^\s*\[(?P<sect>\w+)( "(?P<subsect>\w*)")?\]\s*(?P<cmt>[#;]\s*\w\s*)?$`) //may have comments in front
	rKv    = regexp.MustCompile(`(?im)\A\s*(?P<key>[[:alpha:]]\w*)\s*=\s*(?P<val>\w+)\s*(?P<cmt>#\s*\w\s*)?$`)
)

type ConfigParse struct {
	sections map[string]Section
}

type Section struct {
	count int  //needed so we can arrange it back
	title Line //the title line contains a KV too, but ere, the key is the section string and the value is the subsection string
	subs  []Line
}

//if comment, k is empty. if sect, k is sect, v is subsect if it exists.
type Line struct {
	count int
	_type lineType
	cont  []byte
	kv    KV
	cmt   InlineCmt
}

//stands for comment lines and empty lines
type InlineCmt struct {
	text []byte
}

type KV struct {
	k []byte
	v []byte
}

func newSect(sectCount int) Section {
	return Section{
		count: sectCount,
		title: Line{},
		subs:  []Line{},
	}
}
func parseConfig(basepath string) (*ConfigParse, error) {
	// root, err := os.UserCacheDir()
	//conf := filepath.Join(path, ".git")
	data, err := fs.ReadFile(os.DirFS(basepath), ".config")
	if err != nil {
		return nil, fmt.Errorf("Parsing Config error: %w", err)
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Split(bufio.ScanLines)

	sections := make(map[string]Section)
	sectCount := 0                 //initialize the section count. we need it when we're rearranging to save config as file.
	currSect := newSect(sectCount) //section line that were dealing with currently. starting with the first one
	for scanner.Scan() {
		lineCount := 0 //initialize the line count. needed when we're saving config back
		line := scanner.Bytes()
		if rCmt.Match(line) {
			currSect.subs = append(currSect.subs, Line{count: lineCount, cont: rCmt.FindSubmatch(line)[rCmt.SubexpIndex("cmt")]})
			lineCount += 1
		} else if rSectn.Match(line) {
			//new section discovered. wrap up old one
			sectCount += 1
			currSect = newSect(sectCount) //new currSect
			currSect.title = parseSect(line, lineCount)
			sections[string(currSect.title.kv.k)] = currSect
			lineCount += 1
		} else if rKv.Match(line) {
			kv := parseKv(line, lineCount)
			//TODO check if currSect.title is nil. return error
			currSect.subs = append(currSect.subs, kv)
			lineCount += 1
		} else if rEmt.Match(line) { //empty line comes after the other matches because i'm afraid other matches may match it
			currSect.subs = append(currSect.subs, Line{count: lineCount})
			lineCount += 1
		} else {
			return nil, fmt.Errorf("Could not match line: %d", lineCount+2) // linecount +2 because, first we're indexing from zero, and second, it is the line after the last lineCount that is not being matched
		}
	}
	return nil, nil
}

func parseSect(line []byte, count int) Line {
	match := rSectn.FindSubmatch(line)
	sect := match[rSectn.SubexpIndex("sect")]
	subsect := match[rSectn.SubexpIndex("subsect")]
	cmt := match[rSectn.SubexpIndex("cmt")]
	return Line{
		count: count,
		_type: section,
		cont:  line,
		kv:    KV{k: sect, v: subsect},
		cmt: InlineCmt{
			text: cmt,
		},
	}
}

func parseKv(line []byte, count int) Line {
	match := rKv.FindSubmatch(line)
	return Line{
		count: count,
		_type: kv,
		cont:  line,
		kv:    KV{k: match[rKv.SubexpIndex("key")], v: match[rKv.SubexpIndex("val")]},
		cmt:   InlineCmt{text: match[rKv.SubexpIndex("cmt")]},
	}
}

func parseCmt(line []byte, count int) Line {
	match := rCmt.FindSubmatch(line)
	return Line{
		_type: comment,
		count: count,
		cont:  match[rCmt.SubexpIndex("cmt")],
	}
}

func parseEmpty(count int) Line {
	return Line{
		_type: comment,
		count: count,
	}
}

func (conf *ConfigParse) add(k, v string) error {
	splits := strings.Split(k, ".")
	key, rest := splits[0], splits[1:]
	sect, ok := conf.sections[key]
	if ok { // section already exists
		switch len(sect.title.kv.v) { //does it have a subsection or not? sect.title.kv.v stores the subsection of the section.
		case 0:
			{ //no subsection
				l := len(rest)
				if l != 1 {
					return fmt.Errorf("More or less depth than expected")
				} else {
					for i, l := range conf.sections[key].subs {
						if bytes.Compare(l.kv.k, []byte(rest[0])) == 0 { //this is the k-v pair in question
							conf.sections[key].subs[i].kv.v = []byte(v)
						}
					}
				}
			}
		case 1: // there's a subsection
			{
				l := len(rest)
				if l != 2 {
					return fmt.Errorf("More or Less depth than expected")
				} else if bytes.Compare(sect.title.kv.v, []byte(rest[0])) == 0 { //is it the exact subsection we want?
					for i, val := range conf.sections[key].subs {
						if bytes.Compare(val.kv.k, []byte(rest[1])) == 0 { //is this the exact kv that we want?
							conf.sections[key].subs[i].kv.v = []byte(v)
						}
					}
				} else { //subsection did not exist before, create it

				}
			}
		}
	} else { //new

	}
	//conf.sections[key] = sect
	conf.Save()

	return nil
}

//todo: use the sorted keys trick to sort the map based on the keys here
func (conf *ConfigParse) Save() error {
	return nil
}

func InitCinfig(path string) (*ConfigParse, error) {
	return nil, nil
}
