package proto

import (
	"fmt"
	"strings"
)


// Example of valid refspec: "+refs/heads/*:refs/remotes/origin/*"
type RefSpecRaw string

type RefSpec struct {
	ForceUp, Delete bool
	Src, Dst string
}

func (r *RefSpecRaw) Parse() (*RefSpec, error) {

	//split with the seperator ':'
	splits := strings.FieldsFunc(string(*r), func(r rune) bool {
		return r == ':'
	})

	//if more than two seperators exist, refspec is bad
	if len(splits) > 2 {
		return nil, fmt.Errorf("Bad Refspec")
	}

	res := &RefSpec{}
	
	src, dst := "", ""

	// if splits is just one, then `src` does not exist, and refspec is delete
	if len(splits) == 2 {
		src = splits[0]
		dst = splits[2]
	} else {
		dst = splits[1]
		res.Delete = true
	}

	forceUp := false

	if src[0] == '+' {
		forceUp = true
		src = strings.TrimLeft(src, "+")
	}

	wildCarsS := strings.Count(src, "*")
	wildCardD := strings.Count(dst, "*")
	if wildCardD != wildCarsS || wildCardD > 1 || wildCarsS > 1 {
		return nil, fmt.Errorf("Wildcard in refspec error")
	}

	res.Src = src
	res.Dst = dst
	res.ForceUp = forceUp

	return res, nil
}

