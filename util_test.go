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
