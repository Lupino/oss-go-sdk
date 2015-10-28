# Getting start with oss-go-sdk

In this tutorial you will find how to start oss-go-sdk and get the bucket list from the server

## Create Account

Go to [OSS website](http://www.aliyun.com/product/oss/?lang=en), create a new account for new user.
After account created, you can create the OSS instance and get the `accessKeyId` and `accessKeySecret`.

## Install the sdk

Just use the simple `go get` to install the sdk

```bash
go get -v github.com/Lupino/oss-go-sdk
```

## Initial OSS API

Once you installed the sdk `import` it on you code.

```go
import (
    "github.com/Lupino/oss-go-sdk"
)
```

Get the default api options use `oss.GetDefaultAPIOptioins`,
then set the access key and secret from you OSS account.

```go
var APIOptions = oss.GetDefaultAPIOptioins()
APIOptions.AccessID = AccessKeyID
APIOptions.SecretAccessKey = AccessKeySecret
var OSSAPI = oss.NewAPI(APIOptions)
```

## Get Service

`oss.ListAllMyBucketsResult` is the result of `GetService`,
it also parse some argument for `GetService`,
the argument is `Prefix`, `Marker`, `MaxKeys`.

```go
var result oss.ListAllMyBucketsResult
var headers = make(map[string]string)
var err error
err = OSSAPI.GetService(&result, headers)
```

## Parse the error

`oss-go-sdk` implement error return by OSS into `error` interface by `oss.Error`,
so you can get error return by OSS server from `oss.Error`, just like:

```go
var realErr = err.(*oss.Error)
```

## The end

the source code [main.go](main.go)
