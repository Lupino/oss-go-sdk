package oss

import (
	"encoding/xml"
	"time"
)

// Bucket defined bucket
type Bucket struct {
	XMLName      xml.Name `xml:"Bucket"`
	Location     string
	Name         string
	CreationDate time.Time
}

// Owner defined owner
type Owner struct {
	XMLName     xml.Name `xml:"Owner"`
	ID          string
	DisplayName string
}

// ListAllMyBucketsResult defined list all my buckets result
type ListAllMyBucketsResult struct {
	XMLName     xml.Name `xml:"ListAllMyBucketsResult"`
	Prefix      string
	Marker      string
	MaxKeys     string
	IsTruncated bool
	NextMarker  string
	Owner       Owner
	Buckets     []Bucket `xml:"Buckets>Bucket"`
}

// Content defined get bucket content
type Content struct {
	XMLName      xml.Name `xml:"Contents"`
	Key          string
	LastModified time.Time
	Etag         string
	Type         string
	Size         int
	StorageClass string
	Owner        Owner
}

// ListBucketResult defined list bucket result
type ListBucketResult struct {
	XMLName        xml.Name `xml:"ListBucketResult"`
	Contents       []Content
	CommonPrefixes string
	Delimiter      string
	IsTruncated    bool
	Marker         string
	MaxKeys        string
	Name           string
	Owner          Owner
	Prefix         string
	EncodingType   string `xml:"encoding-type"`
}

// AccessControlPolicy defined access control policy
type AccessControlPolicy struct {
	XMLName           xml.Name `xml:"AccessControlPolicy"`
	Owner             Owner
	AccessControlList []string `xml:"AccessControlList>Grant"`
}

// LocationConstraint defined location constraint
type LocationConstraint string

// BucketLoggingStatus defined bucket logging status
type BucketLoggingStatus struct {
	XMLName xml.Name `xml:"BucketLoggingStatus"`
	Bucket  string   `xml:"LoggingEnabled>TargetBucket"`
	Prefix  string   `xml:"LoggingEnabled>TargetPrefix"`
}

// WebsiteConfiguration defined website configuration
type WebsiteConfiguration struct {
	XMLName     xml.Name `xml:"WebsiteConfiguration"`
	IndexSuffix string   `xml:"IndexDocument>Suffix"`
	ErrorKey    string   `xml:"ErrorDocument>Key"`
}

// Error defined error result
type Error struct {
	XMLName    xml.Name `xml:"Error"`
	Code       string
	Message    string
	BucketName string
	RequestID  string `xml:"RequestId"`
	HostID     string `xml:"HostId"`
}

// RefererConfiguration defined referer configuration
type RefererConfiguration struct {
	XMLName           xml.Name `xml:"RefererConfiguration"`
	AllowEmptyReferer bool
	RefererList       []string `xml:"RefererList>Referer"`
}

// LifecycleRule defined lifecycle configuration rule
type LifecycleRule struct {
	XMLName        xml.Name `xml:"Rule"`
	ID             string
	Prefix         string
	Status         string
	ExpirationDays int       `xml:"Expiration>Days"`
	ExpirationDate time.Time `xml:"Expiration>Date"`
}

// LifecycleConfiguration defined lifecycle configuration
type LifecycleConfiguration struct {
	XMLName xml.Name `xml:"LifecycleConfiguration"`
	Rule    LifecycleRule
}

// CreateBucketConfiguration defined create bucket configuration
type CreateBucketConfiguration struct {
	XMLName            xml.Name `xml:"CreateBucketConfiguration"`
	LocationConstraint string
}
