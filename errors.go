package oss

import (
	"encoding/xml"
)

// Error Each error return by OSS server
type Error struct {
	XMLName xml.Name `xml:"Error"`
	// error name
	Code string
	// error message
	Message string
	// uuid for this request,
	// if you meet some unhandled problem,
	// you can send this request id to OSS engineer to find out what's happend.
	RequestID string `xml:"RequestId"`
	// OSS cluster name for this request
	HostID string `xml:"HostId"`
	// error xml return by OSS server
	Raw []byte `xml:"-"`
}

// Error returns the underlying error's message.
func (e *Error) Error() string {
	return e.Message
}

// parseError parse the error return by OSS server
func parseError(errStr []byte) *Error {
	var err Error
	xml.Unmarshal(errStr, &err)
	err.Raw = errStr
	return &err
}
