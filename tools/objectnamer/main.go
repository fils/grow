package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"time"

	"github.com/minio/minio-go"
	"github.com/tidwall/gjson"
)

var urlprefix, s3addressVal, s3bucketVal, s3prefixVal, keyVal, secretVal string

func init() {
	log.SetFlags(log.Lshortfile)

	flag.StringVar(&s3addressVal, "server", "0.0.0.0:0000", "Address of the object server with port")
	flag.StringVar(&s3bucketVal, "bucket", "website", "bucket which holds the web site objects")
	flag.StringVar(&s3prefixVal, "prefix", "website", "bucket prefix for the objects")
	flag.StringVar(&keyVal, "key", "config", "Object server key")
	flag.StringVar(&secretVal, "secret", "config", "Object server secret")
}

func main() {
	fmt.Println("Remame")

	// parse environment vars
	s3addressVal = os.Getenv("S3ADDRESS")
	s3bucketVal = os.Getenv("S3BUCKET")
	s3prefixVal = os.Getenv("S3PREFIX")
	keyVal = os.Getenv("S3KEY")
	secretVal = os.Getenv("S3SECRET")

	flag.Parse() // parse any command line flags...
	log.Printf("uL %s a: %s  b %s  p %s  k %s  s %s\n", urlprefix, s3addressVal, s3bucketVal, s3prefixVal, keyVal, secretVal)

	start := time.Now()
	inspectRename(urlprefix, s3addressVal, s3bucketVal, s3prefixVal, keyVal, secretVal)
	//objectRename(urlprefix, s3addressVal, s3bucketVal, s3prefixVal, keyVal, secretVal)
	elapsed := time.Since(start)
	log.Printf("Run elaspsed time: %s", elapsed)
}

func inspectRename(prefix, server, bucket, objPrefix, key, secret string) {
	mc := minioConnection(server, "32768", key, secret)

	// Create a done channel.
	doneCh := make(chan struct{})
	defer close(doneCh)
	recursive := true

	for message := range mc.ListObjectsV2(bucket, objPrefix, recursive, doneCh) {

		// read the object
		object, err := mc.GetObject(bucket, message.Key, minio.GetObjectOptions{})
		if err != nil {
			fmt.Println(err)
			return
		}

		var b bytes.Buffer
		foo := bufio.NewWriter(&b)

		stat, err := object.Stat()
		if err != nil {
			log.Fatalln(err)
		}

		if _, err := io.CopyN(foo, object, stat.Size); err != nil {
			log.Fatalln(err)
		}

		j := string(b.Bytes())

		// prse the json to get the value I want
		value := gjson.Get(j, `0.@graph.0.http://schema\.org/value.0.@value`)
		// value := gjson.Get(j, `distribution.contentUrl`)
		println(value.String())

		/*
			u, err := url.Parse(value.String())
			if err != nil {
				log.Println(err)
			}
			p := strings.Split(u.Path, "/")
			//fmt.Println(p[3])
		*/

		// copy the object with a new name
		src := minio.NewSourceInfo(bucket, message.Key, nil)

		//n := fmt.Sprintf("%s.jsonld", p[3])
		n := fmt.Sprintf("%s.jsonld", value.String())
		// Destination object
		dst, err := minio.NewDestinationInfo("dometa", n, nil, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Copy object call
		err = mc.CopyObject(dst, src)
		if err != nil {
			fmt.Println(err)
			return
		}

	}

}

func objectRename(prefix, server, bucket, objPrefix, key, secret string) {
	mc := minioConnection(server, "32768", key, secret)

	// Create a done channel.
	doneCh := make(chan struct{})
	defer close(doneCh)
	recursive := true

	for message := range mc.ListObjectsV2(bucket, objPrefix, recursive, doneCh) {

		src := minio.NewSourceInfo(bucket, message.Key, nil)

		n := fmt.Sprintf("%s.jsonld", message.Key)
		// Destination object
		dst, err := minio.NewDestinationInfo("pkgout", n, nil, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Copy object call
		err = mc.CopyObject(dst, src)
		if err != nil {
			fmt.Println(err)
			return
		}

	}

}

func minioConnection(minioVal, portVal, accessVal, secretVal string) *minio.Client {
	endpoint := fmt.Sprintf("%s:%s", minioVal, portVal)
	accessKeyID := accessVal
	secretAccessKey := secretVal
	useSSL := false
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Println(err)
	}
	return minioClient
}
