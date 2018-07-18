package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/Ankur10gen/skunkthingy/atlasResources"
	"github.com/Ankur10gen/skunkthingy/s3upload"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type infoDoc struct {
	RequestID  bson.ObjectId `bson:"_id"`
	State      string        `bson:"state"`
	Hostname   string        `bson:"hostname"`
	GroupID    string        `bson:"groupid"`
	StartEpoch string        `bson:"start_epoch"`
	EndEpoch   string        `bson:"end_epoch"`
	LogType    string        `bson:"log_type"`
	Username   string        `bson:"username"`
	Password   string        `bson:"password"`
}

type resultDoc struct {
	RequestID     bson.ObjectId
	OutputDocPath string
}

var infoColl *mgo.Collection

func init() {
	uri := "mongodb://vaidyauser:vad1234@cluster0-shard-00-00-vfcl7.mongodb.net:27017,cluster0-shard-00-01-vfcl7.mongodb.net:27017,cluster0-shard-00-02-vfcl7.mongodb.net:27017/test?replicaSet=Cluster0-shard-0&authSource=admin"
	dialInfo, err := mgo.ParseURL(uri)
	if err != nil {
		log.Fatalln(err)
	}

	tlsConfig := &tls.Config{}
	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
		return conn, err
	}
	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		log.Fatalln(err)
	}

	infoColl = session.DB("vaidya").C("info")

	fmt.Println("Connection to database done")
}

func main() {

	requestsChan := make(chan infoDoc, 100) //New Requests go here
	outputChan := make(chan resultDoc, 100) //Path after analysis goes here
	statusChan := make(chan string, 100)    //Final status is given here for logging purposes

	// Three Go routines
	go FetchPendingRequests(requestsChan)

	go DownloadLogsAndAnalyse(requestsChan, outputChan)

	go UploadOutputAndUpdateDatabase(outputChan, statusChan)

	for val := range statusChan {
		log.Println(val)
	}
}

// UploadOutputAndUpdateDatabase reads the result docs
// and uploads the ouput files to S3
// finally updates the database
func UploadOutputAndUpdateDatabase(ch chan resultDoc, statusChan chan string) {
	for val := range ch {
		var rd resultDoc
		rd = val

		fmt.Println("Starting Upload to S3")

		returnCode := s3upload.UploadToS3(rd.OutputDocPath)
		if returnCode != 1 {
			statusChan <- fmt.Sprintf("File upload failed for %s after bad return code", rd.RequestID)
			break
		}

		fmt.Println("S3 Upload Completed")
		fmt.Println("Updating Collection on Atlas")

		err := infoColl.Update(bson.M{"_id": rd.RequestID}, bson.M{"$set": bson.M{"status": "done", "s3ObjectPath": rd.OutputDocPath}})
		if err != nil {
			fmt.Println(err)
			statusChan <- fmt.Sprintf("File upload failed for %s after trying to update db", rd.RequestID)
			break
		}

		fmt.Println("Collection Updated")

		statusChan <- fmt.Sprintf("File upload done for %s", rd.RequestID)
	}
}

// DownloadLogsAndAnalyse reads requests from the request channel
// and facilitates the download and analysis of logs
// It pushes the result document on another channel
func DownloadLogsAndAnalyse(ch chan infoDoc, opchan chan resultDoc) {
	for val := range ch {
		var ri infoDoc
		ri = val

		fmt.Printf("Starting Download of Logs for %s \n", ri.RequestID)

		fmt.Println("====", ri.StartEpoch, ri.EndEpoch, "====")

		logpath, err := atlasResources.DownloadLogs(
			ri.GroupID,
			ri.Hostname,
			ri.LogType,
			ri.Username,
			ri.Password,
			ri.RequestID.Hex(),
			ri.StartEpoch,
			ri.EndEpoch)

		if err != nil {
			log.Fatalln(err)
		}

		resultPath := atlasResources.AnalyseLogs(ri.LogType, ri.RequestID.Hex(), logpath)

		fmt.Printf("Analysed logs for request ID %s on path %s \n", ri.RequestID, resultPath.MLoginfoOutput)

		opchan <- resultDoc{RequestID: ri.RequestID, OutputDocPath: resultPath.MLoginfoOutput}
	}
}

// FetchPendingRequests fetches new requests from the MongoDB Atlas instance
// and puts them on a channel
func FetchPendingRequests(ch chan infoDoc) {
	var lastID bson.ObjectId
	for {
		fmt.Println("Fetching new requests from database")
		var infodocs []infoDoc
		if lastID == "" {
			_ = infoColl.Find(bson.M{"status": "pending"}).All(&infodocs)
		} else {
			_ = infoColl.Find(bson.M{"status": "pending", "_id": bson.M{"$gt": lastID}}).All(&infodocs)
		}
		for _, val := range infodocs {
			fmt.Printf("Got Request ID %s", val.RequestID)
			ch <- val
			lastID = val.RequestID
		}
		fmt.Println("What a busy day! Sleeping for 30 seconds now.")
		time.Sleep(30 * time.Second)
	}
}
