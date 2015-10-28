package main

import (
	"github.com/Lupino/oss-go-sdk"
	"io"
	"log"
	"os"
)

// AccessKeyID defined Access Key ID
const AccessKeyID = ""

// AccessKeySecret defined Access Key Secret
const AccessKeySecret = ""

func main() {
	var APIOptions = oss.GetDefaultAPIOptioins()
	APIOptions.AccessID = AccessKeyID
	APIOptions.SecretAccessKey = AccessKeySecret
	var OSSAPI = oss.NewAPI(APIOptions)

	var bucket = "ossgosdkfileuploaddownload"
	var err error

	if err = OSSAPI.PutBucket(bucket, oss.ACLPublicReadWrite, nil, nil); err != nil {
		log.Printf("%v\n", err)
	}

	upload(OSSAPI, bucket, "test.jpg", "image/jpeg")

	download(OSSAPI, bucket, "test.jpg", "test-download.jpg")
	log.Println("success.")
}

func upload(OSSAPI *oss.API, bucket, filename, contentType string) {
	fp, err := os.Open(filename)
	defer fp.Close()
	if err != nil {
		log.Fatal(err)
	}

	var headers = make(map[string]string)
	headers["Content-Type"] = contentType

	if err = OSSAPI.PutObject(bucket, filename, fp, headers); err != nil {
		log.Printf("%v\n", err)
	}
}

func download(OSSAPI *oss.API, bucket, object, filename string) {
	var fp, err = os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()
	var reader io.ReadCloser
	if reader, err = OSSAPI.GetObject(bucket, object, nil, nil); err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	io.Copy(fp, reader)
}
