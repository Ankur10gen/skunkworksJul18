package atlasResources

import "testing"

func TestDownloadLogs(t *testing.T) {
	var name string
	username := "username"
	password := "Password"
	groupid := "5a9628574e65810e10fe8f8c"
	logtype := "mongodb"
	requestid := "100"
	startEpoch := "1531734697"
	endEpoch := "1531738697"
	hostname := "cluster0-shard-00-00-vfcl7.mongodb.net"

	name, _ = DownloadLogs(groupid, hostname, logtype, username, password, requestid, startEpoch, endEpoch)
	if name != "../downloadedlogs/mongodb_100.gz" {
		t.Error("Expected a different output than ", name)
	}
}

func TestRunMLogInfo(t *testing.T) {
	var name string
	filename := "/Users/ankurmongodb/go/src/github.com/Ankur10gen/skunkthingy/downloadedlogs/mongodb_100"
	logtype := "mongodb"
	requestid := "100"
	name = RunMLogInfo(logtype, requestid, filename)
	if name != "../downloadedlogs/mongodb_100_mloginfo" {
		t.Error("Expected a different output than ", name)
	}
}
