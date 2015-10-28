package main

import (
	"github.com/Lupino/oss-go-sdk"
	"log"
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

	var result oss.ListAllMyBucketsResult
	var headers = make(map[string]string)

	if err = OSSAPI.GetService(&result, headers); err != nil {
		log.Printf("%s\n", err)
	}

	log.Printf("result: %v\n", result)
}
