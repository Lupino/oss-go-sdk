# How to upload a large file to OSS

In this tutorial you will use the `io.LimitReader` to spilt a large file,
and use `oss.MultipartUpload.UploadPart` to upload a file.

First read [getting start](../getting_start)

## Prepare

Get a large file, or create one, and rename to `test.bin`

```bash
wget -O test.bin http://speedtest.tokyo.linode.com/100MB-tokyo.bin
```

## Split the large file

simple use `io.LimitReader` to split a large file

```go
var err error
var fp *os.File

var bufSize = int64(1024 * 1024 * 2) // 2 M
var rd = io.LimitReader(fp, bufSize)
```

## upload the large file

Initial multipart upload with bucket `ossgosdklargefile` or the bucket you want.

```go
var bucket = "ossgosdklargefile"
var object = "largefile.bin"
var multi *oss.MultipartUpload
if multi, err = OSSAPI.NewMultipartUpload(bucket, object, nil); err != nil {
	var e = oss.ParseError(err)
	log.Printf("Code: %s\nMessage: %s\n", e.Code, e.Message)
}
```

use `multi.UploadPart` to upload the file part

```go
var etag string

if etag, err = multi.UploadPart(partNumber, rd); err != nil {
	var e = oss.ParseError(err)
	log.Printf("Code: %s\nMessage: %s\n", e.Code, e.Message)
}
```

Once part upload complete, complete upload use `multi.CompleteUpload` tell the OSS the multipart upload is complete, OSS will merge the large file.

```go
var parts = []oss.Part{
    Part{
        PartNumber: partNumber,
        ETag: etag,
    },
    ...
}
var result oss.CompleteMultipartUploadResult

if err = multi.CompleteUpload(parts, &result); err != nil {
	var e = oss.ParseError(err)
	log.Printf("Code: %s\nMessage: %s\n", e.Code, e.Message)
}
```
## The end

the source code [main.go](main.go)
