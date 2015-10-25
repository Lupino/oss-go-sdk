# How to use OSS as a static website.

In this tutorial will tell you use the `PutBucketWebsite` to build a static website.

First read [getting start](../getting_start)

## create a website able bucket

I' ll use bucket name `ossgosdkwebsite` to build the website.

```go
var bucket = "ossgosdkwebsite"
```

first create a public read write bucket.

```go
if err = OSSAPI.PutBucket(bucket, oss.ACLPublicReadWrite, nil, nil); err != nil {
    var e = oss.ParseError(err)
    log.Printf("Code: %s\nMessage: %s\n", e.Code, e.Message)
}
```

Once create the bucket, set it is website config.

```go
if err = OSSAPI.PutBucketWebsite(bucket, "index.html", "error.html"); err != nil {
	var e = oss.ParseError(err)
    log.Printf("Code: %s\nMessage: %s\n", e.Code, e.Message)
}
```

## create a static website

the website file list is:
```bash
.
├── css
│   └── screen.css
├── error.html
├── index.html
└── js
    └── application.js
```

write the file what you want.

## upload static website

first create an `upload` function.

```go
func upload(OSSAPI *oss.API, bucket, filename string) {
	fp, err := os.Open(filename)
    defer fp.Close()
	if err != nil {
		log.Fatal(err)
	}

	var headers = make(map[string]string)
	headers["Content-Type"] = "text/html"

	if err = OSSAPI.PutObject(bucket, filename, fp, headers); err != nil {
		var e = oss.ParseError(err)
        log.Printf("Code: %s\nMessage: %s\n", e.Code, e.Message)
	}
}
```

then upload the static site

```go
upload(OSSAPI, bucket, "index.html")
upload(OSSAPI, bucket, "error.html")
upload(OSSAPI, bucket, "css/screen.css")
upload(OSSAPI, bucket, "js/application.js")
```

now you visit the website on <http://ossgosdkwebsite.oss-cn-hangzhou.aliyuncs.com>

## The end

there some wrong with site browser will download the file

the source code [main.go](main.go)
