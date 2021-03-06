package main

import (
	"github.com/Lupino/oss-go-sdk"
	"io"
	"log"
	"os"
)

// AccessKeyID defined Access Key ID
const AccessKeyID = "3cpfsfx4f1yqxb3zy9rx1vn4"

// AccessKeySecret defined Access Key Secret
const AccessKeySecret = "T4Re+CRM4oXrqvoqhv0vNRZi2sQ="

func main() {
	var APIOptions = oss.GetDefaultAPIOptioins()
	APIOptions.AccessID = AccessKeyID
	APIOptions.SecretAccessKey = AccessKeySecret
	var OSSAPI, err = oss.NewAPI(APIOptions)

	if err != nil {
		log.Fatal(err)
	}

	var bucket = "ossgosdklargefile"
	var object = "largefile.bin"
	var bufSize = int64(1024 * 1024 * 2) // 2 M
	var file = "test.bin"
	var fp *os.File

	if err = OSSAPI.PutBucket(bucket, oss.ACLPublicReadWrite, "", nil); err != nil {
		log.Printf("%s\n", err)
	}

	var multi *oss.MultipartUpload
	if multi, err = OSSAPI.NewMultipartUpload(bucket, object, nil); err != nil {
		log.Printf("%s\n", err)
	}

	if fp, err = os.Open(file); err != nil {
		log.Fatal(err)
	}

	var stat, _ = fp.Stat()
	var fileLength = stat.Size()
	var filePart = int(fileLength / bufSize)
	if int64(filePart)*bufSize < fileLength {
		filePart = filePart + 1
	}

	var rd io.Reader
	var parts = make([]oss.Part, filePart)
	var etag string

	for i := 1; i <= filePart; i++ {
		rd = io.LimitReader(fp, bufSize)
		if etag, err = multi.UploadPart(i, rd); err != nil {
			log.Printf("%s\n", err)
		}
		log.Printf("PartNumber: %d, ETag: %s\n", i, etag)
		parts[i-1] = oss.Part{
			PartNumber: i,
			ETag:       etag,
		}
	}

	var result oss.CompleteMultipartUploadResult

	if err = multi.CompleteUpload(parts, &result); err != nil {
		log.Printf("%s\n", err)
	}

	log.Printf("Upload result: %s\n", result)
}
