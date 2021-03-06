// Package oss aliyun OSS(open storage service) golang client.
//
// ## Create Account
//
// Go to [OSS website](http://www.aliyun.com/product/oss/?lang=en), create a new account for new user.
//
// After account created, you can create the OSS instance and get the `accessKeyId` and `accessKeySecret`.
//
// ## Initial OSS API
//
//  import (
//      "github.com/Lupino/oss-go-sdk"
//  )
//  var APIOptions = oss.GetDefaultAPIOptioins()
//  APIOptions.AccessID = AccessKeyID
//  APIOptions.SecretAccessKey = AccessKeySecret
//  var OSSAPI, err = oss.NewAPI(APIOptions)
//
// ## Get Service
//
//  var result oss.ListAllMyBucketsResult
//  var headers = make(map[string]string)
//  var err error
//  err = OSSAPI.GetService(&result, headers)
//
// ## Parse the error
//
// `oss-go-sdk` implement error return by OSS into `error` interface by `oss.Error`,
// so you can get error return by OSS server from `oss.Error`, just like:
//
//  var realErr = err.(*oss.Error)
//
package oss

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
)

// AGENT http requset agent
var AGENT string

// Version OSS package version
const Version = "0.0.1"

func init() {
	AGENT = fmt.Sprintf("aliyun-sdk-go/%s (%s/%s;%s)", Version, runtime.GOOS, runtime.GOARCH, runtime.Version())
}

// APIOptions the options of OSS API
type APIOptions struct {
	// The OSS server host
	Host string
	Port int
	// access key you create on aliyun console website
	AccessID string
	// access secret you create
	SecretAccessKey string
	IsSecurity      bool
	StsToken        string
}

// GetDefaultAPIOptioins get default api options for OSS API
func GetDefaultAPIOptioins() *APIOptions {
	return &APIOptions{
		Host: "oss.aliyuncs.com",
		Port: 80,
	}
}

// API A simple OSS API
type API struct {
	host            string
	port            int
	accessID        string
	secretAccessKey string
	isSecurity      bool
	retryTimes      int
	agent           string
	debug           bool
	// instance level timeout for all operations, default is 60s
	timeout     time.Duration
	isOSSDomain bool
	stsToken    string
	provider    string
}

// NewAPI initial simple OSS API
func NewAPI(options *APIOptions) (*API, error) {
	var api = new(API)
	api.host = options.Host
	api.port = options.Port
	api.accessID = options.AccessID
	api.secretAccessKey = options.SecretAccessKey
	api.isSecurity = options.IsSecurity
	api.retryTimes = 5
	api.agent = AGENT
	api.debug = false
	api.timeout = 60 * time.Second
	api.isOSSDomain = false
	api.stsToken = options.StsToken
	api.provider = PROVIDER

	if checkValidHost(api.host, api.port, api.timeout) {
		return api, nil
	}

	return nil, fmt.Errorf("Server: %s:%d is not avaliable.", api.host, api.port)
}

// SetTimeout set timeout for OSS API
func (api *API) SetTimeout(timeout time.Duration) {
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
	// one of PUT, GET, DELETE, HEAD
	Method string
	// HTTP address of bucket or object, eg: http://HOST/bucket/object
	URL string
	// HTTP header
	Headers map[string]string
	// path of bucket or bbject, eg: /bucket/ or /bucket/object
	Resource string
	Timeout  time.Duration
	Params   map[string]string
	// only for SignURL
	Object string
	// only for SignURL
	Bucket string
}

// GetDefaultSignURLOptions get default sign url options
func GetDefaultSignURLOptions() *SignURLOptions {
	var options = new(SignURLOptions)
	options.Method = "GET"
	options.Headers = make(map[string]string)
	options.Params = make(map[string]string)
	options.Resource = "/"
	options.Timeout = 60 * time.Second
	return options
}

// SignURLAuthWithExpireTime create the authorization for OSS based on the input method, url, body and headers
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

// SignURL create the authorization for OSS based on the input method, url, body and headers
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
	var host = api.host
	if api.port != 80 && api.port != 443 {
		host = fmt.Sprintf("%s:%d", api.host, api.port)
	}
	if isIP(api.host) {
		url = fmt.Sprintf("%s://%s/%s/%s", schema, host, options.Bucket, options.Object)
	} else if isOSSHost(api.host, api.isOSSDomain) {
		if checkBucketValid(options.Bucket) {
			url = fmt.Sprintf("%s://%s.%s/%s", schema, options.Bucket, host, options.Object)
		} else {
			url = fmt.Sprintf("%s://%s/%s/%s", schema, host, options.Bucket, options.Object)
		}
	} else {
		url = fmt.Sprintf("%s://%s/%s", schema, host, options.Object)
	}
	var signURL = appendParam(url, options.Params)
	return signURL
}

// createSignForNormalAuth NOT public API
// Create the authorization for OSS based on header input.
// it should be put into "Authorization" parameter of header.
//
//      - method: one of PUT, GET, DELETE, HEAD
//      - headers: HTTP header
//      - resource: path of bucket or object, eg: /bucket/ or /bucket/object
//
// Returns:
//     signature string
func (api *API) createSignForNormalAuth(method string, headers map[string]string, resource string) string {
	var authValue = fmt.Sprintf("%s %s:%s", api.provider, api.accessID,
		getAssign(api.secretAccessKey, method, headers, resource, nil, api.debug))
	return authValue
}

// requestOptions defined requset options
type requestOptions struct {
	// one of PUT, GET, DELETE, HEAD, POST
	Method string
	Bucket string
	Object string
	// HTTP header
	Headers map[string]string
	Body    io.Reader
	Params  map[string]string
	// AutoClose the res.Body
	AutoClose bool
}

// getDefaultRequestOptions get default requrest options
func getDefaultRequestOptions() *requestOptions {
	var options = new(requestOptions)
	options.Method = "GET"
	options.Headers = make(map[string]string)
	options.Params = make(map[string]string)
	options.AutoClose = false
	return options
}

// httpRequest send http request of operation
func (api *API) httpRequest(options *requestOptions) (res *http.Response, err error) {

	var req *http.Request
	var host string

	if options.Headers == nil {
		options.Headers = make(map[string]string)
	}
	if options.Params == nil {
		options.Params = make(map[string]string)
	}
	for i := 0; i < api.retryTimes; i++ {
		var schema = "http://"
		if api.isSecurity || api.port == 443 {
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

		if api.port != 80 && api.port != 443 {
			options.Headers["Host"] = fmt.Sprintf("%s:%s", options.Headers["Host"], api.port)
			host = fmt.Sprintf("%s:%d", host, api.port)
		}

		if req, err = http.NewRequest(options.Method, schema+host+url, options.Body); err != nil {
			continue
		}

		for k, v := range options.Headers {
			req.Header.Add(k, v)
		}

		var client = &http.Client{
			Timeout: api.timeout,
		}

		if res, err = client.Do(req); err != nil {
			continue
		}
		if res.StatusCode/100 != 2 {
			var errStr, _ = ioutil.ReadAll(res.Body)
			res.Body.Close()
			err = parseError(errStr)
		} else if options.AutoClose {
			res.Body.Close()
		}
		break
	}
	return
}

// httpRequestWithUnmarshalXML get http request and xml unmarshal
func (api *API) httpRequestWithUnmarshalXML(options *requestOptions, result interface{}) error {
	var data []byte
	var err error
	var res *http.Response

	if res, err = api.httpRequest(options); err != nil {
		return err
	}
	defer res.Body.Close()

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

// GetService list all buckets of user
func (api *API) GetService(result *ListAllMyBucketsResult, headers map[string]string) error {
	return api.ListAllMyBuckets(result, headers)
}

// ListAllMyBuckets list all buckets of user
func (api *API) ListAllMyBuckets(result *ListAllMyBucketsResult, headers map[string]string) error {
	var options = getDefaultRequestOptions()
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

// GetBucketACL get the bucket ACL.
func (api *API) GetBucketACL(bucket string, result *AccessControlPolicy) error {
	var options = getDefaultRequestOptions()
	options.Bucket = bucket
	options.Params["acl"] = ""
	return api.httpRequestWithUnmarshalXML(options, result)
}

// GetBucketLocation get Location of bucket
func (api *API) GetBucketLocation(bucket string, result *LocationConstraint) error {
	var options = getDefaultRequestOptions()
	options.Bucket = bucket
	options.Params["location"] = ""
	return api.httpRequestWithUnmarshalXML(options, result)
}

// GetBucket list object that in bucket
func (api *API) GetBucket(bucket string, result *ListBucketResult, headers map[string]string) error {
	return api.ListBucket(bucket, result, headers)
}

// ListBucket list object that in bucket
func (api *API) ListBucket(bucket string, result *ListBucketResult, headers map[string]string) error {
	var options = getDefaultRequestOptions()
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

// GetBucketWebsite get bucket website config.
func (api *API) GetBucketWebsite(bucket string, result *WebsiteConfiguration) error {
	var options = getDefaultRequestOptions()
	options.Bucket = bucket
	options.Params["website"] = ""
	return api.httpRequestWithUnmarshalXML(options, result)
}

// GetBucketReferer get the bucket request Referer white list.
func (api *API) GetBucketReferer(bucket string, result *RefererConfiguration) error {
	var options = getDefaultRequestOptions()
	options.Bucket = bucket
	options.Params["referer"] = ""
	return api.httpRequestWithUnmarshalXML(options, result)
}

// GetBucketLifecycle get bucket lifecycle
func (api *API) GetBucketLifecycle(bucket string, result *LifecycleConfiguration) error {
	var options = getDefaultRequestOptions()
	options.Bucket = bucket
	options.Params["lifecycle"] = ""
	return api.httpRequestWithUnmarshalXML(options, result)
}

// GetBucketLogging get bucket logging settings
func (api *API) GetBucketLogging(bucket string, result *BucketLoggingStatus) error {
	var options = getDefaultRequestOptions()
	options.Bucket = bucket
	options.Params["logging"] = ""
	return api.httpRequestWithUnmarshalXML(options, result)
}

// CreateBucket create bucket
func (api *API) CreateBucket(bucket string, acl ACLGrant, headers map[string]string) error {
	return api.PutBucket(bucket, acl, "", headers)
}

// PutBucket create bucket
//
//      - bucket: bucket name If bucket exists and not belong to current account, will throw BucketAlreadyExistsError. If bucket not exists, will create a new bucket and set it's ACL
//      - acl: one of private public-read public-read-write
//      - location: the bucket data region location, Current available: oss-cn-hangzhou, oss-cn-qingdao, oss-cn-beijing, oss-cn-hongkong and oss-cn-shenzhen If change exists bucket region, will throw BucketAlreadyExistsError. If region value invalid, will throw InvalidLocationConstraintError.
//      - headers: HTTP header
func (api *API) PutBucket(bucket string, acl ACLGrant, location string, headers map[string]string) error {
	var options = getDefaultRequestOptions()
	options.Method = "PUT"
	options.Bucket = bucket
	if headers != nil {
		options.Headers = headers
	}
	if acl != "" {
		options.Headers["x-oss-acl"] = string(acl)
	}
	if location != "" {
		var config = CreateBucketConfiguration{LocationConstraint: location}
		var data, _ = xml.Marshal(config)
		options.Body = bytes.NewBuffer(data)
	}
	return api.httpRequestWithUnmarshalXML(options, nil)
}

// PutBucketACL create bucket with acl or update bucket acl when bucket is exists
func (api *API) PutBucketACL(bucket string, acl ACLGrant, headers map[string]string) error {
	var options = getDefaultRequestOptions()
	options.Method = "PUT"
	options.Bucket = bucket
	if headers != nil {
		options.Headers = headers
	}
	options.Params["acl"] = ""
	options.Headers["x-oss-acl"] = string(acl)
	return api.httpRequestWithUnmarshalXML(options, nil)
}

// PutBucketLogging update the bucket logging settings.
// Log file will create every one hour and name format: <prefix><bucket>-YYYY-mm-DD-HH-MM-SS-UniqueString.
func (api *API) PutBucketLogging(sourcebucket, targetbucket, prefix string) error {
	var options = getDefaultRequestOptions()
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

// PutBucketWebsite set the bucket as a static website.
//
//      - indexfile: the object that contain index page
//      - errorfile: the object taht contain error page
func (api *API) PutBucketWebsite(bucket, indexfile, errorfile string) error {
	var options = getDefaultRequestOptions()
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

// PutBucketLifecycle set the bucket object lifecycle.
func (api *API) PutBucketLifecycle(bucket string, rule LifecycleRule) error {
	var options = getDefaultRequestOptions()
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

// PutBucketReferer set the bucket request Referer white list.
func (api *API) PutBucketReferer(bucket string, config RefererConfiguration) error {
	var options = getDefaultRequestOptions()
	options.Method = "PUT"
	options.Bucket = bucket
	var data, _ = xml.Marshal(config)
	options.Body = bytes.NewBuffer(data)
	options.Params["referer"] = ""
	var base64md5 = getBase64MD5(data)
	options.Headers["Content-MD5"] = base64md5
	return api.httpRequestWithUnmarshalXML(options, nil)
}

// DeleteBucket delete an empty bucket.
// If bucket is not empty, will throw BucketNotEmptyError. If bucket is not exists, will throw NoSuchBucketError.
func (api *API) DeleteBucket(bucket string) error {
	var options = getDefaultRequestOptions()
	options.Method = "DELETE"
	options.Bucket = bucket
	return api.httpRequestWithUnmarshalXML(options, nil)
}

// DeleteBucketWebsite delete bucket website config
func (api *API) DeleteBucketWebsite(bucket string) error {
	var options = getDefaultRequestOptions()
	options.Method = "DELETE"
	options.Bucket = bucket
	options.Params["website"] = ""
	return api.httpRequestWithUnmarshalXML(options, nil)
}

// DeleteBucketLifecycle delete bucket object lifecycle
func (api *API) DeleteBucketLifecycle(bucket string) error {
	var options = getDefaultRequestOptions()
	options.Method = "DELETE"
	options.Bucket = bucket
	options.Params["lifecycle"] = ""
	return api.httpRequestWithUnmarshalXML(options, nil)
}

// DeleteBucketLogging delete bucket logging settings
func (api *API) DeleteBucketLogging(bucket string) error {
	var options = getDefaultRequestOptions()
	options.Method = "DELETE"
	options.Bucket = bucket
	options.Params["logging"] = ""
	return api.httpRequestWithUnmarshalXML(options, nil)
}

// DeleteBucketReferer delete the bucket request Referer white list.
func (api *API) DeleteBucketReferer(bucket string) error {
	var options = getDefaultRequestOptions()
	options.Method = "DELETE"
	options.Bucket = bucket
	options.Params["referer"] = ""
	return api.httpRequestWithUnmarshalXML(options, nil)
}

// GetObject get an object from the bucket.
func (api *API) GetObject(bucket, object string, headers, params map[string]string) (io.ReadCloser, error) {
	var options = getDefaultRequestOptions()
	options.Bucket = bucket
	options.Object = object
	options.Headers = headers
	options.Params = params

	var res *http.Response
	var err error
	if res, err = api.httpRequest(options); err != nil {
		return nil, err
	}

	return res.Body, nil
}

// GetObjectACL get object acl
func (api *API) GetObjectACL(bucket, object string, result *AccessControlPolicy) error {
	var options = getDefaultRequestOptions()
	options.Bucket = bucket
	options.Object = object
	options.Params["acl"] = ""
	return api.httpRequestWithUnmarshalXML(options, result)
}

// HeadObject head an object and get the meta info.
func (api *API) HeadObject(bucket, object string, headers map[string]string) (result http.Header, err error) {
	var options = getDefaultRequestOptions()
	options.Method = "HEAD"
	options.Bucket = bucket
	options.Object = object
	options.AutoClose = true
	var res *http.Response
	if res, err = api.httpRequest(options); err != nil {
		return
	}
	return res.Header, nil
}

// PutObject add an object to the bucket.
//
//      - bucket: an exists bucket
//      - object: object name store on OSS
//      - body: readable object
//      - headers: HTTP Header
func (api *API) PutObject(bucket, object string, body io.Reader, headers map[string]string) error {
	var options = getDefaultRequestOptions()
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

	options.AutoClose = true

	var err error
	_, err = api.httpRequest(options)
	return err
}

// PostObject is same to PutObject, but use POST method, so just alisa to PutObject
func (api *API) PostObject(bucket, object string, body io.Reader, headers map[string]string) error {
	return api.PutObject(bucket, object, body, headers)
}

// PutObjectACL update object acl
func (api *API) PutObjectACL(bucket, object, acl string) error {
	var options = getDefaultRequestOptions()
	options.Method = "PUT"
	options.Bucket = bucket
	options.Object = object
	options.Headers["x-oss-object-acl"] = acl
	options.Params["acl"] = ""
	options.AutoClose = true
	_, err := api.httpRequest(options)
	return err
}

// CopyObject copy an object from sourceName to name.
func (api *API) CopyObject(sourceBucket, sourceObject, targetBucket, targetObject string,
	headers map[string]string) (result CopyObjectResult, err error) {
	var options = getDefaultRequestOptions()
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

// AppendObject append data to an appendable object
func (api *API) AppendObject(bucket, object string, position int, body io.Reader,
	headers map[string]string) (result http.Header, err error) {

	var options = getDefaultRequestOptions()
	options.Method = "POST"
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
	options.Params["position"] = strconv.Itoa(position)
	options.AutoClose = true
	var res *http.Response
	if res, err = api.httpRequest(options); err != nil {
		return
	}

	result = res.Header

	return

}

// DeleteObject delete an object from the bucket.
func (api *API) DeleteObject(bucket, object string) error {
	var options = getDefaultRequestOptions()
	options.Method = "DELETE"
	options.Bucket = bucket
	options.Object = object
	return api.httpRequestWithUnmarshalXML(options, nil)
}

// DeleteObjects delete multi objects in one request.
func (api *API) DeleteObjects(bucket string, objects []string, result *DeleteResult) error {
	var options = getDefaultRequestOptions()
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

// NewMultipartUpload initial multipart upload
func (api *API) NewMultipartUpload(bucket, object string, headers map[string]string) (*MultipartUpload, error) {
	var options = getDefaultRequestOptions()
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

// GetMultiPartUpload get multipart upload
func (api *API) GetMultiPartUpload(bucket, object, uploadID string) (*MultipartUpload, error) {
	var multi = new(MultipartUpload)
	multi.Bucket = bucket
	multi.Key = object
	multi.UploadID = uploadID
	multi.api = api
	return multi, nil
}

// UploadPart upload the content of io.Reader as one part.
func (multi *MultipartUpload) UploadPart(partNumber int, body io.Reader) (string, error) {
	var options = getDefaultRequestOptions()
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

	options.AutoClose = true

	var res *http.Response
	var err error
	if res, err = multi.api.httpRequest(options); err != nil {
		return "", err
	}
	return res.Header.Get("ETag"), nil
}

// CopyPart upload a part with data copy from srouce object in source bucket
func (multi *MultipartUpload) CopyPart(sourceBucket, sourceObject string, partNumber int,
	sourceRange string, headers map[string]string) (string, error) {

	var options = getDefaultRequestOptions()
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

// CompleteUpload finish multiupload and merge all the parts as a object.
func (multi *MultipartUpload) CompleteUpload(parts []Part, result *CompleteMultipartUploadResult) error {
	var options = getDefaultRequestOptions()
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

// AbortUpload cancel multiupload and delete all parts
func (multi *MultipartUpload) AbortUpload() error {
	var options = getDefaultRequestOptions()
	options.Method = "DELETE"
	options.Bucket = multi.Bucket
	options.Object = multi.Key
	options.Params["uploadId"] = multi.UploadID
	options.AutoClose = true
	_, err := multi.api.httpRequest(options)
	return err
}

// ListMultipartUploadOptions list multipart upload options
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

// GetDefaultListMultipartUploadOptions get default list multipart upload options
func GetDefaultListMultipartUploadOptions() *ListMultipartUploadOptions {
	var options = new(ListMultipartUploadOptions)
	options.Params = make(map[string]string)
	return options
}

// ListMultipartUpload list all multipart uploads and their parts
func (api *API) ListMultipartUpload(bucket string, opts *ListMultipartUploadOptions) ([]*MultipartUpload, error) {
	var options = getDefaultRequestOptions()
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

// ListParts list all upload parts of current upload_id
func (multi *MultipartUpload) ListParts(maxParts, partNumberMarker int,
	result *ListPartsResult) error {
	var options = getDefaultRequestOptions()
	options.Method = "GET"
	options.Bucket = multi.Bucket
	options.Object = multi.Key
	options.Params["uploadId"] = multi.UploadID
	return multi.api.httpRequestWithUnmarshalXML(options, &result)
}

// PutBucketCORS put bucket cors
func (api *API) PutBucketCORS(bucket string, config CORSConfiguration) error {
	var options = getDefaultRequestOptions()
	options.Method = "PUT"
	options.Bucket = bucket
	var data, _ = xml.Marshal(config)
	options.Body = bytes.NewBuffer(data)
	options.Params["cors"] = ""
	return api.httpRequestWithUnmarshalXML(options, nil)
}

// GetBucketCORS Get bucket cors
func (api *API) GetBucketCORS(bucket string, result *CORSConfiguration) error {
	var options = getDefaultRequestOptions()
	options.Bucket = bucket
	options.Params["cors"] = ""
	return api.httpRequestWithUnmarshalXML(options, result)
}

// DeleteBucketCORS Delete bucket cors
func (api *API) DeleteBucketCORS(bucket string) error {
	var options = getDefaultRequestOptions()
	options.Method = "DELETE"
	options.Bucket = bucket
	options.Params["cors"] = ""
	return api.httpRequestWithUnmarshalXML(options, nil)
}

// OptionObject options object to determine if user can send the actual HTTP request
func (api *API) OptionObject(bucket, object string, headers map[string]string) (http.Header, error) {
	var options = getDefaultRequestOptions()
	options.Method = "OPTIONS"
	options.Bucket = bucket
	options.Object = object
	options.Headers = headers
	options.AutoClose = true
	var err error
	var res *http.Response
	if res, err = api.httpRequest(options); err != nil {
		return nil, err
	}
	return res.Header, nil
}

// UploadLargeFile upload large file, the content is read from filename.
// The large file is splitted into many parts.
// It will put the many parts into bucket and then merge all the parts into one object.
func (api *API) UploadLargeFile(bucket, object, fileName string, bufSize int64,
	headers map[string]string) (result CompleteMultipartUploadResult, err error) {

	if bufSize < 1024*100 {
		err = fmt.Errorf("bufSize must big than %d", 1024*100)
		return
	}

	var fp *os.File
	if fp, err = os.Open(fileName); err != nil {
		return
	}

	defer fp.Close()

	var stat, _ = fp.Stat()
	var fileSize = stat.Size()
	if fileSize <= bufSize {
		err = fmt.Errorf("need a large file, and it's size big than %d\n", bufSize)
		return
	}
	var filePart = int(fileSize / bufSize)
	if int64(filePart)*bufSize < fileSize {
		filePart = filePart + 1
	}

	if api.debug {
		log.Printf("Upload large file: %s\ntotal part: %d\n", fileName, filePart)
	}

	var rd io.Reader
	var parts = make([]Part, filePart)
	var etag string
	var uploadFailed = false

	var multi *MultipartUpload
	if multi, err = api.NewMultipartUpload(bucket, object, headers); err != nil {
		return
	}

	for i := 1; i <= filePart; i++ {
		rd = io.LimitReader(fp, bufSize)
		if etag, err = multi.UploadPart(i, rd); err != nil {
			uploadFailed = true
			break
		}
		if api.debug {
			log.Printf("PartNumber: %d, ETag: %s\n", i, etag)
		}
		parts[i-1] = Part{
			PartNumber: i,
			ETag:       etag,
		}
	}

	if uploadFailed {
		multi.AbortUpload()
		err = errors.New("multi upload file failed")
	} else {
		err = multi.CompleteUpload(parts, &result)
	}

	return
}
