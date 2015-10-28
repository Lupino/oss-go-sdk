package main

import (
	"bytes"
	"fmt"
	"github.com/Lupino/oss-go-sdk"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
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

	var bucket = "ossgosdknotebook"
	var bookName = "notebook.txt"

	if err = OSSAPI.PutBucket(bucket, oss.ACLPrivate, nil, nil); err != nil {
		log.Printf("%v\n", err)
	}

	var args = os.Args
	if len(args) == 1 {
		readNotebook(OSSAPI, bucket, bookName)
	} else {
		addLine(OSSAPI, bucket, bookName, args[1])
	}
}

func addLine(OSSAPI *oss.API, bucket, bookName, line string) {
	var currentTime = time.Now().Format("2006-01-02 15:04:05")
	var buf = bytes.NewBuffer(nil)
	buf.WriteString(currentTime)
	buf.WriteString("\n")
	buf.WriteString(line)
	buf.WriteString("\n\n")

	var headers = make(map[string]string)
	headers["Content-Type"] = "plain/text"

	var err error

	var headResult http.Header
	var contentLength = 0
	if headResult, err = OSSAPI.HeadObject(bucket, bookName, nil); err == nil {
		contentLength, _ = strconv.Atoi(headResult.Get("Content-Length"))
	}

	if _, err = OSSAPI.AppendObject(bucket, bookName, contentLength, buf, headers); err != nil {
		log.Printf("%v\n", err)
	}
}

func readNotebook(OSSAPI *oss.API, bucket, bookName string) {
	var reader io.ReadCloser
	var err error

	if reader, err = OSSAPI.GetObject(bucket, bookName, nil, nil); err != nil {
		log.Printf("%v\n", err)
		return
	}
	defer reader.Close()
	var data, _ = ioutil.ReadAll(reader)
	fmt.Printf("%s\n", data)
}
