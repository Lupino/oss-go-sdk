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

func TestPrint(t *testing.T) {
	fmt.Println(AGENT)
}

func TestSetValue(t *testing.T) {
	options = GetDefaultAPIOptioins()
	api = NewAPI(options)
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
	res, err = api.httpRequest("GET", "bucket", "", nil, nil, nil)
	fmt.Printf("%v, %v\n", res, err)
}
