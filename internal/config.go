package internal

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
)

//NB: this is a much simplified version of Git's configuration
//read: https://git-scm.com/docs/git-config
//I used regexp here. read: https://github.com/google/re2/wiki/Syntax . it's quite interesting and straightforward
//This is actually one of the hardest parts of the project, the third or fourth hardest perhaps. I had to parse the whole thing by hand
//we parse the git config line by line because it is human readable, editable as well as machine-readable and editable

type lineType uint8

const (
	comment lineType = 1 << iota //used for comment lines or empty lines
	section                      //section title
	kv                           //key-value
)

var (
	rCmt   = regexp.MustCompile(`(?m)^\s*(?P<cmt>#\w+)$`)
	rEmt   = regexp.MustCompile(`(?m)^\s*$`)
	rNumb  = regexp.MustCompile(`\A\s*(?P<num>-?\d)\z`)
	rSectn = regexp.MustCompile(`(?im)^\s*\[(?P<sect>\w+)( "(?P<subsect>\w*)")?\]\s*(?P<cmt>#\s*\w\s*)?$`) //may have comments in front
	rKv    = regexp.MustCompile(`(?im)\A\s*(?P<key>[[:alpha:]]\w*)\s*=\s*(?P<val>\w+)\s*(?P<cmt>#\s*\w\s*)?$`)
)

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

type ConfigObj struct {
	sections map[string]Section
}

type Section struct {
	count int //needed so we can arrange it back
	title Line
	subs  []Line
}

//if comment, k is empty. if sect, k is sect, v is subsect if it exists.
type Line struct {
	count int
	ty    lineType
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

func parseConfig() (*ConfigObj, error) {
	root, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("Parsing Cofig error: %w", err)
	}
	conf := filepath.Join(root, ".git")
	data, err := fs.ReadFile(os.DirFS(conf), ".config")
	if err != nil {
		return nil, fmt.Errorf("Parsing Cofig error: %w", err)
	}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Split(bufio.ScanLines)

	sections := make(map[string]Section)
	sectCount := 0
	for scanner.Scan() {
		lineCount := 0
		line := scanner.Bytes()
		currSect := Section{
			count: sectCount,
			title: Line{},
			subs:  []Line{},
		} //section line
		if rCmt.Match(line) {
			currSect.subs = append(currSect.subs, Line{count: lineCount, cont: rCmt.FindSubmatch(line)[rCmt.SubexpIndex("cmt")]})
			lineCount += 1
		} else if rEmt.Match(line) {
			currSect.subs = append(currSect.subs, Line{count: lineCount})
			lineCount += 1
		} else if rSectn.Match(line) {
			//new section discovered. wrap up old one
			//TODO: check if cureSect title == nil
			sections[string(currSect.title.cont)] = currSect
			sectCount += 1
			currSect = Section{} //empty currSect
			currSect.title = parseSect(line, lineCount)
			lineCount += 1

		} else if rKv.Match(line) {
			kv := parseKv(line, lineCount)
			//TODO check if currSect.title is nil. return error
			currSect.subs = append(currSect.subs, kv)
			lineCount += 1
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
		ty:    section,
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
		ty:    kv,
		cont:  line,
		kv:    KV{k: match[rKv.SubexpIndex("key")], v: match[rKv.SubexpIndex("val")]},
		cmt:   InlineCmt{text: match[rKv.SubexpIndex("cmt")]},
	}
}

func parseCmt(line []byte, count int) Line {
	match := rCmt.FindSubmatch(line)
	return Line{
		ty:    comment,
		count: count,
		cont:  match[rCmt.SubexpIndex("cmt")],
	}
}

func parseEmpty(count int) Line {
	return Line{
		ty:    comment,
		count: count,
	}
}
