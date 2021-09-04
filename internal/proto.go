package internal

import (
	"bufio"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func sendPostReq(url, uname, passwd string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, NetworkErr.addContext(err.Error())
	}
	req.SetBasicAuth(uname, passwd)
	cl := http.DefaultClient
	resp, err := cl.Do(req)
	if err != nil {
		return nil, NetworkErr.addContext(err.Error())
	}
	return resp, nil
}

func sendGetReq(url, uname, passwd string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, NetworkErr.addContext(err.Error())
	}
	req.SetBasicAuth(uname, passwd)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, NetworkErr.addContext(err.Error())
	}
	return resp, nil
}

func getRemoteMasterHash(url, uname, passwd string) (string, error) {
	urlstr := url + "/info/refs?service=git-receive-pack"
	resp, err := sendGetReq(urlstr, uname, passwd)
	if err != nil {
		return "", err
	}
	lines := decodePktResp(resp)
	if lines[0] != "# service=git-receive-pack" {
		return "", FormatErr.addContext(fmt.Sprintln("protocol error"))
	}
	if lines[1] != "" {
		return "", FormatErr.addContext(fmt.Sprintln("protocol error"))
	}
	//TODO comeback.seems protocol has changed
	if lines[2][:40] == "0000000000000000000000000000000000000000" {
		//TODO
		return "", FormatErr.addContext(fmt.Sprintln("protocol error"))
	}
	splits := strings.SplitN(lines[2], " ", 2)
	master_sha := splits[0]
	master_ref := strings.SplitN(splits[1], fmt.Sprintf("\\0"), 2)[0]

	if master_ref != "refs/heads/master" {
		return "", FormatErr.addContext(fmt.Sprintln("protocol error"))
	}
	if len(master_sha) != 40 {
		return "", FormatErr.addContext(fmt.Sprintln("sha is bad"))
	}
	return master_sha, nil
}

func decodePktResp(resp *http.Response) []string {
	defer resp.Body.Close()
	lines := []string{}
	rdr := bufio.NewReader(resp.Body)
	scanner := bufio.NewScanner(rdr)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		l := line[0:4]
		b, _ := hex.DecodeString(l)
		len := binary.BigEndian.Uint16(b)
		if len != 0 {
			//-1 allows for the \n we excluded by using ScanLines
			data := line[4 : len-1]
			lines = append(lines, data)
		} else {
			//first append empty string for the demarcator
			lines = append(lines, "")
			line = line[4:]
			b, _ := hex.DecodeString(l)
			len = binary.BigEndian.Uint16(b)
			data := line[4 : len-1]
			lines = append(lines, data)
		}
	}
	return lines
}

func encodePkt(strs []string) string {
	var s strings.Builder
	for _, l := range strs {
		//account for the extra \n char
		s.WriteString(fmt.Sprintf("%4x%s\n", len(l)+5, l))
	}
	s.WriteString("0000")
	return s.String()
}
