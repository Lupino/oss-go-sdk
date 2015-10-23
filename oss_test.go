package oss

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var api *API
var options *APIOptions

func init() {
	var ts = mockHTTPServer()
	options = GetDefaultAPIOptioins()
	options.Host = getHostFromURL(ts.URL)
	api = NewAPI(options)
}

func TestPrint(t *testing.T) {
	fmt.Println(AGENT)
}

func TestSetValue(t *testing.T) {
	var options = GetDefaultAPIOptioins()
	var api = NewAPI(options)
	api.SetTimeout(5)
	api.SetDebug()
	api.SetRetryTimes(5)
	api.SetSendBufferSize(1024)
	api.SetRecvBufferSize(1024 * 1024)
	api.SetIsOSSHost(false)
}

func getHostFromURL(uri string) string {
	var u, _ = url.Parse(uri)
	return u.Host
}

func TestHttpRequest(t *testing.T) {
	var err error
	var reqOptions = new(RequestOptions)
	reqOptions.Method = "GET"
	tls := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer tls.Close()

	var options = GetDefaultAPIOptioins()
	options.Host = getHostFromURL(tls.URL)
	options.IsSecurity = true
	options.StsToken = "sts_token"
	var api = NewAPI(options)

	err = api.httpRequest(reqOptions, nil)
	fmt.Printf("%v\n", err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	options = GetDefaultAPIOptioins()
	options.Host = getHostFromURL(ts.URL)
	api = NewAPI(options)
	err = api.httpRequest(reqOptions, nil)
	fmt.Printf("%v\n", err)
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
