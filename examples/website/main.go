package main

import (
	"github.com/Lupino/oss-go-sdk"
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
	var OSSAPI, err = oss.NewAPI(APIOptions)

	if err != nil {
		log.Fatal(err)
	}

	var bucket = "ossgosdkwebsite"

	if err = OSSAPI.PutBucket(bucket, oss.ACLPublicReadWrite, nil, nil); err != nil {
		log.Printf("%s\n", err)
	}

	if err = OSSAPI.PutBucketWebsite(bucket, "index.html", "error.html"); err != nil {
		log.Printf("%s\n", err)
	}

	upload(OSSAPI, bucket, "index.html")
	upload(OSSAPI, bucket, "error.html")
	upload(OSSAPI, bucket, "css/screen.css")
	upload(OSSAPI, bucket, "js/application.js")
	log.Println("success.")
}

func upload(OSSAPI *oss.API, bucket, filename string) {
	fp, err := os.Open(filename)
	defer fp.Close()
	if err != nil {
		log.Fatal(err)
	}

	var headers = make(map[string]string)
	headers["Content-Type"] = "text/html"

	if err = OSSAPI.PutObject(bucket, filename, fp, headers); err != nil {
		log.Printf("%s\n", err)
	}
}
