package oss

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// PROVIDER defined provider
const PROVIDER = "OSS"

// SelfDefineHeaderPrefix defined oss header prefix
const SelfDefineHeaderPrefix = "x-oss-"

// OSSHostList defined OSS host list
var OSSHostList = []string{"aliyun-inc.com", "aliyuncs.com", "alibaba.net", "s3.amazonaws.com"}

func getHostFromList(hosts string) string {
	var tmpList = strings.Split(hosts, ",")
	var host string
	var port int
	if len(tmpList) <= 1 {
		host, port = getHostPort(hosts)
		return fmt.Sprintf("%s:%d", host, port)
	}
	for _, tmpHost := range tmpList {
		host, port = getHostPort(tmpHost)
		if _, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port)); err == nil {
			return fmt.Sprintf("%s:%d", host, port)
		}
	}
	host, port = getHostPort(tmpList[0])
	return fmt.Sprintf("%s:%d", host, port)
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

func isOSSHost(host string, isOSSHost bool) bool {
	if isOSSHost {
		return true
	}
	for _, OSSHost := range OSSHostList {
		if strings.Contains(host, OSSHost) {
			return true
		}
	}
	return false
}

// getAssign Create the authorization for OSS based on header input.
// You should put it into "Authorization" parameter of header.
func getAssign(secretAccessKey, method string, headers map[string]string,
	resource string, result []string, debug bool) string {

	var contentMd5, contentType, date, canonicalizedOSSHeaders string
	if debug {
		log.Printf("secretAccessKey: %s", secretAccessKey)
	}
	contentMd5 = safeGetElement("Content-MD5", headers)
	contentType = safeGetElement("Content-Type", headers)
	date = safeGetElement("Date", headers)
	var canonicalizedResource = resource
	var tmpHeaders = formatHeader(headers)
	if len(tmpHeaders) > 0 {
		var xHeaderList = make([]string, 0)
		for k := range tmpHeaders {
			if strings.HasPrefix(k, SelfDefineHeaderPrefix) {
				xHeaderList = append(xHeaderList, k)
			}
		}
		sort.Strings(xHeaderList)
		for _, k := range xHeaderList {
			canonicalizedOSSHeaders = fmt.Sprintf("%s%s:%s\n", canonicalizedOSSHeaders, k, tmpHeaders[k])
		}
	}
	var stringToSign = fmt.Sprintf("%s\n%s\n%s\n%s\n%s%s", method, contentMd5, contentType, date, canonicalizedOSSHeaders, canonicalizedResource)
	result = append(result, stringToSign)

	if debug {
		log.Printf("method:%s\n content_md5:%s\n content_type:%s\n data:%s\n canonicalized_oss_headers:%s\n canonicalized_resource:%s\n", method, contentMd5, contentType, date, canonicalizedOSSHeaders, canonicalizedResource)
		log.Printf("string_to_sign:%s\n \nlength of string_to_sign:%d\n", stringToSign, len(stringToSign))

	}
	var h = hmac.New(sha1.New, []byte(secretAccessKey))
	h.Write([]byte(stringToSign))
	var signResult = base64.StdEncoding.EncodeToString(h.Sum(nil))

	if debug {
		log.Printf("sign result: %s", signResult)
	}

	return signResult
}

func safeGetElement(name string, container map[string]string) string {
	for k, v := range container {
		if strings.Trim(strings.ToLower(k), " ") == strings.Trim(strings.ToLower(name), " ") {
			return v
		}
	}
	return ""
}

// formatHeader format the headers that self define
//  convert the self define headers to lower.
func formatHeader(headers map[string]string) map[string]string {
	var tmpHeaders = make(map[string]string)
	for k, v := range headers {

		var lower = strings.ToLower(k)
		if strings.HasPrefix(lower, SelfDefineHeaderPrefix) {
			lower = strings.Trim(lower, " ")
			tmpHeaders[lower] = v
		} else {
			tmpHeaders[strings.Trim(k, " ")] = v
		}
	}
	return tmpHeaders
}

func getResource(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}
	var tmpHeaders = make(map[string]string)
	for k, v := range params {
		tmpHeaders[strings.Trim(strings.ToLower(k), " ")] = v
	}
	var overrideResponseList = []string{"response-content-type", "response-content-language",
		"response-cache-control", "logging", "response-content-encoding",
		"acl", "uploadId", "uploads", "partNumber", "group", "link",
		"delete", "website", "location", "objectInfo",
		"response-expires", "response-content-disposition", "cors", "lifecycle",
		"restore", "qos", "referer", "append", "position"}

	sort.Strings(overrideResponseList)

	var resource = ""
	var separator = "?"
	for _, k := range overrideResponseList {
		if v, ok := tmpHeaders[strings.ToLower(k)]; ok {
			resource = fmt.Sprintf("%s%s%s", resource, separator, k)
			if len(v) != 0 {
				resource = fmt.Sprintf("%s=%s", resource, v)
			}
			separator = "&"
		}
	}

	return resource
}

func quote(str string) string {
	return url.QueryEscape(str)
}

func isIP(s string) bool {
	var host, _ = getHostPort(s)
	if host == "localhost" {
		return true
	}

	var tmpList = strings.Split(host, ".")
	if len(tmpList) != 4 {
		return false
	}
	for _, i := range tmpList {
		tmpI, ok := strconv.Atoi(i)
		if ok != nil || tmpI < 0 || tmpI > 255 {
			return false
		}
	}
	return true
}

// appendParam convert the parameters to query string of URI
func appendParam(uri string, params map[string]string) string {
	var values = url.Values{}
	for k, v := range params {
		k = strings.Replace(k, "_", "-", -1)
		if k == "maxkeys" {
			k = "max-keys"
		}

		if k == "acl" {
			v = ""
		}
		values.Set(k, v)
	}
	return uri + "?" + values.Encode()
}

func checkBucketValid(bucket string) bool {
	var alphabeta = "^[abcdefghijklmnopqrstuvwxyz0123456789-]+$"
	if len(bucket) < 3 || len(bucket) > 63 {
		return false
	}
	if bucket[len(bucket)-1] == '-' || bucket[len(bucket)-1] == '_' {
		return false
	}
	if !((bucket[0] >= 'a' && bucket[0] <= 'z') || (bucket[0] >= '0' && bucket[0] <= '9')) {
		return false

	}

	if matched, _ := regexp.MatchString(alphabeta, bucket); !matched {
		return false
	}
	return true
}

func getBase64MD5(str []byte) string {
	h := md5.New()
	h.Write(str)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func getBase64MD5WithReader(reader io.Reader) string {
	h := md5.New()
	io.Copy(h, reader)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// ParseError defined parse oss api error
func ParseError(err error) (e Error) {
	var data = []byte(err.Error())
	xml.Unmarshal(data, &e)
	return
}
