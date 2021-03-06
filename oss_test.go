package oss

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
)

var api *API
var options *APIOptions

func init() {
	var ts = mockHTTPServer()
	options = GetDefaultAPIOptioins()
	options.Host, options.Port = getHostFromURL(ts.URL)
	api, _ = NewAPI(options)
}

func TestPrint(t *testing.T) {
	fmt.Println(AGENT)
}

func TestSetValue(t *testing.T) {
	var options = GetDefaultAPIOptioins()
	var api, _ = NewAPI(options)
	api.SetTimeout(5)
	api.SetDebug()
	api.SetRetryTimes(5)
	api.SetIsOSSHost(false)
}

func getHostFromURL(uri string) (string, int) {
	var u, _ = url.Parse(uri)
	return getHostPort(u.Host)
}

func getHostPort(origHost string) (host string, port int) {
	host = origHost
	port = 80
	var hostPortList = strings.SplitN(origHost, ":", 2)
	var err error
	if len(hostPortList) == 1 {
		host = strings.Trim(hostPortList[0], " ")
	} else if len(hostPortList) == 2 {
		host = strings.Trim(hostPortList[0], " ")
		if port, err = strconv.Atoi(strings.Trim(hostPortList[1], " ")); err != nil {
			panic("Invalid: port is invalid")
		}
	}
	return
}

func TestHttpRequest(t *testing.T) {
	var err error
	var reqOptions = new(requestOptions)
	reqOptions.Method = "GET"
	tls := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer tls.Close()

	var options = GetDefaultAPIOptioins()
	options.Host, options.Port = getHostFromURL(tls.URL)
	options.IsSecurity = true
	options.StsToken = "sts_token"
	var api, _ = NewAPI(options)

	_, err = api.httpRequest(reqOptions)
	fmt.Printf("%v\n", err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	options = GetDefaultAPIOptioins()
	options.Host, options.Port = getHostFromURL(ts.URL)
	api, _ = NewAPI(options)
	_, err = api.httpRequest(reqOptions)
	fmt.Printf("%v\n", err)

	var acl AccessControlPolicy
	err = api.httpRequestWithUnmarshalXML(reqOptions, &acl)
}

func TestGetService(t *testing.T) {
	var result ListAllMyBucketsResult
	var err error
	result.Prefix = "prefix"
	result.Marker = "marker"
	result.MaxKeys = "max-keys"
	if err = api.GetService(&result, nil); err != nil {
		t.Fatal(err)
	}
}

func TestGetBucket(t *testing.T) {
	var result ListBucketResult
	var err error
	if err = api.GetBucket("bucket", &result, nil); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s\n", result)
}

func TestGetBucketACL(t *testing.T) {
	var result AccessControlPolicy
	var err error
	if err = api.GetBucketACL("bucket", &result); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s\n", result)
}

func TestGetBucketLocation(t *testing.T) {
	var result LocationConstraint
	var err error
	if err = api.GetBucketLocation("bucket", &result); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s\n", result)
}

func TestGetBucketLogging(t *testing.T) {
	var result BucketLoggingStatus
	var err error
	if err = api.GetBucketLogging("bucket", &result); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s\n", result)
}

func TestGetBucketWebsite(t *testing.T) {
	var result WebsiteConfiguration
	var err error
	if err = api.GetBucketWebsite("bucket", &result); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s\n", result)
}

func TestGetBucketReferer(t *testing.T) {
	var result RefererConfiguration
	var err error
	if err = api.GetBucketReferer("bucket", &result); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s\n", result)
}

func TestGetBucketLifecycle(t *testing.T) {
	var result LifecycleConfiguration
	var err error
	if err = api.GetBucketLifecycle("bucket", &result); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s\n", result)
}

func TestCreateBucket(t *testing.T) {
	var err error
	if err = api.CreateBucket("bucket", "", nil); err != nil {
		t.Fatal(err)
	}
	if err = api.CreateBucket("bucket", "public-read-write", map[string]string{"other": "other"}); err != nil {
		t.Fatal(err)
	}
	if err = api.PutBucket("bucket", ACLPublicReadWrite, "oss-cn-hangzhou", nil); err != nil {
		t.Fatal(err)
	}

	if err = api.PutBucketACL("bucket", "public-read-write", map[string]string{"other": "other"}); err != nil {
		t.Fatal(err)
	}
}

func TestPutBucketReferer(t *testing.T) {
	var config = RefererConfiguration{
		AllowEmptyReferer: true,
		RefererList:       []string{"http://test.com", "http://example.com"},
	}
	if err := api.PutBucketReferer("bucket", config); err != nil {
		t.Fatal(err)
	}
}

func TestPutBucketWebsite(t *testing.T) {
	if err := api.PutBucketWebsite("bucket", "index.html", "error.html"); err != nil {
		t.Fatal(err)
	}
}

func TestPutBucketLifecycle(t *testing.T) {
	var rule = LifecycleRule{
		ID:             "ID",
		Prefix:         "Prefix",
		Status:         "Status",
		ExpirationDays: 1,
	}
	if err := api.PutBucketLifecycle("bucket", rule); err != nil {
		t.Fatal(err)
	}
}

func TestPutBucketLogging(t *testing.T) {
	if err := api.PutBucketLogging("bucket", "bucket1", "eaa"); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteBucket(t *testing.T) {
	if err := api.DeleteBucket("bucket"); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteBucketLifecycle(t *testing.T) {
	if err := api.DeleteBucketLifecycle("bucket"); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteBucketWebsite(t *testing.T) {
	if err := api.DeleteBucketWebsite("bucket"); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteBucketLogging(t *testing.T) {
	if err := api.DeleteBucketLogging("bucket"); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteBucketReferer(t *testing.T) {
	if err := api.DeleteBucketReferer("bucket"); err != nil {
		t.Fatal(err)
	}
}

func TestSignURL(t *testing.T) {
	var options = GetDefaultSignURLOptions()
	var signURL = api.SignURL(options)
	fmt.Printf("SignURL: %s\n", signURL)
	var apiOptions = GetDefaultAPIOptioins()
	apiOptions.IsSecurity = true
	var api, _ = NewAPI(apiOptions)
	signURL = api.SignURL(options)
	fmt.Printf("SignURL: %s\n", signURL)

	apiOptions.Host = "example.com"
	api, _ = NewAPI(apiOptions)
	signURL = api.SignURL(options)
	fmt.Printf("SignURL: %s\n", signURL)

	apiOptions.Host = "example.com"
	api, _ = NewAPI(apiOptions)
	options.Bucket = "bucket"
	api.SetIsOSSHost(true)
	signURL = api.SignURL(options)
	fmt.Printf("SignURL: %s\n", signURL)
}

func TestSignURLAuthWithExpireTime(t *testing.T) {
	var options = GetDefaultSignURLOptions()
	var signURL = api.SignURLAuthWithExpireTime(options)
	fmt.Printf("SignURL: %s\n", signURL)
}

func TestObjectAPI(t *testing.T) {
	var bucket = "bucket"
	var object = "object"
	var failBucket = "403"
	var body = bytes.NewBufferString("this is the body")
	var contentType = "plan/text"
	var headers = make(map[string]string)
	headers["Content-Type"] = contentType
	var err error
	if err = api.PutObject(bucket, object, bufio.NewReader(body), headers); err != nil {
		t.Fatal(err)
	}
	fp, err := os.Open("oss.go")
	if err != nil {
		t.Fatal(err)
	}
	if err = api.PostObject(bucket, object, fp, headers); err != nil {
		t.Fatal(err)
	}

	var data io.ReadCloser
	if data, err = api.GetObject(failBucket, object, nil, nil); err == nil {
		t.Fatal("need fail, but success")
	}
	if data, err = api.GetObject(bucket, object, nil, nil); err != nil {
		t.Fatal(err)
	}
	var buf, _ = ioutil.ReadAll(data)
	fmt.Printf("%s\n", buf)

	var acl AccessControlPolicy
	if err = api.GetObjectACL(bucket, object, &acl); err != nil {
		t.Fatal(err)
	}

	if err = api.PutObjectACL(bucket, object, "public-read"); err != nil {
		t.Fatal(err)
	}

	if _, err = api.HeadObject(bucket, object, nil); err != nil {
		t.Fatal(err)
	}

	if _, err = api.HeadObject(failBucket, object, nil); err == nil {
		t.Fatal("need fail, but success")
	}

	if err = api.DeleteObject(bucket, object); err != nil {
		t.Fatal(err)
	}

	var deleteResult DeleteResult

	if err = api.DeleteObjects(bucket, []string{"object1", "object2"}, &deleteResult); err != nil {
		t.Fatal(err)
	}

	if err = api.DeleteObjects(bucket, []string{"object1", "object2"}, nil); err != nil {
		t.Fatal(err)
	}

	body = bytes.NewBufferString("this is the body")
	if _, err = api.AppendObject(failBucket, object, 0, bufio.NewReader(body), headers); err == nil {
		t.Fatal("need fail, but success")
	}
	body = bytes.NewBufferString("this is the body")
	if _, err = api.AppendObject(bucket, object, 0, bufio.NewReader(body), headers); err != nil {
		t.Fatal(err)
	}
	fp, err = os.Open("oss.go")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = api.AppendObject(bucket, object, 0, fp, headers); err != nil {
		t.Fatal(err)
	}
	if _, err = api.CopyObject(bucket, object, "bucket", "object", headers); err != nil {
		t.Fatal(err)
	}
}

func TestMultipartUploadAPI(t *testing.T) {
	var bucket = "bucket"
	var failBucket = "403"
	var object = "object"
	var uploadID = "uploadID"
	var contentType = "plan/text"
	var headers = make(map[string]string)
	headers["Content-Type"] = contentType
	var err error
	var multi *MultipartUpload
	var badMulti *MultipartUpload

	badMulti, err = api.GetMultiPartUpload(failBucket, object, uploadID)

	if multi, err = api.NewMultipartUpload(failBucket, object, headers); err == nil {
		t.Fatal("need fail, but success")
	}

	if multi, err = api.NewMultipartUpload(bucket, object, headers); err != nil {
		t.Fatal(err)
	}

	var body = bytes.NewBufferString("this is the body")
	var etag = "etag"
	if etag, err = badMulti.UploadPart(1, body); err == nil {
		t.Fatal("need fail, but success")
	}
	body = bytes.NewBufferString("this is the body")
	etag = "etag"
	if etag, err = multi.UploadPart(1, body); err != nil {
		t.Fatal(err)
	}
	fp, err := os.Open("oss.go")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("ETag: %s\n", etag)
	if etag, err = multi.UploadPart(2, fp); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("ETag: %s\n", etag)

	if etag, err = badMulti.CopyPart("bucket1", "object1", 4, "bytes=0-10", headers); err == nil {
		t.Fatal("need fail, but success")
	}

	if etag, err = multi.CopyPart("bucket1", "object1", 4, "bytes=0-10", headers); err != nil {
		t.Fatal(err)
	}
	fmt.Printf("ETag: %s\n", etag)

	if err = multi.AbortUpload(); err != nil {
		t.Fatal(err)
	}
	var result CompleteMultipartUploadResult
	var parts = []Part{
		Part{
			PartNumber: 1,
			ETag:       "ETag",
		},
		Part{
			PartNumber: 2,
			ETag:       "ETag",
		},
	}
	if err = multi.CompleteUpload(parts, &result); err != nil {
		t.Fatal(err)
	}

	var result2 ListPartsResult
	if err = multi.ListParts(1000, 3, &result2); err != nil {
		t.Fatal(err)
	}

	var opts = GetDefaultListMultipartUploadOptions()
	if _, err = api.ListMultipartUpload(bucket, opts); err != nil {
		t.Fatal(err)
	}

	if _, err = api.ListMultipartUpload(failBucket, opts); err == nil {
		t.Fatal("need fail, but success")
	}

}

func TestBucketCORSAPI(t *testing.T) {
	var bucket = "bucket"
	var object = "object"
	var failBucket = "403"
	var rules = []CORSRule{
		CORSRule{
			AllowedOrigin: []string{"*"},
			AllowedMethod: []string{"GET", "POST"},
			AllowedHeader: []string{"Authorization"},
			ExposeHeader:  []string{"x-oss-test"},
			MaxAgeSeconds: 100,
		},
		CORSRule{
			AllowedOrigin: []string{"*"},
			AllowedMethod: []string{"GET", "POST"},
			AllowedHeader: []string{"Authorization"},
			ExposeHeader:  []string{"x-oss-test"},
			MaxAgeSeconds: 100,
		},
	}
	var config = CORSConfiguration{
		Rules: rules,
	}

	var err error
	if err = api.PutBucketCORS(bucket, config); err != nil {
		t.Fatal(err)
	}

	if err = api.GetBucketCORS(bucket, &config); err != nil {
		t.Fatal(err)
	}

	if err = api.DeleteBucketCORS(bucket); err != nil {
		t.Fatal(err)
	}

	if _, err = api.OptionObject(bucket, object, nil); err != nil {
		t.Fatal(err)
	}
	if _, err = api.OptionObject(failBucket, object, nil); err == nil {
		t.Fatal("need fail, but success")
	}
}

func TestUploadLargeFile(t *testing.T) {
	res, _ := http.Get("https://huabot.b0.upaiyun.com/tweet/2681545d58a63be82851ad59b49461ecaa9d1555")
	defer res.Body.Close()
	var fileName = "/tmp/oss-go-sdk-test.png"
	var fp, _ = os.Create(fileName)
	io.Copy(fp, res.Body)
	fp.Close()
	api.SetDebug()
	api.UploadLargeFile("bucket", "object", fileName, 1024*101, nil)
}
