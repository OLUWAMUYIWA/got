package internal

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/sergi/go-diff/diffmatchpatch"
)

//great, now I can remove every damn wkdir, I think.
func IsGit() (bool, error) {
	dirs, err := os.ReadDir(".")
	if err != nil {
		return false, OpenErr.addContext(err.Error())
	}
	for _, ent := range dirs {
		if ent.IsDir() && ent.Name() == ".git" {
			return true, nil
		}
	}
	return false, OpenErr
}

func HashFile(name string, w, std bool) error {
	var b []byte
	var err error
	if std {
		b, err = io.ReadAll(os.Stdin)
		if err != nil {
			return CopyErr.addContext(err.Error())
		}
	} else {
		f, err := os.Open(name)
		if err != nil {
			return OpenErr.addContext(err.Error())	
		}	
		var bread bytes.Buffer
		_, err = io.Copy(&bread, f)
		if err != nil {
			return CopyErr.addContext(err.Error())
		}
		b = bread.Bytes()
	}
	got := NewGot()
	var h []byte
	if w {
		h = got.HashObject(b, "blob")	
	} else {
		h = justhash(b)
	}
	if _, err := fmt.Fprintf(os.Stdout, "%x", h); err != nil {
		return IOWriteErr.addContext(err.Error())
	}
	return nil
}

func justhash(data []byte) []byte {
	//create a  new hasher
	hahser := sha1.New()
	//write the bytes to the hasher
	//the hasher write implementation automatically computes the hash written to it
	_, err := hahser.Write(data)
	if err != nil {
		log.Fatalln("could not write data to hasher")
	}
	//to get the hash, we only need to sum a nill byte, which appends a write of its argument(byte slice) to the currently written hash
	//and returns the result
	return hahser.Sum(nil)
}

//getConfig gets into got's config and gets the username and email
//I chose to use json because reading and writing it is clean and easy. No stress
func getConfig() (string, string, error) {
	confRoot, err := os.UserConfigDir()
	if err != nil {
		return "", "", NotDefinedErr.addContext(err.Error())
	}
	f, err := os.Open(filepath.Join(confRoot, ".git", ".config"))
	dec := json.NewDecoder(f)
	var config  ConfigObject
	err = dec.Decode(&config)
	if err != nil {
		return "", "", IoReadErr.addContext(err.Error())
	}
	return config.uname, config.uname, nil
}


func diff(stra, strb string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(stra, strb, true)
	return dmp.DiffPrettyText(diffs)
}

func (got *Got) writeToFile(path string, b []byte) error {
	f, err := os.OpenFile(path, os.O_APPEND, 0)
	defer f.Close()
	got.GotErr(err)
	
	bufWriter := bufio.NewWriter(f)
	_, err = bufWriter.Write(b)
	got.GotErr(err)
	return err
}

func compress(writer io.Writer, data []byte) error {
	comp := zlib.NewWriter(writer)
	_, err := comp.Write(data)
	if err != nil {
		return IOWriteErr.addContext(err.Error())
	}
	err = comp.Flush()
	if err != nil {
		return IOWriteErr.addContext(err.Error())
	}
	return nil
}


func uncompress(rdr io.Reader, dst io.Writer) error {
	comp, err := zlib.NewReader(rdr)
	if err != nil {
		return err
	}
	//to decompress data from zlib, Go is marvelously helpful here, you only have to read from the zlib reader
	//I use copy here because it is cool and fast
	if _,err := io.Copy(dst, comp); err != nil {
		return err
	}
	return nil
}