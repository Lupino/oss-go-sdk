package oss

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

func handle() *http.ServeMux {
	var mux = http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, `
<?xml version="1.0" encoding="UTF-8"?>
<ListAllMyBucketsResult>
  <Prefix>xz02tphky6fjfiuc</Prefix>
  <Marker></Marker>
  <MaxKeys>1</MaxKeys>
  <IsTruncated>true</IsTruncated>
  <NextMarker>xz02tphky6fjfiuc0</NextMarker>
  <Owner>
    <ID>ut_test_put_bucket</ID>
    <DisplayName>ut_test_put_bucket</DisplayName>
  </Owner>
  <Buckets>
    <Bucket>
      <Location>oss-cn-hangzhou-a</Location>
      <Name>xz02tphky6fjfiuc0</Name>
      <CreationDate>2014-05-15T11:18:32.000Z</CreationDate>
    </Bucket>
  </Buckets>
</ListAllMyBucketsResult>
        `)
	})
	mux.HandleFunc("/bucket/", func(w http.ResponseWriter, req *http.Request) {
		var query = req.URL.Query()
		var method = req.Method
		if method == "POST" {
			if _, ok := query["delete"]; ok {
				fmt.Fprintf(w, `
<?xml version="1.0" encoding="UTF-8"?>
<DeleteResult xmlns="http://doc.oss-cn-hangzhou.aliyuncs.com">
    <Deleted>
       <Key>multipart.data</Key>
    </Deleted>
    <Deleted>
       <Key>test.jpg</Key>
    </Deleted>
    <Deleted>
       <Key>demo.jpg</Key>
    </Deleted>
</DeleteResult>
                `)
				return
			}
		}
		if method != "GET" {
			fmt.Fprintf(w, "success")
			return
		}
		if _, ok := query["acl"]; ok {
			fmt.Fprintf(w, `
            <?xml version="1.0" ?>
<AccessControlPolicy>
    <Owner>
        <ID>00220120222</ID>
        <DisplayName>user_example</DisplayName>
    </Owner>
    <AccessControlList>
        <Grant>public-read</Grant>
    </AccessControlList>
</AccessControlPolicy>
            `)
			return
		}

		if _, ok := query["location"]; ok {
			fmt.Fprintf(w, `
            <?xml version="1.0" encoding="UTF-8"?>
<LocationConstraint xmlns="http://doc.oss-cn-hangzhou.aliyuncs.com">oss-cn-hangzhou</LocationConstraint >
            `)
			return
		}

		if _, ok := query["logging"]; ok {
			fmt.Fprintf(w, `
            <?xml version="1.0" encoding="UTF-8"?>
<BucketLoggingStatus xmlns="http://doc.oss-cn-hangzhou.aliyuncs.com">
<LoggingEnabled>
<TargetBucket>mybucketlogs</TargetBucket>
<TargetPrefix>mybucket-access_log/</TargetPrefix>
</LoggingEnabled>
</BucketLoggingStatus>
            `)
			return
		}

		if _, ok := query["website"]; ok {
			fmt.Fprintf(w, `
<?xml version="1.0" encoding="UTF-8"?>
<WebsiteConfiguration xmlns="http://doc.oss-cn-hangzhou.aliyuncs.com">
<IndexDocument>
<Suffix>index.html</Suffix>
</IndexDocument>
<ErrorDocument>
<Key>error.html</Key>
</ErrorDocument>
</WebsiteConfiguration>
            `)
			return
		}

		if _, ok := query["referer"]; ok {
			fmt.Fprintf(w, `
            <?xml version="1.0" encoding="UTF-8"?>
<RefererConfiguration>
<AllowEmptyReferer>true</AllowEmptyReferer >
<RefererList>
<Referer> http://www.aliyun.com</Referer>
<Referer> https://www.aliyun.com</Referer>
<Referer> http://www.*.com</Referer>
<Referer> https://www.?.aliyuncs.com</Referer>
</RefererList>
</RefererConfiguration>
            `)
			return
		}

		if _, ok := query["lifecycle"]; ok {
			fmt.Fprintf(w, `
            <?xml version="1.0" encoding="UTF-8"?>
<LifecycleConfiguration>
  <Rule>
    <ID>delete after one day</ID>
    <Prefix>logs/</Prefix>
    <Status>Enabled</Status>
    <Expiration>
      <Days>1</Days>
    </Expiration>
  </Rule>
</LifecycleConfiguration>
            `)
			return
		}

		if _, ok := query["uploads"]; ok {
			fmt.Fprintf(w, `
            <?xml version="1.0" encoding="UTF-8"?>
<ListMultipartUploadsResult xmlns="http://doc.oss-cn-hangzhou.aliyuncs.com">
    <Bucket>oss-example</Bucket>
    <KeyMarker></KeyMarker>
    <UploadIdMarker></UploadIdMarker>
    <NextKeyMarker>oss.avi</NextKeyMarker>
    <NextUploadIdMarker>0004B99B8E707874FC2D692FA5D77D3F</NextUploadIdMarker>
    <Delimiter></Delimiter>
    <Prefix></Prefix>
    <MaxUploads>1000</MaxUploads>
    <IsTruncated>false</IsTruncated>
    <Upload>
        <Key>multipart.data</Key>
        <UploadId>0004B999EF518A1FE585B0C9360DC4C8</UploadId>
        <Initiated>2012-02-23T04:18:23.000Z</Initiated>
    </Upload>
    <Upload>
        <Key>multipart.data</Key>
        <UploadId>0004B999EF5A239BB9138C6227D69F95</UploadId>
        <Initiated>2012-02-23T04:18:23.000Z</Initiated>
    </Upload>
    <Upload>
        <Key>oss.avi</Key>
        <UploadId>0004B99B8E707874FC2D692FA5D77D3F</UploadId>
        <Initiated>2012-02-23T06:14:27.000Z</Initiated>
    </Upload>
</ListMultipartUploadsResult>
            `)
			return
		}

		fmt.Fprintf(w, `
<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult xmlns="http://doc.oss-cn-hangzhou.aliyuncs.com">
<Name>bucket</Name>
<Prefix>fun</Prefix>
<Marker></Marker>
<MaxKeys>100</MaxKeys>
<Delimiter></Delimiter>
    <IsTruncated>false</IsTruncated>
    <Contents>
        <Key>fun/movie/001.avi</Key>
        <LastModified>2012-02-24T08:43:07.000Z</LastModified>
        <ETag>&quot;5B3C1A2E053D763E1B002CC607C5A0FE&quot;</ETag>
        <Type>Normal</Type>
        <Size>344606</Size>
        <StorageClass>Standard</StorageClass>
        <Owner>
            <ID>00220120222</ID>
            <DisplayName>user_example</DisplayName>
        </Owner>
    </Contents>
    <Contents>
        <Key>fun/movie/007.avi</Key>
        <LastModified>2012-02-24T08:43:27.000Z</LastModified>
        <ETag>&quot;5B3C1A2E053D763E1B002CC607C5A0FE&quot;</ETag>
        <Type>Normal</Type>
        <Size>344606</Size>
        <StorageClass>Standard</StorageClass>
        <Owner>
            <ID>00220120222</ID>
            <DisplayName>user_example</DisplayName>
        </Owner>
    </Contents>
    <Contents>
        <Key>fun/test.jpg</Key>
        <LastModified>2012-02-24T08:42:32.000Z</LastModified>
        <ETag>&quot;5B3C1A2E053D763E1B002CC607C5A0FE&quot;</ETag>
        <Type>Normal</Type>
        <Size>344606</Size>
        <StorageClass>Standard</StorageClass>
        <Owner>
            <ID>00220120222</ID>
            <DisplayName>user_example</DisplayName>
        </Owner>
    </Contents>
</ListBucketResult>
        `)
	})
	mux.HandleFunc("/bucket/object", func(w http.ResponseWriter, req *http.Request) {
		var query = req.URL.Query()
		var method = req.Method
		if method == "GET" {
			if _, ok := query["acl"]; ok {
				fmt.Fprintf(w, `
<?xml version="1.0" ?>
<AccessControlPolicy>
    <Owner>
        <ID>00220120222</ID>
        <DisplayName>00220120222</DisplayName>
    </Owner>
    <AccessControlList>
        <Grant>public-read </Grant>
    </AccessControlList>
</AccessControlPolicy>
                `)
				return
			}
			if _, ok := query["uploadId"]; ok {
				fmt.Fprintf(w, `
                <?xml version="1.0" encoding="UTF-8"?>
<ListPartsResult xmlns="http://doc.oss-cn-hangzhou.aliyuncs.com">
    <Bucket>multipart_upload</Bucket>
    <Key>multipart.data</Key>
    <UploadId>0004B999EF5A239BB9138C6227D69F95</UploadId>
    <NextPartNumberMarker>5</NextPartNumberMarker>
    <MaxParts>1000</MaxParts>
    <IsTruncated>false</IsTruncated>
    <Part>
        <PartNumber>1</PartNumber>
        <LastModified>2012-02-23T07:01:34.000Z</LastModified>
        <ETag>&quot;3349DC700140D7F86A078484278075A9&quot;</ETag>
        <Size>6291456</Size>
    </Part>
    <Part>
        <PartNumber>2</PartNumber>
        <LastModified>2012-02-23T07:01:12.000Z</LastModified>
        <ETag>&quot;3349DC700140D7F86A078484278075A9&quot;</ETag>
        <Size>6291456</Size>
    </Part>
    <Part>
        <PartNumber>5</PartNumber>
        <LastModified>2012-02-23T07:02:03.000Z</LastModified>
        <ETag>&quot;7265F4D211B56873A381D321F586E4A9&quot;</ETag>
        <Size>1024</Size>
    </Part>
</ListPartsResult>
                `)
				return
			}
			fmt.Fprintf(w, `this is the object body`)
			return
		}
		if method == "PUT" {
			var headers = req.Header
			if source, ok := headers["X-Oss-Copy-Source"]; ok {
				fmt.Printf("x-oss-copy-source: %s\n", source)
				if _, ok := query["uploadId"]; ok {
					fmt.Fprintf(w, `
                    <?xml version="1.0" encoding="UTF-8"?>
<CopyPartResult xmlns="http://doc.oss-cn-hangzhou.aliyuncs.com">
    <LastModified>2014-07-17T06:27:54.000Z </LastModified>
    <ETag>"5B3C1A2E053D763E1B002CC607C5A0FE"</ETag>
</CopyPartResult>
                    `)
					return
				}
				fmt.Fprintf(w, `
                <?xml version="1.0" encoding="UTF-8"?>
<CopyObjectResult xmlns="http://doc.oss-cn-hangzhou.aliyuncs.com">
    <LastModified>2014-05-15T11:18:32.000Z</LastModified>
    <ETag>"5B3C1A2E053D763E1B002CC607C5A0FE"</ETag>
</CopyObjectResult>
                `)
				return
			}
			var buf, _ = ioutil.ReadAll(req.Body)
			fmt.Printf("object length is: %d\n", len(buf))
		}
		if method == "POST" {
			if _, ok := query["uploadId"]; ok {
				fmt.Fprintf(w, `
                <?xml version="1.0" encoding="UTF-8"?>
<CompleteMultipartUploadResult xmlns="http://doc.oss-cn-hangzhou.aliyuncs.com">
    <Location>http://oss-example.oss-cn-hangzhou.aliyuncs.com /multipart.data</Location>
    <Bucket>oss-example</Bucket>
    <Key>multipart.data</Key>
    <ETag>B864DB6A936D376F9F8D3ED3BBE540DD-3</ETag>
</CompleteMultipartUploadResult>
                `)
				return
			}
			fmt.Fprintf(w, `
<?xml version="1.0" encoding="UTF-8"?>
<InitiateMultipartUploadResult xmlns="http://doc.oss-cn-hangzhou.aliyuncs.com">
    <Bucket>bucket</Bucket>
    <Key>object</Key>
    <UploadId>0004B9894A22E5B1888A1E29F8236E2D</UploadId>
</InitiateMultipartUploadResult>
            `)
			if _, ok := query["uploads"]; ok {
				return
			}
		}
		fmt.Fprintf(w, "success")
	})
	return mux
}

func mockHTTPServer() *httptest.Server {
	return httptest.NewServer(handle())
}
