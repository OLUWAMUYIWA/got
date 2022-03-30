package pkg

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const RFC2822 = "Mon Jan 2 15:04:05 2006 -0700"

//!!!!COMMIT OBJECT AND ITS GOTOBJECT IMPLEMENTATION!!!!!
type commitObj struct {
	sha  string
	data []byte
}

func (c *commitObj) Hash(wkdir string) ([]byte, error) {
	return nil, nil
}



type Sign struct {
	name, email string
	time time.Time
}

func parseSign(b [][]byte) Sign {
	s := Sign{}
	s.name = string(b[0])
	s.email = string(bytes.TrimFunc(b[1], func(r rune) bool {
		return r == '<' || r == '>'
	}))

	//comeback for the time parsing
	timestamp, err  := strconv.ParseInt(string(b[2]), 10, 0)
	if err != nil {
		return nil, err
	}

	zoneHr, err := strconv.ParseInt(string(b[3])[0:3], 10, 0)
	if err != nil {
		return nil, err
	}

	zoneMin, err := strconv.ParseInt(string(b[3])[3:], 10, 0)
	if err != nil {
		return nil, err
	}

	if zoneHr < 0 {
		zoneMin *= -1
	}

	s.time := time.Unix(timestamp, 0).In(time.FixedZone("", int(zoneHr * 60 * 60 + zoneMin * 60))) //.Format("Wed Jan _5 15:04:05 2010")

	return s
}

func (s *Sign) Format() string {
	var str strings.Builder
	str.WriteString(fmt.Sprintf("%s <%s>", s.name, s.email))
	str.WriteString("%d %s", s.time.Unix(), s.time.Format("-7000"))
	return str.String()
}

type Comm struct {
	sha Sha1
	treeSha Sha1
	parents []Sha1
	committer Sign
	author Sign
	msg string
}


type LineType = string

const (
	lineTree      LineType = "tree"
	lineComm         = "committer"
	linePar            = "parent"
	lineAuth            = "author"
)

func parseCommit(rdr io.Reader, got *Got) (*Comm, error) {

	scanner := bufio.NewScanner(rdr)
	scanner.Split(bufio.ScanLines)
	msg := bytes.Buffer{}
	comm := &Comm{}
	var msgOn bool
	for scanner.Scan() {
		line := scanner.Bytes()
		splits := bytes.FieldsFunc(line, func(r rune) bool {
			return unicode.IsSpace(r)
		})

		prefix := string(splits[0])

		if len(line) == 0 && !msgOn {
			msgOn = true
			continue
		}

		if !msgOn {
			switch prefix {
			case lineTree: {
				if len(splits) != 2 {
					return nil, fmt.Errorf("Tree line in commit object faulty")
				}
				treeSha := splits[1]
				comm.treeSha = bytesToSha(treeSha)
			}

			case linePar: {
				if len(splits) == 2 || len(splits[2]) != 20 {
					return nil, fmt.Errorf("Parent line in commit faulty")
				}
				p := splits[1]
				parent := (*[20]byte)(p) //comeback to this cool hack. its a tip. it panics if the lengths of slice and array are different
				comm.parents = append(comm.parents, *parent)
			}

			case lineAuth: {
				rem := splits[1:]
				comm.author = parseSign(rem)
			}

			case lineComm: {
				rem := splits[1:]
				comm.committer = parseSign(rem)
			}
		
			}
		} else {
			msg.Write(append(line, byte('\n'))) // we stripped it of the '\n' earlier
		}

		
	}
	comm.msg = msg.String()
	return comm, nil
}


func (c *Comm) Hash(wkdir string) ([]byte, error) {
	b, err := HashObj(c.Type(), c.data, wkdir)
	if err != nil {
		c.sha = hex.EncodeToString(b)
		return b, nil
	}
	return nil, fmt.Errorf("Could not hash commit obect: %w", err)
}

func (c *Comm) Type() string {
	return "commit"
}


func (c *Comm) Encode(w io.Writer) error {
	if _, err = fmt.Fprintf(w, "tree %s\n", string(c.treeSha[:])); err != nil {
		return err
	}

	for _, p := range c.parents {
		if _, err = fmt.Fprintf(w, "parent %s\n", string(p[:])); err != nil {
			return err
		}
	}

	if _, err = fmt.Fprintf(w, "author %s\n", c.author.Format()); err != nil {
		return err
	}

	if c.committer != nil {
		if _, err = fmt.Fprintf(w, "committer %s\n", c.committer.Format()); err != nil {
			return err
		}
	}

	if _, err = fmt.Fprintf(w, "\n%s", c.msg); err != nil {
		return err
	}

	return nil
}

func (c *Comm) String() string {
	var s c
}
