oss-go-sdk
==================================

This repo is under development.

[![Build Status](https://travis-ci.org/Lupino/oss-go-api.svg?branch=master)](https://travis-ci.org/Lupino/oss-go-api)
[![Coveralls](https://coveralls.io/repos/Lupino/oss-go-api/badge.png?branch=master)](https://coveralls.io/r/Lupino/oss-go-api)

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

```golang
import (
    "github.com/Lupino/oss-go-sdk"
)
var APIOptions = oss.GetDefaultAPIOptioins()
APIOptions.AccessID = AccessKeyID
APIOptions.SecretAccessKey = AccessKeySecret
var OSSAPI = oss.NewAPI(APIOptions)
```

## Get Service

```golang
var result oss.ListAllMyBucketsResult
var headers = make(map[string]string)
var err error
err = OSSAPI.GetService(&result, headers)
```

## Parse the error

all the error is the error xml so you shoud parse the error and see the error message.

```golang
var e = oss.ParseError(err)
```

## API docs

[API docs](https://godoc.org/github.com/Lupino/oss-go-api)
