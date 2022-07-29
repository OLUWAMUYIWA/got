package pkg

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"regexp"
	"sort"
)

type User struct {
	Uname string
	Email string
}

//NB: this is a much simplified version of Git's configuration
//read: https://git-scm.com/docs/git-config
//I used regexp here. read: https://github.com/google/re2/wiki/Syntax . it's quite interesting and straightforward
//This is actually one of the hardest parts of the project, the third or fourth hardest perhaps. I had to parse the whole thing by hand
//we parse the git config line by line because it is human readable, editable as well as machine-readable and editable

type lineType uint8

const (
	comment lineType = 1 << iota //used for comment lines
	section                      //section title
	kv                           //key-value. i.e. property
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
	rNumb  = regexp.MustCompile(`(?m)^\s*(?P<num>-?\d+)$`)                                                         //might not need this
	rSectn = regexp.MustCompile(`(?im)^\s*\[(?P<sect>\w+)(\s+"(?P<subsect>\w*)\s*")?\]\s*(?P<cmt>[#;]\s*\w\s*)?$`) //may have comments in front
	rKv    = regexp.MustCompile(`(?im)^\s*(?P<key>\w+)\s+=\s+(?P<val>\w+)\s*(?P<cmt>[#;]\s*\w\s*)?$`)
)

type Config struct {
	sections map[string]Section
}

type key struct {
	pos  int
	name string
}

type Section struct {
	count int  //needed so we can arrange it back
	title Line //the title line contains a KV too, but here, the key is the section string and the value is the subsection string
	subs  []Line
}

func (s Section) write(w io.Writer) error {
	var err error
	// first write title
	_, err = w.Write(s.title.cont)
	if err != nil {
		return err
	}
	// write everything else including empty lines and comments that precede the next section
	for _, l := range s.subs {
		_, err = w.Write(l.cont)
		if err != nil {
			return err
		}
	}
	return nil
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

func parseConfig(basepath string) (*Config, error) {
	// root, err := os.UserCacheDir()
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
			currSect.subs = append(currSect.subs, parseCmt(line, lineCount))
		} else if rSectn.Match(line) {
			//new section discovered. wrap up old one
			sectCount += 1
			currSect = newSect(sectCount) //new currSect
			currSect.title = parseSect(line, lineCount)
			sections[string(currSect.title.kv.k)] = currSect
		} else if rKv.Match(line) {
			kv := parseKv(line, lineCount)
			//TODO check if currSect.title is nil. return error
			currSect.subs = append(currSect.subs, kv)
		} else if rEmt.Match(line) { //empty line comes after the other matches because i'm afraid other matches may match it
			currSect.subs = append(currSect.subs, parseEmpty(lineCount))
		} else {
			return nil, fmt.Errorf("Could not match line: %d", lineCount+2) // linecount +2 because, first we're indexing from zero, and second, it is the line after the last lineCount that is not being matched
		}
		lineCount += 1
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

func joinKV(a, b []byte, endCmt []byte) []byte {
	joined := bytes.Join([][]byte{a, b}, []byte(" = "))
	if len(endCmt) == 0 { // no comment.
		return joined
	} else {
		return bytes.Join([][]byte{joined, endCmt}, bytes.Join([][]byte{[]byte(";"), endCmt}, []byte(" ")))
	}
}

func insert[T any](slice []T, item T, i int) []T {
	l := slice[:i]
	r := make([]T, len(slice)-i)
	copy(r, slice[i:])
	l = append(l, item)
	l = append(l, r...)
	return l
}

func (conf *Config) add(path []string, v string, w io.WriteCloser) error {
	if len(path) == 2 {
		sect, ok := conf.sections[path[0]]
		if ok { // section already exists
			if len(sect.title.kv.v) != 0 { //does it have a subsection or not?
				return fmt.Errorf("less depth provided than expected")
			}
			for _, l := range sect.subs {
				if bytes.Compare(l.kv.k, []byte(path[1])) == 0 { //this is the k-v pair in question
					l.kv.v = []byte(v)
					l.cont = joinKV(l.kv.k, l.kv.v, l.cmt.text) // so it will be printable
					break
				}
			}
		} else {
			// create a new line and append it
			l := Line{} // i dont know how to specify the count for tis particular line, and i dot know if it is useful
			l._type = kv
			l.kv.k = []byte(path[1])
			l.kv.v = []byte(v)
			l.cont = joinKV(l.kv.k, l.kv.v, nil)
			last := 0
			for i, l := range sect.subs { // record the last known kv line inside last
				if l._type == kv {
					last = i
				}
			}
			sect.subs = insert(sect.subs, l, last)
		}

	} else if len(path) == 3 {
		sect, ok := conf.sections[path[0]]
		if ok { // section already exists
			if len(sect.title.kv.v) != 0 && bytes.Equal(sect.title.kv.v, []byte(path[1])) { // match the key exists in that path
				for i, l := range sect.subs {
					if bytes.Compare(l.kv.k, []byte(path[2])) == 0 { //this is the k-v pair in question
						sect.subs[i].kv.v = []byte(v)
						break
					}
				}
			} else if !bytes.Equal(sect.title.kv.v, []byte(path[1])) { // new subsection. create it and then create a new kv

			} else { //error
				return fmt.Errorf(" depth provided is more than expected or invalid path")
			}

		} else {
			// create a new line and append it
			l := Line{} // i dont know how to specify the count for tis particular line, and i dot know if it is useful
			l._type = kv
			l.kv.k = []byte(path[1])
			l.kv.v = []byte(v)
			l.cont = joinKV(l.kv.k, l.kv.v, nil)
			last := 0
			for i, l := range sect.subs { // record the last known kv line inside last
				if l._type == kv {
					last = i
				}
			}
			sect.subs = insert(sect.subs, l, last)
		}
	} else {
		return fmt.Errorf("we dont expect paths with length greater than 3 or less than 2")
	}

	return conf.save(w)
}

func (conf *Config) save(w io.Writer) error {
	sect := make([]Section, len(conf.sections))
	i := 0
	for _, s := range conf.sections {
		sect[i] = s
		i++
	}
	sort.Slice(sect, func(i, j int) bool {
		return sect[i].count < sect[j].count
	})
	for _, s := range sect {
		if err := s.write(w); err != nil {
			return err
		}
	}

	return nil
}

func InitCinfig(path string) (*Config, error) {
	return nil, nil
}

const (
	local int = iota
	global
	system
)

// comeback
func (g *Got) ShowConf(path []string, where int) (io.Reader, error) {

	return nil, nil
}

// comeback
func (g *Got) UpdateConf(path []string, value string, where int) error {
	return nil
}

func (g *Got) Delete(path []string, where int) error {
	return nil
}
