package http

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/OLUWAMUYIWA/got/internal/proto"
)

var sep byte = 0
var nullstr = hex.EncodeToString([]byte{sep})
var zeroId = hex.EncodeToString(make([]byte, 20))
var flushPacket = fmt.Sprintf("%.4x", 0)

//https://github.com/git/git/blob/master/Documentation/technical/protocol-common.txt
// ----
//   pkt-line     =  data-pkt / flush-pkt

//   data-pkt     =  pkt-len pkt-payload
//   pkt-len      =  4*(HEXDIG)
//   pkt-payload  =  (pkt-len - 4)*(OCTET)

//   flush-pkt    = "0000"
// ----

type PktLine struct {
	len     int16
	id      string
	refname string
	payload []byte
}

//pkt-line stream describing each ref and its current value
type PktStream struct {
	capabilities []string
	stream       []PktLine
}

func decodePkts(stream []byte) PktStream {
	return PktStream{}
}

func (stream PktStream) encodePkts() {

}

func negotiatePkFile() {

}

// func sendReq(serviceName string) func(body io.Reader) {

// }

func sendPostReq(url, uname, passwd string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, body)
	if err != nil {
		return nil, &proto.ProtoErr{
			Context: "Network Error",
			Inner:   err,
		}
	}
	req.SetBasicAuth(uname, passwd)
	cl := http.DefaultClient
	resp, err := cl.Do(req)
	if err != nil {
		return nil, &proto.ProtoErr{
			Context: "Network Error",
			Inner:   err,
		}
	}
	return resp, nil
}

func sendGetReq(url, uname, passwd string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, &proto.ProtoErr{
			Context: "Network Error",
			Inner:   err,
		}
	}
	req.SetBasicAuth(uname, passwd)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, &proto.ProtoErr{
			Context: "Network Error",
			Inner:   err,
		}
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
		return "", &proto.ProtoErr{
			Context: "Network Error",
			Inner:   err,
		}
	}
	if lines[1] != "" {
		return "", &proto.ProtoErr{
			Context: "Network Error",
			Inner:   err,
		}
	}
	//TODO comeback.seems protocol has changed
	if lines[2][:40] == "0000000000000000000000000000000000000000" {
		//TODO
		return "", &proto.ProtoErr{
			Context: "Network Error",
			Inner:   err,
		}
	}
	splits := strings.SplitN(lines[2], " ", 2)
	master_sha := splits[0]
	master_ref := strings.SplitN(splits[1], fmt.Sprintf("\\0"), 2)[0]

	if master_ref != "refs/heads/master" {
		return "", &proto.ProtoErr{
			Context: "Network Error",
			Inner:   err,
		}
	}
	if len(master_sha) != 40 {
		return "", &proto.ProtoErr{
			Context: "SHA is bad",
		}
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

func buildLines(lines []byte) []byte {
	buff := new(bytes.Buffer)
	for _, l := range lines {
		buff.WriteString(fmt.Sprintf(""))
		fmt.Println(l)
	}
	//undone
	return []byte{0}
}
