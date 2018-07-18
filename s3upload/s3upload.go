package s3upload

import (
	"bytes"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	S3_REGION = "ap-south-1"
	S3_BUCKET = "vaidyamumbai"
)

func UploadToS3(fileName string) uint8 {
	s, err := session.NewSession(&aws.Config{Region: aws.String(S3_REGION)})
	if err != nil {
		log.Fatalln(err)
	}

	err = AddFileToS3(s, fileName)
	if err != nil {
		log.Fatalln(err)
	}
	return 1
}

func AddFileToS3(s *session.Session, fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	var size = fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(S3_BUCKET),
		Key:                  aws.String(fileName),
		ACL:                  aws.String("private"),
		Body:                 bytes.NewReader(buffer),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(http.DetectContentType(buffer)),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
	})
	return err
}
