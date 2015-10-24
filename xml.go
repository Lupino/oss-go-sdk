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
	ETag         string
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
	ExpirationDays int `xml:"Expiration>Days"`
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

// CopyObjectResult defined copy object result
type CopyObjectResult struct {
	XMLName      xml.Name `xml:"CopyObjectResult"`
	LastModified time.Time
	ETag         string
}

// ObjectKey defined delete multiple objects object key
type ObjectKey struct {
	XMLName xml.Name `xml:"Object"`
	Key     string
}

// DeleteXML defined delete multiple objects xml
type DeleteXML struct {
	XMLName xml.Name `xml:"Delete"`
	Quiet   bool
	Objects []ObjectKey `xml:"--"`
}

// DeleteResult defined delete result
type DeleteResult struct {
	XMLName xml.Name `xml:"DeleteResult"`
	Objects []string `xml:"Deleted>Key"`
}

// InitiateMultipartUploadResult defined initiate multipart upload result
type InitiateMultipartUploadResult struct {
	XMLName  xml.Name `xml:"InitiateMultipartUploadResult"`
	Bucket   string
	Key      string
	UploadID string `xml:"UploadId"`
}

// CopyPartResult defined copy part result
type CopyPartResult struct {
	XMLName      xml.Name `xml:"CopyPartResult"`
	LastModified string
	ETag         string
}

// Part defined part xml
type Part struct {
	XMLName      xml.Name `xml:"Part"`
	PartNumber   int
	ETag         string
	LastModified time.Time
	Size         int
}

// CompleteMultipartUpload defined complete multipart upload xml
type CompleteMultipartUpload struct {
	XMLName xml.Name `xml:"CompleteMultipartUpload"`
	Parts   []Part
}

// CompleteMultipartUploadResult defined complete multipart upload result xml
type CompleteMultipartUploadResult struct {
	XMLName  xml.Name `xml:"CompleteMultipartUploadResult"`
	Location string
	Bucket   string
	Key      string
	ETag     string
}

// Upload defined upload xml
type Upload struct {
	XMLName   xml.Name `xml:"Upload"`
	Key       string
	UploadID  string `xml:"UploadId"`
	Initiated time.Time
}

// ListMultipartUploadsResult defined list multipart uploads result
type ListMultipartUploadsResult struct {
	XMLName            xml.Name `xml:"ListMultipartUploadsResult"`
	Bucket             string
	KeyMarker          string
	UploadIDMarker     string `xml:"UploadIdMarker"`
	NextKeyMarker      string
	NextUploadIDMarker string `xml:"NextUploadIdMarker"`
	Delimiter          string
	Prefix             string
	MaxUploads         string
	IsTruncated        bool
	Uploads            []Upload
}

// ListPartsResult defined list parts result
type ListPartsResult struct {
	XMLName              xml.Name `xml:"ListPartsResult"`
	Bucket               string
	Key                  string
	UploadID             string `xml:"UploadId"`
	NextPartNumberMarker int
	MaxParts             int
	IsTruncated          bool
	Parts                []Part
}

// CORSRule defined cors rule
type CORSRule struct {
	XMLName       xml.Name `xml:"CORSRule"`
	AllowedOrigin []string
	AllowedMethod []string
	AllowedHeader []string
	ExposeHeader  []string
	MaxAgeSeconds int
}

// CORSConfiguration defined cors configuration
type CORSConfiguration struct {
	XMLName xml.Name `xml:"CORSConfiguration"`
	Rules   []CORSRule
}
