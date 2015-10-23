package oss

import (
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"
)

// AGENT defined a http requset agent
var AGENT string

// Version defined OSS package version
const Version = "0.0.1"

func init() {
	AGENT = fmt.Sprintf("aliyun-sdk-go/%s (%s/%s;%s)", Version, runtime.GOOS, runtime.GOARCH, runtime.Version())
}

// APIOptions the options of OSS API
type APIOptions struct {
	Host string
	// Port            int
	AccessID        string
	SecretAccessKey string
	IsSecurity      bool
	StsToken        string
}

// GetDefaultAPIOptioins get default api options for OSS API
func GetDefaultAPIOptioins() *APIOptions {
	return &APIOptions{
		Host: "oss.aliyuncs.com:80",
	}
}

// API A simple OSS API
type API struct {
	sendBufferSize  int
	recvBufferSize  int
	host            string
	port            int
	accessID        string
	secretAccessKey string
	showBar         bool
	isSecurity      bool
	retryTimes      int
	agent           string
	debug           bool
	timeout         int
	isOSSDomain     bool
	stsToken        string
	provider        string
}

// NewAPI initial simple OSS API
func NewAPI(options *APIOptions) *API {
	var api = new(API)
	api.sendBufferSize = 8192
	api.recvBufferSize = 1024 * 1024 * 10
	api.host = getHostFromList(options.Host)
	api.accessID = options.AccessID
	api.secretAccessKey = options.SecretAccessKey
	api.showBar = false
	api.isSecurity = options.IsSecurity
	api.retryTimes = 5
	api.agent = AGENT
	api.debug = false
	api.timeout = 60
	api.isOSSDomain = false
	api.stsToken = options.StsToken
	api.provider = PROVIDER
	return api
}

// SetTimeout set timeout for OSS API
func (api *API) SetTimeout(timeout int) {
	api.timeout = timeout
}

// SetDebug set debug for OSS API
func (api *API) SetDebug() {
	api.debug = true
}

// SetRetryTimes set retry times for OSS API
func (api *API) SetRetryTimes(retryTimes int) {
	api.retryTimes = retryTimes
}

// SetSendBufferSize set send buffer size for OSS API
func (api *API) SetSendBufferSize(bufSize int) {
	api.sendBufferSize = bufSize
}

// SetRecvBufferSize set recv buffer size for OSS API
func (api *API) SetRecvBufferSize(bufSize int) {
	api.recvBufferSize = bufSize
}

// SetIsOSSHost set is oss host for OSS API
func (api *API) SetIsOSSHost(isOSSHost bool) {
	api.isOSSDomain = isOSSHost
}

// SignURLAuthWithExpireTime Create the authorization for OSS based on the input method, url, body and headers
//
// :type method: string
// :param method: one of PUT, GET, DELETE, HEAD
//
// :type url: string
// :param:HTTP address of bucket or object, eg: http://HOST/bucket/object
//
// :type headers: dict
// :param: HTTP header
//
// :type resource: string
// :param:path of bucket or object, eg: /bucket/ or /bucket/object
//
// :type timeout: int
// :param
//
// Returns:
//     signature url.
func (api *API) SignURLAuthWithExpireTime(method, url string, headers map[string]string,
	resource string, timeout time.Duration, params map[string]string) string {
	if headers != nil {
		headers = make(map[string]string)
	}
	if params != nil {
		params = make(map[string]string)
	}
	if len(resource) == 0 {
		resource = "/"
	}
	if timeout < 0 {
		timeout = 60 * time.Second
	}
	var sendTime = time.Now().Add(timeout).Format("Wed, 21 Oct 2015 07:17:58 GMT")
	headers["Date"] = sendTime
	var authValue = getAssign(api.secretAccessKey, method, headers, resource, nil, api.debug)
	params["OSSAccessKeyId"] = api.accessID
	params["Expires"] = sendTime
	params["Signature"] = authValue
	var signURL = appendParam(url, params)
	return signURL
}

// SignURL Create the authorization for OSS based on the input method, url, body and headers
//
// :type method: string
// :param method: one of PUT, GET, DELETE, HEAD
//
// :type bucket: string
// :param:
//
// :type object: string
// :param:
//
// :type timeout: int
// :param
//
// :type headers: dict
// :param: HTTP header
//
// :type params: dict
// :param: the parameters that put in the url address as query string
//
// :type resource: string
// :param:path of bucket or object, eg: /bucket/ or /bucket/object
//
// Returns:
//     signature url.
func (api *API) SignURL(method, bucket, object string, timeout time.Duration,
	headers map[string]string, params map[string]string) string {
	if headers != nil {
		headers = make(map[string]string)
	}
	if params != nil {
		params = make(map[string]string)
	}
	if timeout < 0 {
		timeout = 60 * time.Second
	}
	var sendTime = time.Now().Add(timeout).Format("Wed, 21 Oct 2015 07:17:58 GMT")
	headers["Date"] = sendTime
	var resource = fmt.Sprintf("/%s/%s%s", bucket, object, getResource(params))
	var authValue = getAssign(api.secretAccessKey, method, headers, resource, nil, api.debug)
	params["OSSAccessKeyId"] = api.accessID
	params["Expires"] = sendTime
	params["Signature"] = authValue
	var url = ""
	object = quote(object)
	var schema = "http"
	if api.isSecurity {
		schema = "https"
	}
	if isIP(api.host) {
		url = fmt.Sprintf("%s://%s/%s/%s", schema, api.host, bucket, object)
	} else if isOSSHost(api.host, api.isOSSDomain) {
		if checkBucketValid(bucket) {
			url = fmt.Sprintf("%s://%s.%s/%s", schema, bucket, api.host, object)
		} else {
			url = fmt.Sprintf("%s://%s/%s/%s", schema, api.host, bucket, object)
		}
	} else {
		url = fmt.Sprintf("%s://%s/%s", schema, api.host, object)
	}
	var signURL = appendParam(url, params)
	return signURL
}

// createSignForNormalAuth NOT public API
// Create the authorization for OSS based on header input.
// it should be put into "Authorization" parameter of header.
//
// :type method: string
// :param:one of PUT, GET, DELETE, HEAD
//
// :type headers: dict
// :param: HTTP header
//
// :type resource: string
// :param:path of bucket or object, eg: /bucket/ or /bucket/object
//
// Returns:
//     signature string
func (api *API) createSignForNormalAuth(method string, headers map[string]string, resource string) string {
	var authValue = fmt.Sprintf("%s %s:%s", api.provider, api.accessID,
		getAssign(api.secretAccessKey, method, headers, resource, nil, api.debug))
	return authValue
}

// httpRequest Send http request of operation
//
// :type method: string
// :param method: one of PUT, GET, DELETE, HEAD, POST
//
// :type bucket: string
// :param
//
// :type object: string
// :param
//
// :type headers: dict
// :param: HTTP header
//
// :type body: string
// :param
//
// Returns:
//     HTTP Response
func (api *API) httpRequest(method, bucket, object string,
	headers map[string]string,
	body io.Reader,
	params map[string]string) (res *http.Response, err error) {

	var req *http.Request
	var host string

	if headers == nil {
		headers = make(map[string]string)
	}
	if params == nil {
		params = make(map[string]string)
	}
	for i := 0; i < api.retryTimes; i++ {
		var _, port = getHostPort(api.host)
		var schema = "http://"
		if api.isSecurity || port == 443 {
			api.isSecurity = true
			schema = "https://"
		}

		var resource string

		if len(api.stsToken) > 0 {
			headers["x-oss-security-token"] = api.stsToken
		}

		if len(bucket) == 0 {
			resource = "/"
			headers["Host"] = api.host
		} else {
			headers["Host"] = fmt.Sprintf("%s.%s", bucket, api.host)
			if !isOSSHost(api.host, api.isOSSDomain) {
				headers["Host"] = api.host
			}
			resource = fmt.Sprintf("/%s/", bucket)
		}

		resource = fmt.Sprintf("%s%s%s", resource, object, getResource(params))
		object = quote(object)
		var url = fmt.Sprintf("/%s", object)
		if isIP(api.host) {
			url = fmt.Sprintf("/%s/%s", bucket, object)
			if len(bucket) == 0 {
				url = fmt.Sprintf("/%s", object)
			}
			headers["Host"] = api.host
		}

		url = appendParam(url, params)
		// date = time.strftime("%a, %d %b %Y %H:%M:%S GMT", time.gmtime())
		headers["Date"] = time.Now().Format("Wed, 21 Oct 2015 07:17:58 GMT")
		headers["Authorization"] = api.createSignForNormalAuth(method, headers, resource)
		headers["User-Agent"] = api.agent
		if checkBucketValid(bucket) && !isIP(api.host) {
			host = headers["Host"]
		} else {
			host = api.host
		}

		if req, err = http.NewRequest(method, schema+host+url, body); err != nil {
			continue
		}

		for k, v := range headers {
			req.Header.Add(k, v)
		}

		var client = &http.Client{}

		if res, err = client.Do(req); err != nil {
			continue
		}
		if res.Request.Host != api.host {
			api.host = res.Request.Host
		}
		break
	}
	return
}
