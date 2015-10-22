package oss

import (
	"testing"
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
		t.Fatalf("isOSSHost: expect: false, got true")
	}
	got = isOSSHost("example.com", true)
	if !got {
		t.Fatalf("isOSSHost: expect: true, got false")
	}
	got = isOSSHost("aliyuncs.com", false)
	if !got {
		t.Fatalf("isOSSHost: expect: true, got false")
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
	var expecttAuthValue = "z3GSMKAock34CpDqxmdTEg81V0k="
	if authValue != expecttAuthValue {
		t.Fatalf("authValue: expect: %s, got %s\n", expecttAuthValue, authValue)
	}
	headers["x-oss-key"] = "test-x-oss-key"
	authValue = getAssign(secretAccessKey, "GET", headers, "/", nil, false)
	// fmt.Printf("authValue: %s\n", authValue)
	expecttAuthValue = "QqFGy3l4JKba4YL2FXrTgVoYVMk="
	if authValue != expecttAuthValue {
		t.Fatalf("authValue: expect: %s, got %s\n", expecttAuthValue, authValue)
	}
}
