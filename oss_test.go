package oss

import (
	"fmt"
	"testing"
)

var api *API
var options *APIOptions

func init() {
	options = GetDefaultAPIOptioins()
	api = NewAPI(options)
}

func TestPrint(t *testing.T) {
	fmt.Println(AGENT)
}

func TestSetValue(t *testing.T) {
	api.SetTimeout(5)
	api.SetDebug()
	api.SetRetryTimes(5)
	api.SetSendBufferSize(1024)
	api.SetRecvBufferSize(1024 * 1024)
	api.SetIsOSSHost(false)
}
