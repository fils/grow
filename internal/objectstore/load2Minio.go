package objectstore

import (
	"bytes"
	"log"

	minio "github.com/minio/minio-go"
)

// LoadToMinio loads jsonld into the specified bucket
func LoadToMinio(b []byte, bucketName, objectName string, mc *minio.Client) (int64, error) {

	// set up some elements for PutObject
	// TODO, get content based on objectName extension
	contentType := "application/xml"
	bb := bytes.NewBuffer(b)
	usermeta := make(map[string]string) // what do I want to know?
	// usermeta["url"] = urlloc
	// usermeta["sha1"] = bss

	//log.Println(bucketName)
	n, err := mc.PutObject(bucketName, objectName, bb, int64(bb.Len()), minio.PutObjectOptions{ContentType: contentType, UserMetadata: usermeta})
	if err != nil {
		log.Printf("%s/%s", bucketName, objectName)
		log.Println(err)
		// TODO   should return 0, err here and deal with it on the other end
	}

	// log.Printf("#%d Uploaded Bucket:%s File:%s Size %d\n", i, bucketName, objectName, n)

	return n, nil
}
