package proto

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

)

//where the client and server determine what the minimal packfile necessary for transport is
func NegotiatePkFile() {
	//tell the server 
	//send wants
	//send shallows
	//send optional commit depth
	//send flush packet to inform the server we're done
	//no point specifying commitdepth. we will skip that
	//send have lines
	//what we get in return are packfiles
	//send 32 haves per time, before a flush packet. then send the next 32, until you finish
	//server responds with 'ACK obj-id continue' for every 'have' that we send and it has
	//once the server has found an acceptable common base commit and is ready to make a packfile, it will blindly ACK all have obj-ids back to the client.
	//server sends a 'NAK', to which the client responds with a 'done' or another list of 'have' lines
	//the momemt the client receives enough 'ACK' to color every packetline it has sent as an have, or when it 
	//gives up because it has sent 256 have lines without getting any of them ACKed by the server, it sends 'done'
	//when server receives 'done', it sends a 'ACK obj-id' (obj-id of the latest commit they share) or a 'NAK',
	//server sends 'NAK' if there is no common base  

}

func sendReq(serviceName string) func(body io.Reader) {
	return nil
}
func sendPostReq(url, uname, passwd string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, body)
	if err != nil {
		return nil, &ProtoErr{ErrString: "While sending post request", inner: err}
	}
	req.SetBasicAuth(uname, passwd)
	cl := http.DefaultClient
	resp, err := cl.Do(req)
	if err != nil {
		return nil, &ProtoErr{ErrString: "While exec-ing post request", inner: err}
	}
	return resp, nil
}

func sendGetReq(url, uname, passwd string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, &ProtoErr{ErrString: "While exec-ing get request", inner: err}
	}
	req.SetBasicAuth(uname, passwd)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, &ProtoErr{ErrString: "While exec-ing get request", inner: err}
	}
	return resp, nil
}

func GetRemoteMasterHash(url, uname, passwd string) (string, error) {
	urlstr := url + "/info/refs?service=git-receive-pack"
	resp, err := sendGetReq(urlstr, uname, passwd)
	if err != nil {
		return "", err
	}
	lines := decodePktResp(resp)
	if lines[0] != "# service=git-receive-pack" {
		return "", &ProtoErr{ErrString: "While getting remote master hash", inner: fmt.Errorf("Protocol error")}
	}
	if lines[1] != "" {
		return "", &ProtoErr{ErrString: "While getting remote master hash", inner: fmt.Errorf("Protocol error")}
	}
	//TODO comeback.seems protocol has changed
	if lines[2][:40] == "0000000000000000000000000000000000000000" {
		//TODO
		return "", &ProtoErr{ErrString: "While getting remote master hash", inner: fmt.Errorf("Protocol error")}
	}
	splits := strings.SplitN(lines[2], " ", 2)
	master_sha := splits[0]
	master_ref := strings.SplitN(splits[1], fmt.Sprintf("\\0"), 2)[0]

	if master_ref != "refs/heads/master" {
		return "", &ProtoErr{ErrString: "While getting remote master hash", inner: fmt.Errorf("Protocol error")}
	}
	if len(master_sha) != 40 {
		return "", &ProtoErr{ErrString: "While getting remote master hash", inner: fmt.Errorf("SHA is bad")}
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
	for _, _ =  range lines {
		buff.WriteString(fmt.Sprintf(""))
	}
	//TODO
	return nil
}