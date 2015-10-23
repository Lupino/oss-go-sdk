package oss

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
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

// BucketOperation defined bucket operation
func (api *API) BucketOperation(method, bucket string, headers map[string]string,
	params map[string]string) (*http.Response, error) {
	return api.httpRequest(method, bucket, "", headers, nil, params)
}

// ObjectOperation defined object operation
func (api *API) ObjectOperation(method, bucket, object string, headers map[string]string,
	body io.Reader, params map[string]string) (*http.Response, error) {
	return api.httpRequest(method, bucket, object, headers, body, params)
}

// RequestOptions defined requset options
type RequestOptions struct {
	Method  string // one of PUT, GET, DELETE, HEAD, POST
	Bucket  string
	Object  string
	Headers map[string]string // HTTP header
	Body    io.Reader
	Params  map[string]string
}

// GetDefaultRequestOptions get default requrest options
func GetDefaultRequestOptions() *RequestOptions {
	var options = new(RequestOptions)
	options.Method = "GET"
	options.Headers = make(map[string]string)
	options.Params = make(map[string]string)
	return options
}

// httpRequest Send http request of operation
//
// :type options: RequestOptions
// :param
//
// :type result: interface{}
// :param
func (api *API) httpRequest(options *RequestOptions, result interface{}) (err error) {

	var req *http.Request
	var res *http.Response
	var host string

	if options.Headers == nil {
		options.Headers = make(map[string]string)
	}
	if options.Params == nil {
		options.Params = make(map[string]string)
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
			options.Headers["x-oss-security-token"] = api.stsToken
		}

		if len(options.Bucket) == 0 {
			resource = "/"
			options.Headers["Host"] = api.host
		} else {
			options.Headers["Host"] = fmt.Sprintf("%s.%s", options.Bucket, api.host)
			if !isOSSHost(api.host, api.isOSSDomain) {
				options.Headers["Host"] = api.host
			}
			resource = fmt.Sprintf("/%s/", options.Bucket)
		}

		resource = fmt.Sprintf("%s%s%s", resource, options.Object, getResource(options.Params))
		options.Object = quote(options.Object)
		var url = fmt.Sprintf("/%s", options.Object)
		if isIP(api.host) {
			url = fmt.Sprintf("/%s/%s", options.Bucket, options.Object)
			if len(options.Bucket) == 0 {
				url = fmt.Sprintf("/%s", options.Object)
			}
			options.Headers["Host"] = api.host
		}

		url = appendParam(url, options.Params)
		options.Headers["Date"] = time.Now().Format("Wed, 21 Oct 2015 07:17:58 GMT")
		options.Headers["Authorization"] = api.createSignForNormalAuth(options.Method, options.Headers, resource)
		options.Headers["User-Agent"] = api.agent
		if checkBucketValid(options.Bucket) && !isIP(api.host) {
			host = options.Headers["Host"]
		} else {
			host = api.host
		}

		fmt.Printf("%s %s%s%s %s\n", options.Method, schema, host, url, options.Headers["Host"])
		if req, err = http.NewRequest(options.Method, schema+host+url, options.Body); err != nil {
			continue
		}

		for k, v := range options.Headers {
			req.Header.Add(k, v)
		}

		var client = &http.Client{}

		if res, err = client.Do(req); err != nil {
			continue
		}
		if res.Request.Host != api.host {
			api.host = res.Request.Host
		}
		if result != nil {
			var data []byte
			if data, err = ioutil.ReadAll(res.Body); err != nil {
				continue
			}

			if err = xml.Unmarshal(data, result); err != nil {
				continue
			}
		}
		break
	}
	return
}

// GetService List all buckets of user
func (api *API) GetService(result *ListAllMyBucketsResult, headers map[string]string) error {
	return api.ListAllMyBuckets(result, headers)
}

// ListAllMyBuckets List all buckets of user
func (api *API) ListAllMyBuckets(result *ListAllMyBucketsResult, headers map[string]string) error {
	var options = GetDefaultRequestOptions()
	if result.Prefix != "" {
		options.Params["prefix"] = result.Prefix
	}
	if result.Marker != "" {
		options.Params["marker"] = result.Marker
	}
	if result.MaxKeys != "" {
		options.Params["max-keys"] = result.MaxKeys
	}
	options.Headers = headers
	return api.httpRequest(options, result)
}

// GetBucketACL Get Access Control Level of bucket
func (api *API) GetBucketACL(bucket string, result *AccessControlPolicy) error {
	var options = GetDefaultRequestOptions()
	options.Bucket = bucket
	options.Params["acl"] = "acl"
	return api.httpRequest(options, result)
}

// GetBucketLocation Get Location of bucket
func (api *API) GetBucketLocation(bucket string, result *LocationConstraint) error {
	var options = GetDefaultRequestOptions()
	options.Bucket = bucket
	options.Params["location"] = "location"
	return api.httpRequest(options, result)
}

// GetBucket List object that in bucket
func (api *API) GetBucket(bucket string, result *ListBucketResult, headers map[string]string) error {
	return api.ListBucket(bucket, result, headers)
}

// ListBucket List object that in bucket
func (api *API) ListBucket(bucket string, result *ListBucketResult, headers map[string]string) error {
	var options = GetDefaultRequestOptions()
	options.Bucket = bucket
	options.Method = "GET"
	options.Params["prefix"] = result.Prefix
	options.Params["marker"] = result.Marker
	options.Params["delimiter"] = result.Delimiter
	options.Params["max-keys"] = result.MaxKeys
	options.Params["encoding-type"] = result.EncodingType
	options.Headers = headers
	return api.httpRequest(options, result)
}

// GetBucketWebsite Get bucket website
func (api *API) GetBucketWebsite(bucket string, result *WebsiteConfiguration) error {
	var options = GetDefaultRequestOptions()
	options.Bucket = bucket
	options.Params["website"] = "website"
	return api.httpRequest(options, result)
}

// GetBucketReferer Get bucket referer list
func (api *API) GetBucketReferer(bucket string, result *RefererConfiguration) error {
	var options = GetDefaultRequestOptions()
	options.Bucket = bucket
	options.Params["referer"] = "referer"
	return api.httpRequest(options, result)
}

// GetBucketLifecycle Get bucket lifecycle
func (api *API) GetBucketLifecycle(bucket string, result *LifecycleConfiguration) error {
	var options = GetDefaultRequestOptions()
	options.Bucket = bucket
	options.Params["lifecycle"] = "lifecycle"
	return api.httpRequest(options, result)
}

// GetBucketLogging Get bucket logging
func (api *API) GetBucketLogging(bucket string, result *BucketLoggingStatus) error {
	var options = GetDefaultRequestOptions()
	options.Bucket = bucket
	options.Params["logging"] = "logging"
	return api.httpRequest(options, result)
}

// CreateBucket defined create bucket
func (api *API) CreateBucket(bucket, acl string, headers map[string]string) error {
	return api.PutBucket(bucket, acl, nil, headers)
}

// PutBucket create bucket
//
// :type bucket: string
// :param
//
// :type acl: string
// :param: one of private public-read public-read-write
//
// :type headers: map[string]string
// :param: HTTP header
func (api *API) PutBucket(bucket, acl string, config *CreateBucketConfiguration, headers map[string]string) error {
	var options = GetDefaultRequestOptions()
	options.Method = "PUT"
	if headers != nil {
		options.Headers = headers
	}
	if acl != "" {
		options.Headers["x-oss-acl"] = acl
	}
	if config != nil {
		var data, _ = xml.Marshal(config)
		options.Body = bytes.NewBuffer(data)
	}
	return api.httpRequest(options, nil)
}

// PutBucketACL create bucket with acl or update bucket acl when bucket is exists
func (api *API) PutBucketACL(bucket, acl string, headers map[string]string) error {
	return api.PutBucket(bucket, acl, nil, headers)
}

// PutBucketLogging Put bucket logging
func (api *API) PutBucketLogging(sourcebucket, targetbucket, prefix string) error {
	var options = GetDefaultRequestOptions()
	options.Method = "PUT"
	options.Bucket = sourcebucket
	var status = BucketLoggingStatus{
		Bucket: targetbucket,
		Prefix: prefix,
	}
	var data, _ = xml.Marshal(status)
	options.Body = bytes.NewBuffer(data)
	options.Params["logging"] = "logging"
	return api.httpRequest(options, nil)
}

// PutBucketWebsite Put bucket website
//
// :type bucket: string
// :param
//
// :type indexfile: string
// :param: the object that contain index page
//
// :type errorfile: string
// :param: the object taht contain error page
func (api *API) PutBucketWebsite(bucket, indexfile, errorfile string) error {
	var options = GetDefaultRequestOptions()
	options.Method = "PUT"
	options.Bucket = bucket
	var config = WebsiteConfiguration{
		IndexSuffix: indexfile,
		ErrorKey:    errorfile,
	}
	var data, _ = xml.Marshal(config)
	options.Body = bytes.NewBuffer(data)
	options.Params["website"] = "website"
	return api.httpRequest(options, nil)
}

// PutBucketLifecycle put bucket lifecycle
func (api *API) PutBucketLifecycle(bucket string, rule LifecycleRule) error {
	var options = GetDefaultRequestOptions()
	options.Method = "PUT"
	options.Bucket = bucket
	var config = LifecycleConfiguration{
		Rule: rule,
	}
	var data, _ = xml.Marshal(config)
	options.Body = bytes.NewBuffer(data)
	options.Params["lifecycle"] = "lifecycle"
	return api.httpRequest(options, nil)
}

// PutBucketReferer put bucket referer
func (api *API) PutBucketReferer(bucket string, config RefererConfiguration) error {
	var options = GetDefaultRequestOptions()
	options.Method = "PUT"
	options.Bucket = bucket
	var data, _ = xml.Marshal(config)
	options.Body = bytes.NewBuffer(data)
	options.Params["referer"] = "referer"
	var base64md5 = getBase64MD5(data)
	options.Headers["Content-MD5"] = base64md5
	return api.httpRequest(options, nil)
}

// DeleteBucket List object that in bucket
func (api *API) DeleteBucket(bucket string) error {
	var options = GetDefaultRequestOptions()
	options.Method = "DELETE"
	options.Bucket = bucket
	return api.httpRequest(options, nil)
}

// DeleteBucketWebsite Delete bucket website
func (api *API) DeleteBucketWebsite(bucket string) error {
	var options = GetDefaultRequestOptions()
	options.Method = "DELETE"
	options.Bucket = bucket
	options.Params["website"] = "website"
	return api.httpRequest(options, nil)
}

// DeleteBucketLifecycle Delete bucket lifecycle
func (api *API) DeleteBucketLifecycle(bucket string) error {
	var options = GetDefaultRequestOptions()
	options.Method = "DELETE"
	options.Bucket = bucket
	options.Params["lifecycle"] = "lifecycle"
	return api.httpRequest(options, nil)
}

// DeleteLogging Delete bucket logging
func (api *API) DeleteLogging(bucket string) error {
	var options = GetDefaultRequestOptions()
	options.Method = "DELETE"
	options.Bucket = bucket
	options.Params["logging"] = "logging"
	return api.httpRequest(options, nil)
}
