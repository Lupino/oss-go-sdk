oss-go-sdk
==================================

This repo is under development.

[![Build Status](https://travis-ci.org/Lupino/oss-go-sdk.svg?branch=master)](https://travis-ci.org/Lupino/oss-go-sdk)
[![Coveralls](https://coveralls.io/repos/Lupino/oss-go-sdk/badge.png?branch=master)](https://coveralls.io/r/Lupino/oss-go-sdk)

aliyun OSS(open storage service) golang client.

## Install

```bash
go get -v github.com/Lupino/oss-go-sdk
```

## License

[Apache2.0](LICENSE)

# OSS Usage

OSS, Open Storage Service. Equal to well known Amazon [S3](http://aws.amazon.com/s3/).

## Create Account

Go to [OSS website](http://www.aliyun.com/product/oss/?lang=en), create a new account for new user.

After account created, you can create the OSS instance and get the `accessKeyId` and `accessKeySecret`.

## Initial OSS API

```go
import (
    "github.com/Lupino/oss-go-sdk"
)
var APIOptions = oss.GetDefaultAPIOptioins()
APIOptions.AccessID = AccessKeyID
APIOptions.SecretAccessKey = AccessKeySecret
var OSSAPI = oss.NewAPI(APIOptions)
```

## Get Service

```go
var result oss.ListAllMyBucketsResult
var headers = make(map[string]string)
var err error
err = OSSAPI.GetService(&result, headers)
```

## Parse the error

all the error is the error xml so you shoud parse the error and see the error message.

```go
var e = oss.ParseError(err)
```

## Tutorial

* [Getting start with oss-go-sdk](examples/getting_start)
* [How to create a static website on OSS](examples/website)
* [How to upload file and download file use oss-go-sdk](examples/file_upload_download)
* [A lite notebook use `AppendObject`](examples/notebook)
* [How to upload a large file to OSS](examples/upload_large_file)

## API docs

[API docs](https://godoc.org/github.com/Lupino/oss-go-sdk)
