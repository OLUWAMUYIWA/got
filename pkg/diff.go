package pkg

import (
	"fmt"
	"io"
	"os"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
)

// diff takes 2 file paths
func diff(a, b string) (string, error) {
	f1, err := os.Open(a)
	if err != nil {
		return "", err
	}
	f2, err := os.Open(b)
	if err != nil {
		return "", err
	}
	var str1, str2 string
	if by, err := io.ReadAll(f1); err != nil {
		return "", err
	} else {
		str1 = string(by)
	}
	if by, err := io.ReadAll(f2); err != nil {
		return "", err
	} else {
		str2 = string(by)
	}
	edits := myers.ComputeEdits(span.URIFromPath(a), str1, str2)
	diff := fmt.Sprint(gotextdiff.ToUnified(a, b, str1, edits))
	return diff, nil
}
