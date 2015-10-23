package oss

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strconv"
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

// GetService List all buckets of user
func (api *API) GetService(headers map[string]string, prefix, marker, maxKeys string) (*http.Response, error) {
	return api.ListAllMyBuckets(headers, prefix, marker, maxKeys)
}

// ListAllMyBuckets List all buckets of user
// type headers: dict
// :param
//
// Returns:
//     HTTP Response
func (api *API) ListAllMyBuckets(headers map[string]string, prefix, marker, maxKeys string) (*http.Response, error) {
	var params = make(map[string]string)
	if prefix != "" {
		params["prefix"] = prefix
	}
	if marker != "" {
		params["marker"] = marker
	}
	if maxKeys != "" {
		params["max-keys"] = maxKeys
	}
	return api.httpRequest("GET", "", "", headers, nil, params)
}

// GetBucketACL Get Access Control Level of bucket
//
// :type bucket: string
// :param
//
// Returns:
//     HTTP Response
func (api *API) GetBucketACL(bucket string) (*http.Response, error) {
	return api.httpRequest("GET", bucket, "", nil, nil, map[string]string{"acl": ""})
}

// GetBucketLocation Get Location of bucket
func (api *API) GetBucketLocation(bucket string) (*http.Response, error) {
	return api.httpRequest("GET", bucket, "", nil, nil, map[string]string{"location": ""})
}

// GetBucket List object that in bucket
func (api *API) GetBucket(bucket, prefix, marker, delimiter, maxKeys string, headers map[string]string, encodingType string) (*http.Response, error) {
	return api.ListBucket(bucket, prefix, marker, delimiter, maxKeys, headers, encodingType)
}

// ListBucket List object that in bucket
//
// :type bucket: string
// :param
//
// :type prefix: string
// :param
//
// :type marker: string
// :param
//
// :type delimiter: string
// :param
//
// :type maxkeys: string
// :param
//
// :type headers: dict
// :param: HTTP header
//
// :type maxkeys: string
// :encoding_type
//
// Returns:
//     HTTP Response
func (api *API) ListBucket(bucket, prefix, marker, delimiter, maxkeys string, headers map[string]string, encodingType string) (*http.Response, error) {
	var params = make(map[string]string)
	params["prefix"] = prefix
	params["marker"] = marker
	params["delimiter"] = delimiter
	params["max-keys"] = maxkeys
	params["encoding-type"] = encodingType
	return api.httpRequest("GET", bucket, "", headers, nil, params)
}

// GetWebsite Get bucket website
//
// :type bucket: string
// :param
//
// Returns:
//     HTTP Response
func (api *API) GetWebsite(bucket string, headers map[string]string) (*http.Response, error) {
	return api.httpRequest("GET", bucket, "", headers, nil, map[string]string{"website": ""})
}

// GetLifecycle Get bucket lifecycle
//
// :type bucket: string
// :param
//
// Returns:
//     HTTP Response
func (api *API) GetLifecycle(bucket string, headers map[string]string) (*http.Response, error) {
	return api.httpRequest("GET", bucket, "", headers, nil, map[string]string{"lifecycle": ""})
}

// GetLogging Get bucket logging
//
// :type bucket: string
// :param
//
// Returns:
//     HTTP Response
func (api *API) GetLogging(bucket string, headers map[string]string) (*http.Response, error) {
	return api.httpRequest("GET", bucket, "", headers, nil, map[string]string{"logging": ""})
}

// GetCors Get bucket cors
//
// :type bucket: string
// :param
//
// Returns:
//     HTTP Response
func (api *API) GetCors(bucket string, headers map[string]string) (*http.Response, error) {
	return api.httpRequest("GET", bucket, "", headers, nil, map[string]string{"cors": ""})
}

// CreateBucket defined create bucket
func (api *API) CreateBucket(bucket, acl string, headers map[string]string) (*http.Response, error) {
	return api.PutBucket(bucket, acl, headers)
}

// PutBucket create bucket
//
// :type bucket: string
// :param
//
// :type acl: string
// :param: one of private public-read public-read-write
//
// :type headers: dict
// :param: HTTP header
//
// Returns:
//     HTTP Response
func (api *API) PutBucket(bucket, acl string, headers map[string]string) (*http.Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	if acl != "" {
		if "AWS" == api.provider {
			headers["x-amz-acl"] = acl
		} else {
			headers["x-oss-acl"] = acl
		}
	}
	return api.httpRequest("PUT", bucket, "", headers, nil, nil)
}

// PutLogging Put bucket logging
//
// :type sourcebucket: string
// :param
//
// :type targetbucket: string
// :param: Specifies the bucket where you want Aliyun OSS to store server access logs
//
// :type prefix: string
// :param: This element lets you specify a prefix for the objects that the log files will be stored
//
// Returns:
//     HTTP Response
func (api *API) PutLogging(sourcebucket, targetbucket, prefix string) (*http.Response, error) {
	var buffer = bytes.NewBuffer(nil)
	buffer.WriteString("<BucketLoggingStatus>")
	if len(targetbucket) > 0 {
		buffer.WriteString("<LoggingEnabled>")
		buffer.WriteString("<TargetBucket>" + targetbucket + "</TargetBucket>")
		if len(prefix) > 0 {
			buffer.WriteString("<TargetPrefix>" + prefix + "</TargetPrefix>")
		}
		buffer.WriteString("</LoggingEnabled>")
	}
	buffer.WriteString("</BucketLoggingStatus>")

	return api.httpRequest("PUT", sourcebucket, "", nil, buffer, map[string]string{"logging": ""})
}

// PutWebsite Put bucket website
//
// :type bucket: string
// :param
//
// :type indexfile: string
// :param: the object that contain index page
//
// :type errorfile: string
// :param: the object taht contain error page
//
// Returns:
//     HTTP Response
func (api *API) PutWebsite(bucket, indexfile, errorfile string) (*http.Response, error) {
	var buffer = bytes.NewBufferString(
		fmt.Sprintf("<WebsiteConfiguration><IndexDocument><Suffix>%s</Suffix></IndexDocument><ErrorDocument><Key>%s</Key></ErrorDocument></WebsiteConfiguration>", indexfile, errorfile))
	return api.httpRequest("PUT", bucket, "", nil, buffer, map[string]string{"website": ""})
}

// PutLifecycle put bucket lifecycle
// :type bucket: string
// :param
//
// :type lifecycle: string
// :param: lifecycle configuration
//
// Returns:
//     HTTP Response
func (api *API) PutLifecycle(bucket, lifecycle string) (*http.Response, error) {
	var buffer = bytes.NewBufferString(lifecycle)
	return api.httpRequest("PUT", bucket, "", nil, buffer, map[string]string{"lifecycle": ""})
}

// PutCors put bucket cors
//
// :type bucket: string
// :param
//
// :type cors_xml: string
// :param: the xml that contain cors rules
//
// Returns:
//     HTTP Response
func (api *API) PutCors(bucket, corsXML string, headers map[string]string) (*http.Response, error) {
	var buffer = bytes.NewBufferString(corsXML)
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Length"] = strconv.Itoa(len(corsXML))
	var base64md5 = getStringBase64MD5(corsXML)
	headers["Content-MD5"] = base64md5
	return api.httpRequest("PUT", bucket, "", headers, buffer, map[string]string{"cors": ""})
}

// PutBucketWithLocation create bucket with location
//
// :type bucket: string
// :param
//
// :type acl: string
// :param: one of private public-read public-read-write
//
// :type location: string
// :param:
//
// :type headers: dict
// :param: HTTP header
//
// Returns:
//     HTTP Response
func (api *API) PutBucketWithLocation(bucket, acl, location string, headers map[string]string) (*http.Response, error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	if acl != "" {
		if "AWS" == api.provider {
			headers["x-amz-acl"] = acl
		} else {
			headers["x-oss-acl"] = acl
		}
	}
	var buffer = bytes.NewBuffer(nil)
	if location != "" {
		buffer.WriteString("<CreateBucketConfiguration>")
		buffer.WriteString("<LocationConstraint>")
		buffer.WriteString(location)
		buffer.WriteString("</LocationConstraint>")
		buffer.WriteString("</CreateBucketConfiguration>")
	}
	return api.httpRequest("PUT", bucket, "", headers, buffer, nil)
}

// DeleteBucket List object that in bucket
//
// :type bucket: string
// :param
//
// :type headers: dict
// :param: HTTP header
//
// Returns:
//     HTTP Response
func (api *API) DeleteBucket(bucket string, headers map[string]string) (*http.Response, error) {
	return api.httpRequest("DELETE", bucket, "", headers, nil, nil)
}

// DeleteWebsite Delete bucket website
//
// :type bucket: string
// :param
//
// Returns:
//     HTTP Response
func (api *API) DeleteWebsite(bucket string, headers map[string]string) (*http.Response, error) {
	return api.httpRequest("DELETE", bucket, "", headers, nil, map[string]string{"website": ""})
}

// DeleteLifecycle Delete bucket lifecycle
//
// :type bucket: string
// :param
//
// Returns:
//     HTTP Response
func (api *API) DeleteLifecycle(bucket string, headers map[string]string) (*http.Response, error) {
	return api.httpRequest("DELETE", bucket, "", headers, nil, map[string]string{"lifecycle": ""})
}

// DeleteLogging Delete bucket logging
//
// :type bucket: string
// :param
//
// Returns:
//     HTTP Response
func (api *API) DeleteLogging(bucket string, headers map[string]string) (*http.Response, error) {
	return api.httpRequest("DELETE", bucket, "", headers, nil, map[string]string{"logging": ""})
}

// DeleteCors Delete bucket cors
//
// :type bucket: string
// :param
//
// Returns:
//     HTTP Response
func (api *API) DeleteCors(bucket string, headers map[string]string) (*http.Response, error) {
	return api.httpRequest("DELETE", bucket, "", headers, nil, map[string]string{"cors": ""})
}
