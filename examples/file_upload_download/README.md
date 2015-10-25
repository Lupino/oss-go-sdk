# how to upload file and download file use `oss-go-sdk`

In this tutorial you will use the `PutObject` to upload a file,
and use `GetObject` to download a file.

First read [getting start](../getting_start)

## Preapare

Upload a file  need a exists bucket,
if you create the bucket skip this setup.

```go
var bucket = "ossgosdkfileuploaddownload"

if err = OSSAPI.PutBucket(bucket, oss.ACLPublicReadWrite, nil, nil); err != nil {
	var e = oss.ParseError(err)
	log.Printf("Code: %s\nMessage: %s\n", e.Code, e.Message)
}
```

## upload a file

Use `PutObject` to upload a file.

First defined a function upload.

```go
func upload(OSSAPI *oss.API, bucket, filename, contentType string) {
	fp, err := os.Open(filename)
	defer fp.Close()
	if err != nil {
		log.Fatal(err)
	}

	var headers = make(map[string]string)
	headers["Content-Type"] = contentType

	if err = OSSAPI.PutObject(bucket, filename, fp, headers); err != nil {
		var e = oss.ParseError(err)
		log.Printf("Code: %s\nMessage: %s\n", e.Code, e.Message)
	}
}
```

Once `upload` created upload the <test.jpg> to OSS

```go
upload(OSSAPI, bucket, "test.jpg", "image/jpeg")
```

then you can view the file on <http://ossgosdkfileuploaddownload.oss-cn-hangzhou.aliyuncs.com/test.jpg>

## download a file from OSS

Use `GetObject` to download file from bucket.

`GetObject` return a `io.Reader` that use `io.Copy` copy the return data to a file.

first import the require package `os`, `io`, then defined `download` function.

```go
func download(OSSAPI *oss.API, bucket, object, filename string) {
	var fp, err = os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()
	var reader io.Reader
	if reader, err = OSSAPI.GetObject(bucket, object, nil, nil); err != nil {
		log.Fatal(err)
	}
	io.Copy(fp, reader)
}
```

Once create the `download`, download <test.jpg>
```go
download(OSSAPI, bucket, "test.jpg", "test-download.jpg")
```

then you can find the `test-download.jpg` on the current dectory.

## The end

the source code [main.go](main.go)
