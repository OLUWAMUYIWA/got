package pkg

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode"
)

//!!!!COMMIT OBJECT AND ITS GOTOBJECT IMPLEMENTATION!!!!!
type commitObj struct {
	sha  string
	data []byte
}


type Sign struct {
	name, email string
	time time.Time
}

type Comm struct {
	sha [20]byte
	treeSha string
	parents [][20] byte
	committer Sign
	author Sign
	msg string
	pgp string
}


func parseCommit(rdr io.Reader, got *Got) (*Comm, error) {
	//TODO
	scanner := bufio.NewScanner(rdr)
	scanner.Split(bufio.ScanLines)

	comm := &Comm{}
	for scanner.Scan() {
		line := scanner.Bytes()
		tLine := bytes.TrimRight(line, "\n") //is it "/n" or "/r/n" 
		splits := bytes.FieldsFunc(tLine, func(r rune) bool {
			return unicode.IsSpace(r)
		})

		prefix := string(splits[0])

		switch prefix {
			case "tree": {
				if len(splits) != 2 {
					return nil, fmt.Errorf("Tree line in commi object faulty")
				}
				treeSha := splits[1]
				comm.treeSha = string(treeSha)
			}

			case "parent": {
				if len(splits) = 2 || len(splits[2]) != 20 {
					return nil, fmt.Errorf("Parent line in commit faulty")
				}
				p := splits[1]
				parent := (*[20]byte)(p) //comeback to this cool hack. its a tip. it panics if the lengths aof slice and array are different
				c.parents = append(c.parents, *parent)
			}

			case "author": {
				sauth := Sign{}
				rem := splits[1:]
				auth.name = rem[0]
				auth.email = string(bytes.TrimFunc(rem[1], func(r rune) bool {
					return r == '<' || r == '>'
				}))
				//comeback for the time parsing
				timestamp, err  := strconv.ParseInt(string(rem[2]), 10, 0)
				if err != nil {
					return nil, err
				}

				zoneHr, err := strconv.ParseInt(string(rem[3])[0:3], 10, 0)
				if err != nil {
					return nil, err
				}
				zoneMin, err := strconv.ParseInt(string(rem[3])[3:], 10, 0)
				if err != nil {
					return nil, err
				}

				if zoneHr < 0 {
					zoneMin *= -1
				}
				unixTime := time.Unix(timestamp, 0).In(time.FixedZone("", zoneHr * 60 * 60 + zoneMin * 60)).Format("Wed Jan _5 15:04:05 2010")
				
				fullUnixTime := fmt.Sprintf("%s %s", unixTime, zone)
				time, err := time.Parse("Mon Dec _5 15:04:05 2010 -0700", fullUnixTime)
				if err != nil {
					return nil, fmt.Errorf("Error parsing commit time")
				}
				auth.time = time

			}

			case "committer": {

			}


		
		}
	}
	
	c := &Comm{}
	return c, nil
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

func (c *Comm) Tree() {
	
}