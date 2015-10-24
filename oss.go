package oss

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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

// SetIsOSSHost set is oss host for OSS API
func (api *API) SetIsOSSHost(isOSSHost bool) {
	api.isOSSDomain = isOSSHost
}

// SignURLOptions defined sign url options
type SignURLOptions struct {
	Method   string            // one of PUT, GET, DELETE, HEAD
	URL      string            // HTTP address of bucket or object, eg: http://HOST/bucket/object
	Headers  map[string]string // HTTP header
	Resource string            // path of bucket or bbject, eg: /bucket/ or /bucket/object
	Timeout  time.Duration
	Params   map[string]string
	Object   string // only for SignURL
	Bucket   string // only for SignURL
}

// GetDefaultSignURLOptions defined default sign url options
func GetDefaultSignURLOptions() *SignURLOptions {
	var options = new(SignURLOptions)
	options.Method = "GET"
	options.Headers = make(map[string]string)
	options.Params = make(map[string]string)
	options.Resource = "/"
	options.Timeout = 60 * time.Second
	return options
}

// SignURLAuthWithExpireTime Create the authorization for OSS based on the input method, url, body and headers
//
// Returns:
//     signature url.
func (api *API) SignURLAuthWithExpireTime(options *SignURLOptions) string {
	var sendTime = time.Now().Add(options.Timeout).UTC().Format("Mon, 02 Oct 2006 03:04:05 GMT")
	options.Headers["Date"] = sendTime
	var authValue = getAssign(api.secretAccessKey, options.Method, options.Headers,
		options.Resource, nil, api.debug)
	options.Params["OSSAccessKeyId"] = api.accessID
	options.Params["Expires"] = sendTime
	options.Params["Signature"] = authValue
	var signURL = appendParam(options.URL, options.Params)
	return signURL
}

// SignURL Create the authorization for OSS based on the input method, url, body and headers
//
// Returns:
//     signature url.
func (api *API) SignURL(options *SignURLOptions) string {
	var sendTime = time.Now().Add(options.Timeout).UTC().Format("Mon, 02 Oct 2006 03:04:05 GMT")
	options.Headers["Date"] = sendTime
	var resource = fmt.Sprintf("/%s/%s%s", options.Bucket, options.Object, getResource(options.Params))
	var authValue = getAssign(api.secretAccessKey, options.Method, options.Headers, resource, nil, api.debug)
	options.Params["OSSAccessKeyId"] = api.accessID
	options.Params["Expires"] = sendTime
	options.Params["Signature"] = authValue
	var url = ""
	options.Object = quote(options.Object)
	var schema = "http"
	if api.isSecurity {
		schema = "https"
	}
	if isIP(api.host) {
		url = fmt.Sprintf("%s://%s/%s/%s", schema, api.host, options.Bucket, options.Object)
	} else if isOSSHost(api.host, api.isOSSDomain) {
		if checkBucketValid(options.Bucket) {
			url = fmt.Sprintf("%s://%s.%s/%s", schema, options.Bucket, api.host, options.Object)
		} else {
			url = fmt.Sprintf("%s://%s/%s/%s", schema, api.host, options.Bucket, options.Object)
		}
	} else {
		url = fmt.Sprintf("%s://%s/%s", schema, api.host, options.Object)
	}
	var signURL = appendParam(url, options.Params)
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
func (api *API) httpRequest(options *RequestOptions) (res *http.Response, err error) {

	var req *http.Request
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
		options.Headers["Date"] = time.Now().UTC().Format("Mon, 02 Oct 2006 03:04:05 GMT")
		options.Headers["Authorization"] = api.createSignForNormalAuth(options.Method, options.Headers, resource)
		options.Headers["User-Agent"] = api.agent
		if checkBucketValid(options.Bucket) && !isIP(api.host) {
			host = options.Headers["Host"]
		} else {
			host = api.host
		}

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
		if res.StatusCode/100 != 2 {
			var errStr, _ = ioutil.ReadAll(res.Body)
			err = errors.New(string(errStr))
		}
		break
	}
	return
}

// httpRequestWithUnmarshalXML http request and xml unmarshal
func (api *API) httpRequestWithUnmarshalXML(options *RequestOptions, result interface{}) error {
	var data []byte
	var err error
	var res *http.Response

	if res, err = api.httpRequest(options); err != nil {
		return err
	}

	if result != nil {
		if data, err = ioutil.ReadAll(res.Body); err != nil {
			return err
		}

		if err = xml.Unmarshal(data, result); err != nil {
			return err
		}
	}
	return nil
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
	return api.httpRequestWithUnmarshalXML(options, result)
}

// GetBucketACL Get Access Control Level of bucket
func (api *API) GetBucketACL(bucket string, result *AccessControlPolicy) error {
	var options = GetDefaultRequestOptions()
	options.Bucket = bucket
	options.Params["acl"] = ""
	return api.httpRequestWithUnmarshalXML(options, result)
}

// GetBucketLocation Get Location of bucket
func (api *API) GetBucketLocation(bucket string, result *LocationConstraint) error {
	var options = GetDefaultRequestOptions()
	options.Bucket = bucket
	options.Params["location"] = ""
	return api.httpRequestWithUnmarshalXML(options, result)
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
	return api.httpRequestWithUnmarshalXML(options, result)
}

// GetBucketWebsite Get bucket website
func (api *API) GetBucketWebsite(bucket string, result *WebsiteConfiguration) error {
	var options = GetDefaultRequestOptions()
	options.Bucket = bucket
	options.Params["website"] = ""
	return api.httpRequestWithUnmarshalXML(options, result)
}

// GetBucketReferer Get bucket referer list
func (api *API) GetBucketReferer(bucket string, result *RefererConfiguration) error {
	var options = GetDefaultRequestOptions()
	options.Bucket = bucket
	options.Params["referer"] = ""
	return api.httpRequestWithUnmarshalXML(options, result)
}

// GetBucketLifecycle Get bucket lifecycle
func (api *API) GetBucketLifecycle(bucket string, result *LifecycleConfiguration) error {
	var options = GetDefaultRequestOptions()
	options.Bucket = bucket
	options.Params["lifecycle"] = ""
	return api.httpRequestWithUnmarshalXML(options, result)
}

// GetBucketLogging Get bucket logging
func (api *API) GetBucketLogging(bucket string, result *BucketLoggingStatus) error {
	var options = GetDefaultRequestOptions()
	options.Bucket = bucket
	options.Params["logging"] = ""
	return api.httpRequestWithUnmarshalXML(options, result)
}

// CreateBucket defined create bucket
func (api *API) CreateBucket(bucket string, acl ACLGrant, headers map[string]string) error {
	return api.PutBucket(bucket, acl, nil, headers)
}

// PutBucket create bucket
//
// :type bucket: string
// :param
//
// :type acl: ACLGrant
// :param: one of private public-read public-read-write
//
// :type headers: map[string]string
// :param: HTTP header
func (api *API) PutBucket(bucket string, acl ACLGrant, config *CreateBucketConfiguration, headers map[string]string) error {
	var options = GetDefaultRequestOptions()
	options.Method = "PUT"
	options.Bucket = bucket
	if headers != nil {
		options.Headers = headers
	}
	if acl != "" {
		options.Headers["x-oss-acl"] = string(acl)
	}
	if config != nil {
		var data, _ = xml.Marshal(config)
		options.Body = bytes.NewBuffer(data)
	}
	return api.httpRequestWithUnmarshalXML(options, nil)
}

// PutBucketACL create bucket with acl or update bucket acl when bucket is exists
func (api *API) PutBucketACL(bucket string, acl ACLGrant, headers map[string]string) error {
	var options = GetDefaultRequestOptions()
	options.Method = "PUT"
	options.Bucket = bucket
	if headers != nil {
		options.Headers = headers
	}
	options.Params["acl"] = ""
	options.Headers["x-oss-acl"] = string(acl)
	return api.httpRequestWithUnmarshalXML(options, nil)
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
	options.Params["logging"] = ""
	return api.httpRequestWithUnmarshalXML(options, nil)
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
	options.Params["website"] = ""
	return api.httpRequestWithUnmarshalXML(options, nil)
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
	options.Params["lifecycle"] = ""
	return api.httpRequestWithUnmarshalXML(options, nil)
}

// PutBucketReferer put bucket referer
func (api *API) PutBucketReferer(bucket string, config RefererConfiguration) error {
	var options = GetDefaultRequestOptions()
	options.Method = "PUT"
	options.Bucket = bucket
	var data, _ = xml.Marshal(config)
	options.Body = bytes.NewBuffer(data)
	options.Params["referer"] = ""
	var base64md5 = getBase64MD5(data)
	options.Headers["Content-MD5"] = base64md5
	return api.httpRequestWithUnmarshalXML(options, nil)
}

// DeleteBucket List object that in bucket
func (api *API) DeleteBucket(bucket string) error {
	var options = GetDefaultRequestOptions()
	options.Method = "DELETE"
	options.Bucket = bucket
	return api.httpRequestWithUnmarshalXML(options, nil)
}

// DeleteBucketWebsite Delete bucket website
func (api *API) DeleteBucketWebsite(bucket string) error {
	var options = GetDefaultRequestOptions()
	options.Method = "DELETE"
	options.Bucket = bucket
	options.Params["website"] = ""
	return api.httpRequestWithUnmarshalXML(options, nil)
}

// DeleteBucketLifecycle Delete bucket lifecycle
func (api *API) DeleteBucketLifecycle(bucket string) error {
	var options = GetDefaultRequestOptions()
	options.Method = "DELETE"
	options.Bucket = bucket
	options.Params["lifecycle"] = ""
	return api.httpRequestWithUnmarshalXML(options, nil)
}

// DeleteBucketLogging Delete bucket logging
func (api *API) DeleteBucketLogging(bucket string) error {
	var options = GetDefaultRequestOptions()
	options.Method = "DELETE"
	options.Bucket = bucket
	options.Params["logging"] = ""
	return api.httpRequestWithUnmarshalXML(options, nil)
}

// GetObject defined get object
func (api *API) GetObject(bucket, object string, headers, params map[string]string) (io.Reader, error) {
	var options = GetDefaultRequestOptions()
	options.Bucket = bucket
	options.Object = object
	options.Headers = headers
	options.Params = params

	var res *http.Response
	var err error
	if res, err = api.httpRequest(options); err != nil {
		return nil, err
	}

	var reader = bufio.NewReader(res.Body)
	return reader, nil
}

// GetObjectACL defined get object acl
func (api *API) GetObjectACL(bucket, object string, result *AccessControlPolicy) error {
	var options = GetDefaultRequestOptions()
	options.Bucket = bucket
	options.Object = object
	options.Params["acl"] = ""
	return api.httpRequestWithUnmarshalXML(options, result)
}

// HeadObject defined head object
func (api *API) HeadObject(bucket, object string, headers map[string]string) (result http.Header, err error) {
	var options = GetDefaultRequestOptions()
	options.Method = "HEAD"
	options.Bucket = bucket
	options.Object = object
	var res *http.Response
	if res, err = api.httpRequest(options); err != nil {
		return
	}
	return res.Header, nil
}

// PutObject defined put object
func (api *API) PutObject(bucket, object string, body io.Reader, headers map[string]string) error {
	var options = GetDefaultRequestOptions()
	options.Method = "PUT"
	options.Bucket = bucket
	options.Object = object
	if headers != nil {
		options.Headers = headers
	}

	if bodySeeker, ok := body.(io.ReadSeeker); ok {
		options.Headers["Content-MD5"] = getBase64MD5WithReader(bodySeeker)
		bodySeeker.Seek(0, 0)
		options.Body = bodySeeker
	} else {
		var data, _ = ioutil.ReadAll(body)
		options.Headers["Content-MD5"] = getBase64MD5(data)
		options.Body = bytes.NewBuffer(data)
	}

	var err error
	_, err = api.httpRequest(options)
	return err
}

// PostObject is same to PutObject, but use POST method, so just alisa to PutObject
func (api *API) PostObject(bucket, object string, body io.Reader, headers map[string]string) error {
	return api.PutObject(bucket, object, body, headers)
}

// PutObjectACL defined put object acl
func (api *API) PutObjectACL(bucket, object, acl string) error {
	var options = GetDefaultRequestOptions()
	options.Method = "PUT"
	options.Bucket = bucket
	options.Object = object
	options.Headers["x-oss-object-acl"] = acl
	options.Params["acl"] = ""
	_, err := api.httpRequest(options)
	return err
}

// CopyObject defined copy object
func (api *API) CopyObject(sourceBucket, sourceObject, targetBucket, targetObject string,
	headers map[string]string) (result CopyObjectResult, err error) {
	var options = GetDefaultRequestOptions()
	options.Method = "PUT"
	options.Bucket = targetBucket
	options.Object = targetObject
	if headers != nil {
		options.Headers = headers
	}

	options.Headers["x-oss-copy-source"] = fmt.Sprintf("/%s/%s", sourceBucket, quote(sourceObject))
	err = api.httpRequestWithUnmarshalXML(options, &result)
	return
}

// AppendObject defined append object
func (api *API) AppendObject(bucket, object string, position int, body io.Reader,
	headers map[string]string) (result http.Header, err error) {

	var options = GetDefaultRequestOptions()
	options.Method = "PUT"
	options.Bucket = bucket
	options.Object = object
	if headers != nil {
		options.Headers = headers
	}

	if bodySeeker, ok := body.(io.ReadSeeker); ok {
		options.Headers["Content-MD5"] = getBase64MD5WithReader(bodySeeker)
		bodySeeker.Seek(0, 0)
		options.Body = bodySeeker
	} else {
		var data, _ = ioutil.ReadAll(body)
		options.Headers["Content-MD5"] = getBase64MD5(data)
		options.Body = bytes.NewBuffer(data)
	}
	options.Params["append"] = ""
	options.Params["postion"] = strconv.Itoa(position)
	var res *http.Response
	if res, err = api.httpRequest(options); err != nil {
		return
	}

	result = res.Header

	return

}

// DeleteObject defined delete object
func (api *API) DeleteObject(bucket, object string) error {
	var options = GetDefaultRequestOptions()
	options.Method = "DELETE"
	options.Bucket = bucket
	options.Object = object
	return api.httpRequestWithUnmarshalXML(options, nil)
}

// DeleteObjects defined delete multiple object
func (api *API) DeleteObjects(bucket string, objects []string, result *DeleteResult) error {
	var options = GetDefaultRequestOptions()
	options.Method = "POST"
	options.Bucket = bucket

	var quiet = false
	if result == nil {
		quiet = true
	}

	var keys = make([]ObjectKey, len(objects))
	for idx, object := range objects {
		keys[idx] = ObjectKey{Key: object}
	}

	var deleteXML = DeleteXML{
		Quiet:   quiet,
		Objects: keys,
	}

	var data, _ = xml.Marshal(deleteXML)
	options.Headers["Content-MD5"] = getBase64MD5(data)
	options.Body = bytes.NewBuffer(data)
	options.Params["delete"] = ""
	if result != nil {
		return api.httpRequestWithUnmarshalXML(options, result)
	}
	return api.httpRequestWithUnmarshalXML(options, nil)
}

// MultipartUpload defined multipart upload struct
type MultipartUpload struct {
	api       *API
	Bucket    string
	Key       string
	UploadID  string
	Initiated time.Time
}

// NewMultipartUpload defined initial multipart upload
func (api *API) NewMultipartUpload(bucket, object string, headers map[string]string) (*MultipartUpload, error) {
	var options = GetDefaultRequestOptions()
	options.Method = "POST"
	options.Bucket = bucket
	options.Object = object
	options.Params["uploads"] = ""
	var result InitiateMultipartUploadResult
	if err := api.httpRequestWithUnmarshalXML(options, &result); err != nil {
		return nil, err
	}
	var multi = new(MultipartUpload)
	multi.Bucket = result.Bucket
	multi.Key = result.Key
	multi.UploadID = result.UploadID
	multi.api = api
	return multi, nil
}

// GetMultiPartUpload defined get multipart upload
func (api *API) GetMultiPartUpload(bucket, object, uploadID string) (*MultipartUpload, error) {
	var multi = new(MultipartUpload)
	multi.Bucket = bucket
	multi.Key = object
	multi.UploadID = uploadID
	multi.api = api
	return multi, nil
}

// UploadPart defined upload part
func (multi *MultipartUpload) UploadPart(partNumber int, body io.Reader) (string, error) {
	var options = GetDefaultRequestOptions()
	options.Method = "PUT"
	options.Bucket = multi.Bucket
	options.Object = multi.Key
	options.Params["partNumber"] = strconv.Itoa(partNumber)
	options.Params["uploadId"] = multi.UploadID

	if bodySeeker, ok := body.(io.ReadSeeker); ok {
		options.Headers["Content-MD5"] = getBase64MD5WithReader(bodySeeker)
		bodySeeker.Seek(0, 0)
		options.Body = bodySeeker
	} else {
		var data, _ = ioutil.ReadAll(body)
		options.Headers["Content-MD5"] = getBase64MD5(data)
		options.Body = bytes.NewBuffer(data)
	}

	var res *http.Response
	var err error
	if res, err = multi.api.httpRequest(options); err != nil {
		return "", err
	}
	return res.Header.Get("ETag"), nil
}

// CopyPart defined copy an exists object part
func (multi *MultipartUpload) CopyPart(sourceBucket, sourceObject string, partNumber int,
	sourceRange string, headers map[string]string) (string, error) {

	var options = GetDefaultRequestOptions()
	options.Method = "PUT"
	options.Bucket = multi.Bucket
	options.Object = multi.Key
	options.Params["partNumber"] = strconv.Itoa(partNumber)
	options.Params["uploadId"] = multi.UploadID
	if headers != nil {
		options.Headers = headers
	}
	options.Headers["x-oss-copy-source"] = fmt.Sprintf("/%s/%s", sourceBucket, quote(sourceObject))

	if len(sourceRange) > 0 {
		options.Headers["x-oss-copy-source-range"] = sourceRange
	}

	var result CopyPartResult
	var err error
	if err = multi.api.httpRequestWithUnmarshalXML(options, &result); err != nil {
		return "", err
	}
	return result.ETag, nil
}

// CompleteUpload defined multipart complete upload
func (multi *MultipartUpload) CompleteUpload(parts []Part, result *CompleteMultipartUploadResult) error {
	var options = GetDefaultRequestOptions()
	options.Method = "POST"
	options.Bucket = multi.Bucket
	options.Object = multi.Key
	options.Params["uploadId"] = multi.UploadID
	var partXML = CompleteMultipartUpload{
		Parts: parts,
	}
	var data, _ = xml.Marshal(partXML)
	options.Body = bytes.NewBuffer(data)
	return multi.api.httpRequestWithUnmarshalXML(options, &result)
}

// AbortUpload defined abort multipart upload
func (multi *MultipartUpload) AbortUpload() error {
	var options = GetDefaultRequestOptions()
	options.Method = "DELETE"
	options.Bucket = multi.Bucket
	options.Object = multi.Key
	options.Params["uploadId"] = multi.UploadID
	_, err := multi.api.httpRequest(options)
	return err
}

// ListMultipartUploadOptions defined list multipart upload options
type ListMultipartUploadOptions struct {
	Params             map[string]string
	KeyMarker          string
	UploadIDMarker     string
	NextKeyMarker      string
	NextUploadIDMarker string
	Delimiter          string
	Prefix             string
	MaxUploads         string
	EncodingType       string
}

// GetDefaultListMultipartUploadOptions defined get default list multipart upload options
func GetDefaultListMultipartUploadOptions() *ListMultipartUploadOptions {
	var options = new(ListMultipartUploadOptions)
	options.Params = make(map[string]string)
	return options
}

// ListMultipartUpload defined list multipart upload
func (api *API) ListMultipartUpload(bucket string, opts *ListMultipartUploadOptions) ([]*MultipartUpload, error) {
	var options = GetDefaultRequestOptions()
	options.Method = "GET"
	options.Bucket = bucket
	options.Params = opts.Params
	options.Params["uploads"] = ""
	options.Params["delimiter"] = opts.Delimiter
	options.Params["max-uploads"] = opts.MaxUploads
	options.Params["key-marker"] = opts.KeyMarker
	options.Params["prefix"] = opts.Prefix
	options.Params["upload-id-marker"] = opts.UploadIDMarker
	options.Params["encoding-type"] = opts.EncodingType
	var result ListMultipartUploadsResult
	if err := api.httpRequestWithUnmarshalXML(options, &result); err != nil {
		return nil, err
	}
	opts.KeyMarker = result.KeyMarker
	opts.UploadIDMarker = result.UploadIDMarker
	opts.NextKeyMarker = result.NextKeyMarker
	opts.NextUploadIDMarker = result.NextUploadIDMarker
	opts.Delimiter = result.Delimiter
	opts.Prefix = result.Prefix
	opts.MaxUploads = result.MaxUploads
	var uploads = make([]*MultipartUpload, len(result.Uploads))

	for id, v := range result.Uploads {
		uploads[id] = &MultipartUpload{
			api:       api,
			Bucket:    result.Bucket,
			Key:       v.Key,
			UploadID:  v.UploadID,
			Initiated: v.Initiated,
		}
	}

	return uploads, nil
}

// ListParts defined list parts
func (multi *MultipartUpload) ListParts(maxParts, partNumberMarker int,
	result *ListPartsResult) error {
	var options = GetDefaultRequestOptions()
	options.Method = "GET"
	options.Bucket = multi.Bucket
	options.Object = multi.Key
	options.Params["uploadId"] = multi.UploadID
	return multi.api.httpRequestWithUnmarshalXML(options, &result)
}

// PutBucketCORS defined put bucket cors
func (api *API) PutBucketCORS(bucket string, config CORSConfiguration) error {
	var options = GetDefaultRequestOptions()
	options.Method = "PUT"
	options.Bucket = bucket
	var data, _ = xml.Marshal(config)
	options.Body = bytes.NewBuffer(data)
	options.Params["cors"] = ""
	return api.httpRequestWithUnmarshalXML(options, nil)
}

// GetBucketCORS Get bucket cors
func (api *API) GetBucketCORS(bucket string, result *CORSConfiguration) error {
	var options = GetDefaultRequestOptions()
	options.Bucket = bucket
	options.Params["cors"] = ""
	return api.httpRequestWithUnmarshalXML(options, result)
}

// DeleteBucketCORS Delete bucket cors
func (api *API) DeleteBucketCORS(bucket string) error {
	var options = GetDefaultRequestOptions()
	options.Method = "DELETE"
	options.Bucket = bucket
	options.Params["cors"] = ""
	return api.httpRequestWithUnmarshalXML(options, nil)
}

// OptionObject defined option object
func (api *API) OptionObject(bucket, object string, headers map[string]string) (http.Header, error) {
	var options = GetDefaultRequestOptions()
	options.Method = "OPTIONS"
	options.Bucket = bucket
	options.Object = object
	options.Headers = headers
	var err error
	var res *http.Response
	if res, err = api.httpRequest(options); err != nil {
		return nil, err
	}
	return res.Header, nil
}
