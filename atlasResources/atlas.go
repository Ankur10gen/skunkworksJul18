package atlasResources

import (
	"github.com/kr/pty"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

import dac "github.com/xinsnake/go-http-digest-auth-client"

var results Results

// Results is a complete set of paths to the various tests
type Results struct {
	MLoginfoOutput string
}

// DownloadLogs creates a file with the downloaded logs and appends
// the request id with name
func DownloadLogs(groupid, hostname, logtype, username, password, requestID, startEpoch, endEpoch string) (string, error) {

	// Construct the uri
	uri := "https://cloud.mongodb.com/api/atlas/v1.0" + "/groups/" +
		groupid + "/clusters/" + hostname + "/logs/" + logtype + ".gz" +
		"?startDate=" + startEpoch + "&endDate=" + endEpoch

	// Create output file
	out, err := os.Create("downloadedlogs/" + logtype + "_" + requestID + ".gz")
	if err != nil {
		log.Fatalln(err)
	}
	defer out.Close()

	t := dac.NewTransport(username, password)
	req, err := http.NewRequest("GET", uri, nil)

	if err != nil {
		log.Fatalln(err)
	}

	resp, err := t.RoundTrip(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	return out.Name(), nil
}

// AnalyseLogs runs mtools
// for now only mloginfo
func AnalyseLogs(logtype, requestID, path string) Results {
	mloginfoOutput := RunMLogInfo(logtype, requestID, path)
	results.MLoginfoOutput = mloginfoOutput
	return results
}

// RunMLogInfo returns the path of file containing mloginfo
func RunMLogInfo(logtype, requestID, path string) string {
	cmd := exec.Command("gunzip", path)
	tty, err := pty.Start(cmd)
	if err != nil {
		log.Fatalln(err)
	}

	cmd = exec.Command("mloginfo", strings.Trim(path, ".gz"))
	tty, err = pty.Start(cmd)
	if err != nil {
		log.Fatalln(err)
	}
	defer tty.Close()
	f, err := os.Create("downloadedlogs/" + logtype + "_" + requestID + "_mloginfo")
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	io.Copy(f, tty)

	return f.Name()
}
