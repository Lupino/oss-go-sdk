package oss

import (
	"errors"
	"testing"
	//    "fmt"
)

func TestGetHostFromList(t *testing.T) {
	var host string
	host = getHostFromList("oss.aliyuncs.com")
	if host != "oss.aliyuncs.com:80" {
		t.Fatalf("host: except: %s, got: %s", "oss.aliyuncs.com:80", host)
	}

	host = getHostFromList("example.com,example1.com,oss.aliyuncs.com")
	if host != "example.com:80" {
		t.Fatalf("host: except: %s, got: %s", "example.com:80", host)
	}
	host = getHostFromList("example.com:90,example1.com:90")
	if host != "example.com:90" {
		t.Fatalf("host: except: %s, got: %s", "example.com:90", host)
	}
}

func testPanic(t *testing.T) {
	x := recover()
	if x == nil {
		t.Fatalf("there no panic")
	}
}

func TestGetHostPort(t *testing.T) {
	var host, port = getHostPort("example.com")
	if host != "example.com" || port != 80 {
		t.Fatalf("host: except: %s:%d, got: %s:%d", "example.com", 80, host, port)
	}

	host, port = getHostPort("example.com:443")
	if host != "example.com" || port != 443 {
		t.Fatalf("host: except: %s:%d, got: %s:%d", "example.com", 443, host, port)
	}

	defer testPanic(t)
	host, port = getHostPort("example.com:invaid")
}

func TestIsOSSHost(t *testing.T) {
	var got = isOSSHost("example.com", false)
	if got {
		t.Fatalf("isOSSHost: except: false, got true")
	}
	got = isOSSHost("example.com", true)
	if !got {
		t.Fatalf("isOSSHost: except: true, got false")
	}
	got = isOSSHost("aliyuncs.com", false)
	if !got {
		t.Fatalf("isOSSHost: except: true, got false")
	}
}

func TestGetAssign(t *testing.T) {
	var headers = make(map[string]string)
	headers["Content-MD5"] = "81f770d0950fe6a2c158fc7ee9cfb6d8"
	headers["Content-type"] = "image/jpeg"
	headers["date"] = "Wed, 21 Oct 2015 07:17:58 GMT"

	var secretAccessKey = "secretAccessKey"
	var authValue = getAssign(secretAccessKey, "GET", headers, "/", nil, true)
	// fmt.Printf("authValue: %s\n", authValue)
	var excepttAuthValue = "z3GSMKAock34CpDqxmdTEg81V0k="
	if authValue != excepttAuthValue {
		t.Fatalf("authValue: except: %s, got %s\n", excepttAuthValue, authValue)
	}
	headers["x-oss-key"] = "test-x-oss-key"
	authValue = getAssign(secretAccessKey, "GET", headers, "/", nil, false)
	// fmt.Printf("authValue: %s\n", authValue)
	excepttAuthValue = "QqFGy3l4JKba4YL2FXrTgVoYVMk="
	if authValue != excepttAuthValue {
		t.Fatalf("authValue: except: %s, got %s\n", excepttAuthValue, authValue)
	}
}

func TestSafeGetElement(t *testing.T) {
	var container = map[string]string{"key": "value"}
	var got = safeGetElement("key", container)
	if got != "value" {
		t.Fatalf("safeGetElement: except: %s, but got: %s\n", "value", got)
	}

	got = safeGetElement("key1", container)
	if got != "" {
		t.Fatalf("safeGetElement: except: %s, but got: %s\n", "", got)
	}
}

func TestGetResouse(t *testing.T) {
	var params = make(map[string]string)
	var got = getResource(params)
	if got != "" {
		t.Fatalf("getResource: except: %s, but got: %s\n", "", got)
	}
	params["referer"] = "http://example.com/test"
	params["qos"] = "100"
	params["other"] = "other"
	params["other1"] = "other1"
	got = getResource(params)
	var except = "?qos=100&referer=http://example.com/test"
	if got != except {
		t.Fatalf("getResource: except: %s, but got: %s\n", except, got)
	}
}

func TestQuote(t *testing.T) {
	var got = quote("test.com/test?key=value&key1=value1")
	var except = "test.com%2Ftest%3Fkey%3Dvalue%26key1%3Dvalue1"
	if got != except {
		t.Fatalf("quote: except: %s, but got: %s\n", except, got)
	}
}

func TestIsIP(t *testing.T) {
	var got = isIP("localhost")
	var except = true
	if got != except {
		t.Fatalf("isIP: except: %s, but got: %s\n", except, got)
	}
	got = isIP("test.com:90")
	except = false
	if got != except {
		t.Fatalf("isIP: except: %s, but got: %s\n", except, got)
	}
	got = isIP("192.168.1.189:90")
	except = true
	if got != except {
		t.Fatalf("isIP: except: %s, but got: %s\n", except, got)
	}
	got = isIP("aa.168.1.189:90")
	except = false
	if got != except {
		t.Fatalf("isIP: except: %s, but got: %s\n", except, got)
	}
}

func TestAppendParam(t *testing.T) {
	var uri = "/test"
	var params = make(map[string]string)
	params["content-type"] = "image/jpeg"
	params["maxkeys"] = "maxkeys"
	params["x_oss_key_value"] = "nothing"
	params["referer"] = "http://test.com/test?key=value&aa"
	params["test&&example"] = "value"
	params["acl"] = "acl"
	var got = appendParam(uri, params)
	var except = "/test?acl=&content-type=image%2Fjpeg&max-keys=maxkeys&referer=http%3A%2F%2Ftest.com%2Ftest%3Fkey%3Dvalue%26aa&test%26%26example=value&x-oss-key-value=nothing"
	if got != except {
		t.Fatalf("appendParam: except: %s, but got: %s\n", except, got)
	}
}

func TestCheckBucketValid(t *testing.T) {
	var got = checkBucketValid("bucket")
	var except = true
	if got != except {
		t.Fatalf("checkBucketValid: except: %s, but got: %s\n", except, got)
	}
	got = checkBucketValid("-bucket")
	except = false
	if got != except {
		t.Fatalf("checkBucketValid: except: %s, but got: %s\n", except, got)
	}
	got = checkBucketValid("_bucket")
	except = false
	if got != except {
		t.Fatalf("checkBucketValid: except: %s, but got: %s\n", except, got)
	}
	got = checkBucketValid("bucket-")
	except = false
	if got != except {
		t.Fatalf("checkBucketValid: except: %s, but got: %s\n", except, got)
	}
	got = checkBucketValid("bucket-")
	except = false
	if got != except {
		t.Fatalf("checkBucketValid: except: %s, but got: %s\n", except, got)
	}
	got = checkBucketValid("bucKet")
	except = false
	if got != except {
		t.Fatalf("checkBucketValid: except: %s, but got: %s\n", except, got)
	}
}

func TestParseError(t *testing.T) {
	err := errors.New(`<?xml version="1.0" ?>
<Error xmlns="http://doc.oss-cn-hangzhou.aliyuncs.com">
        <Code>InvalidArgument</Code>
        <Message>
        </Message>
        <ArgumentValue>
                error-acl
        </ArgumentValue>
        <ArgumentName>
                x-oss-acl
        </ArgumentName>
        <RequestId>
                4e63c87a-71dc-87f7-11b5-583a600e0038
        </RequestId>
        <HostId>
                oss-cn-hangzhou.aliyuncs.com
        </HostId>
</Error>`)

	var e = ParseError(err)
	if e.Code != "InvalidArgument" {
		t.Fatalf("ParseError: except: %s, but got: %s\n", "InvalidArgument", e.Code)
	}
}
