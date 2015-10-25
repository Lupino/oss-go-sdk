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
    var OSSAPI = oss.NewAPI(APIOptions)

    var result oss.ListAllMyBucketsResult
    var headers = make(map[string]string)
    var err error
    if err = OSSAPI.GetService(&result, headers); err != nil {
        var e = oss.ParseError(err)
        log.Printf("Code: %s\nMessage: %s\n", e.Code, e.Message)
    }

    log.Printf("result: %v\n", result)
}
