package main

import (
	"bufio"
	"bytes"
	"flag"
	"github.com/Lupino/oss-go-sdk"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// AccessKeyID defined Access Key ID
var AccessKeyID string

// AccessKeySecret defined Access Key Secret
var AccessKeySecret string

// OSSAPI defined oss api
var OSSAPI *oss.API

func init() {
	flag.StringVar(&AccessKeyID, "key", "", "AccessKeyID")
	flag.StringVar(&AccessKeySecret, "secret", "", "AccessKeySecret")
	flag.Parse()
}

func main() {
	var APIOptions = oss.GetDefaultAPIOptioins()
	APIOptions.AccessID = AccessKeyID
	APIOptions.SecretAccessKey = AccessKeySecret
	OSSAPI = oss.NewAPI(APIOptions)
	var result oss.ListAllMyBucketsResult
	var err error

	var bucket = "mainosstest"
	var loggingBucket = "loggingbucket"

	log.Println("GetService from oss server")
	if err = OSSAPI.GetService(&result, nil); err != nil {
		log.Printf("GetService Error: %s\n", err)
	}
	log.Printf("GetService result: %s\n", result.Buckets)

	log.Println("PutBucket")
	if err = OSSAPI.PutBucket(bucket, oss.ACLPublicReadWrite, nil, nil); err != nil {
		log.Printf("PutBucket Error: %s\n", err)
	}

	log.Println("PutBucket")
	if err = OSSAPI.PutBucket(loggingBucket, oss.ACLPublicReadWrite, nil, nil); err != nil {
		log.Printf("PutBucket Error: %s\n", err)
	}

	log.Println("GetBucket")
	var result1 oss.ListBucketResult
	if err = OSSAPI.GetBucket(bucket, &result1, nil); err != nil {
		log.Printf("GetBucket Error: %s\n", err)
	}
	log.Printf("GetBucket result: %s\n", result1.Name)

	log.Printf("PutBucketACL")
	if err = OSSAPI.PutBucketACL(bucket, oss.ACLPublicReadWrite, nil); err != nil {
		log.Printf("PutBucketACL Error: %s\n", err)
	}

	log.Println("GetBucketACL")
	var result2 oss.AccessControlPolicy
	if err = OSSAPI.GetBucketACL(bucket, &result2); err != nil {
		log.Printf("GetBucketACL Error: %s\n", err)
	}
	log.Printf("GetBucketACL result: %s\n", result2.AccessControlList)

	var result4 oss.LocationConstraint
	log.Println("GetBucketLocation")
	if err = OSSAPI.GetBucketLocation(bucket, &result4); err != nil {
		log.Printf("GetBucketLocation Error: %s\n", err)
	}
	log.Printf("GetBucketLocation result: %s\n", result4)

	log.Printf("PutBucketLogging")
	if err = OSSAPI.PutBucketLogging(bucket, loggingBucket, "bucket.log"); err != nil {
		log.Printf("PutBucketLogging Error: %s\n", err)
	}

	log.Println("GetBucketLogging")
	var result5 oss.BucketLoggingStatus
	if err = OSSAPI.GetBucketLogging(bucket, &result5); err != nil {
		log.Printf("GetBucketLogging Error: %s\n", err)
	}
	log.Printf("GetBucketLogging result: %s\n", result5)

	log.Printf("PutBucketWebsite")
	if err = OSSAPI.PutBucketWebsite(bucket, "index.html", "error.html"); err != nil {
		log.Printf("PutBucketWebsite Error: %s\n", err)
	}

	log.Println("GetBucketWebsite")
	var result3 oss.WebsiteConfiguration
	if err = OSSAPI.GetBucketWebsite(bucket, &result3); err != nil {
		log.Printf("GetBucketWebsite Error: %s\n", err)
	}
	log.Printf("GetBucketWebsite result: %s\n", result3)

	log.Println("PutBucketReferer")
	var refererConfig = oss.RefererConfiguration{
		AllowEmptyReferer: true,
		RefererList:       []string{"http://test.com", "http://example.com"},
	}
	if err := OSSAPI.PutBucketReferer(bucket, refererConfig); err != nil {
		log.Printf("PutBucketReferer Error: %s\n", err)
	}

	log.Println("GetBucketReferer")
	var result6 oss.RefererConfiguration
	if err = OSSAPI.GetBucketReferer(bucket, &result6); err != nil {
		log.Printf("GetBucketReferer Error: %s\n", err)
	}
	log.Printf("GetBucketReferer result: %s\n", result6)

	log.Printf("PutBucketLifecycle")
	var lcRule = oss.LifecycleRule{
		ID:             "ID",
		Prefix:         "Prefix",
		Status:         "Enabled",
		ExpirationDays: 1,
	}
	if err = OSSAPI.PutBucketLifecycle(bucket, lcRule); err != nil {
		log.Printf("PutBucketLifecycle Error: %s\n", err)
	}

	log.Printf("GetBucketLifecycle")
	var result7 oss.LifecycleConfiguration
	if err = OSSAPI.GetBucketLifecycle(bucket, &result7); err != nil {
		log.Printf("GetBucketLifecycle Error: %s\n", err)
	}
	log.Printf("GetBucketLifecycle result: %s\n", result7)

	log.Println("DeleteBucketLogging")
	if err = OSSAPI.DeleteBucketLogging(bucket); err != nil {
		log.Printf("DeleteBucketLogging Error: %s\n", err)
	}

	log.Println("DeleteBucketWebsite")
	if err = OSSAPI.DeleteBucketWebsite(bucket); err != nil {
		log.Printf("DeleteBucketWebsite Error: %s\n", err)
	}

	log.Println("DeleteBucketLifecycle")
	if err = OSSAPI.DeleteBucketLifecycle(bucket); err != nil {
		log.Printf("DeleteBucketLifecycle Error: %s\n", err)
	}
	////////////////////////////////////////////////////////////////////////////

	var object = "object.data"
	var body = bytes.NewBufferString("this is the body")
	var contentType = "plan/text"
	var headers = make(map[string]string)
	headers["Content-Type"] = contentType

	log.Println("PutObject")
	if err = OSSAPI.PutObject(bucket, object, bufio.NewReader(body), headers); err != nil {
		log.Printf("PutObject Error: %s\n", err)
	}

	log.Println("PutObject")
	fp, err := os.Open("main.go")
	if err != nil {
		log.Fatal(err)
	}
	if err = OSSAPI.PostObject(bucket, object, fp, headers); err != nil {
		log.Printf("PutObject Error: %s\n", err)
	}

	log.Println("GetObject")
	var data io.ReadCloser
	if data, err = OSSAPI.GetObject(bucket, object, nil, nil); err != nil {
		log.Printf("GetObject Error: %s\n", err)
	} else {
		defer data.Close()
		var buf, _ = ioutil.ReadAll(data)
		log.Printf("GetObject result: the data length is: %d\n", len(buf))
	}

	log.Println("PutObjectACL")
	if err = OSSAPI.PutObjectACL(bucket, object, "public-read"); err != nil {
		log.Printf("PutObjectACL Error: %s\n", err)
	}

	log.Println("GetObjectACL")
	var acl oss.AccessControlPolicy
	if err = OSSAPI.GetObjectACL(bucket, object, &acl); err != nil {
		log.Printf("GetObjectACL Error: %s\n", err)
	}
	log.Printf("GetObjectACL result: %s\n", acl.AccessControlList)

	log.Println("HeadObject")
	var headResult http.Header
	if headResult, err = OSSAPI.HeadObject(bucket, object, nil); err != nil {
		log.Printf("HeadObject Error: %s\n", err)
	}
	log.Printf("HeadObject result: %s\n", headResult)

	log.Println("DeleteObject")
	if err = OSSAPI.DeleteObject(bucket, object); err != nil {
		log.Printf("DeleteObject Error: %s\n", err)
	}

	log.Println("DeleteObjects")
	var deleteResult oss.DeleteResult
	if err = OSSAPI.DeleteObjects(bucket, []string{"object1", "object2"}, &deleteResult); err != nil {
		log.Printf("DeleteObjects Error: %s\n", err)
	}
	log.Printf("DeleteObjects result: %s\n", deleteResult)

	log.Println("AppendObject")
	var appendObject = "appendObject.data"
	var appendObject1 = "appendObject1.data"
	var etag http.Header
	body = bytes.NewBufferString("this is the body")
	if etag, err = OSSAPI.AppendObject(bucket, appendObject, 0, bufio.NewReader(body), headers); err != nil {
		log.Printf("AppendObject Error: %s\n", err)
	}
	log.Printf("AppendObject result: %s\n", etag)

	log.Println("AppendObject")
	fp, err = os.Open("main.go")
	if err != nil {
		log.Fatal(err)
	}
	if etag, err = OSSAPI.AppendObject(bucket, appendObject, 1, fp, headers); err != nil {
		log.Printf("AppendObject Error: %s\n", err)
	}
	log.Printf("AppendObject result: %s\n", etag)

	log.Println("CopyObject")
	if _, err = OSSAPI.CopyObject(bucket, appendObject, bucket, appendObject1, headers); err != nil {
		log.Printf("CopyObject Error: %s\n", err)
	}

	log.Println("DeleteObjects")
	if err = OSSAPI.DeleteObjects(bucket, []string{appendObject, appendObject1}, &deleteResult); err != nil {
		log.Printf("DeleteObjects Error: %s\n", err)
	}
	log.Printf("DeleteObjects result: %s\n", deleteResult)

	////////////////////////////////////////////////////////////////////////////

	log.Println("DeleteBucket")
	if err = OSSAPI.DeleteBucket(bucket); err != nil {
		log.Printf("DeleteBucket Error: %s\n", err)
	}

	log.Println("DeleteBucket")
	if err = OSSAPI.DeleteBucket(loggingBucket); err != nil {
		log.Printf("DeleteBucket Error: %s\n", err)
	}

}
