package oss

import (
	"fmt"
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
	return mux
}

func mockHTTPServer() *httptest.Server {
	return httptest.NewServer(handle())
}
