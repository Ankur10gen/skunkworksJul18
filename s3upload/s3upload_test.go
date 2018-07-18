package s3upload

import (
	"testing"
)

func TestUploadToS3(t *testing.T) {
	var retcode uint8
	retcode = UploadToS3("../downloadedlogs/mongodb_100")
	if retcode != 1 {
		t.Error("Upload didn't succeed", retcode)
	}
}
