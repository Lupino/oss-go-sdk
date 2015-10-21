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
