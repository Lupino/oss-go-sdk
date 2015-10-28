package oss

import (
	"fmt"
	"testing"
)

func TestParseError(t *testing.T) {
	err := []byte(`<?xml version="1.0" ?>
<Error xmlns="http://doc.oss-cn-hangzhou.aliyuncs.com">
        <Code>InvalidArgument</Code>
        <Message>this is the error message.</Message>
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

	var e = parseError(err)
	if e.Code != "InvalidArgument" {
		t.Fatalf("ParseError: except: %s, but got: %s\n", "InvalidArgument", e.Code)
	}
	var realErr = error(e)
	fmt.Printf("%v\n", realErr.(*Error).Code)
	fmt.Printf("error: %s\n", e.Error())
}
