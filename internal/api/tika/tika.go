package tika

import (
	"bufio"
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/bbalet/stopwords"
	"github.com/minio/minio-go"
)

// Build launches a go func to build.   Needs to return a
func Build(mc *minio.Client, bucket, prefix, domain string, w http.ResponseWriter, r *http.Request) {
	// go builder(bucket, prefix, domain, mc)
	go builder(bucket, prefix, domain, mc)
}

func builder(bucket, prefix, domain string, mc *minio.Client) {
	log.Printf("Bucket: %s  Prefix: %s   Domain: %s\n", bucket, prefix, domain)

	// Create a done channel.
	doneCh := make(chan struct{})
	defer close(doneCh)
	recursive := true

	// Pipecopy elements
	pr, pw := io.Pipe()     // TeeReader of use?
	lwg := sync.WaitGroup{} // work group for the pipe writes...
	lwg.Add(2)

	go func() {
		defer lwg.Done()
		defer pw.Close()

		// WARNING hard coded "prefix" here
		for message := range mc.ListObjectsV2(bucket, fmt.Sprintf("%s/csdco/do", prefix), recursive, doneCh) {

			if !strings.HasSuffix(message.Key, ".jsonld") {
				log.Println(message.Key)

				s, err := processObject(mc, bucket, prefix, message)
				if err != nil {
					log.Println(err)
				}

				pw.Write([]byte(s))
			}
		}

	}()

	go func() {
		defer lwg.Done()
		var op string
		if prefix == "" {
			op = fmt.Sprint("website/fulltext.nq") // should this be website?
		} else {
			op = fmt.Sprintf("%s/website/fulltext.nq", prefix)
		}

		log.Println(op)

		_, err := mc.PutObject(bucket, op, pr, -1, minio.PutObjectOptions{}) // TODO  this is potentially dangerous..  it will over write this object at least
		if err != nil {
			log.Println(err)
		}
	}()

	lwg.Wait() // wait for the pipe read writes to finish
	pw.Close()
	pr.Close()

	log.Println("Builder call done")

}

// SingleBuild is a test function to call a single item
func SingleBuild(mc *minio.Client, bucket, prefix, domain string, w http.ResponseWriter, r *http.Request) {
	go singlebuilder(bucket, prefix, domain, mc)
}

// TESTING function
func singlebuilder(bucket, prefix, domain string, mc *minio.Client) {

	// Pipecopy elements
	pr, pw := io.Pipe()     // TeeReader of use?
	lwg := sync.WaitGroup{} // work group for the pipe writes...
	lwg.Add(2)

	go func() {
		defer lwg.Done()
		defer pw.Close()

		bucket = "ocdprod"
		object := "/csdco/do/000003a5ee30630237ae9690fd10a576b8bfb3d6c3e2ce541924522ef5b69f2c"

		message, err := mc.StatObject(bucket, object, minio.StatObjectOptions{})
		if err != nil {
			log.Print(err)
		}

		s, err := processObject(mc, bucket, prefix, message)
		if err != nil {
			log.Println(err)
		}

		pw.Write([]byte(s))

	}()

	go func() {
		defer lwg.Done()
		var op string
		if prefix == "" {
			op = fmt.Sprint("website/fulltext.nq", prefix) // should this be website?
		} else {
			op = fmt.Sprintf("%s/website/fulltext.nq", prefix)
		}

		_, err := mc.PutObject(bucket, op, pr, -1, minio.PutObjectOptions{}) // TODO  this is potentially dangerous..  it will over write this object at least
		if err != nil {
			log.Println(err)
		}
	}()

	lwg.Wait() // wait for the pipe read writes to finish
	pw.Close()
	pr.Close()

	log.Println("Builder call done")

}

func processObject(mc *minio.Client, bucket, prefix string, message minio.ObjectInfo) (string, error) {
	fo, err := mc.GetObject(bucket, message.Key, minio.GetObjectOptions{})
	if err != nil {
		log.Printf("get object %s", err)
		// return "", err
	}

	var b bytes.Buffer
	bw := bufio.NewWriter(&b)

	_, err = io.Copy(bw, fo)
	if err != nil {
		log.Printf("iocopy %s", err)
	}

	s, err := EngineTika(b.Bytes())
	t, err := fullTextTrpls(s, message.Key)

	if err != nil {
		log.Println(err)
	}

	return t, err
}

func fullTextTrpls(s, obj string) (string, error) {

	// log.Println(obj)

	t := fmt.Sprintf("<https://opencoredata.org/id%s>  <https://schema.org/text> \"%s\" ", obj, s)
	// https://opencoredata.org/id/csdco/do/22ee510dde701eda9df86022561d5019fe259768604824b653b2df8fae8055bd
	//                             prefix
	return t, nil
}

// EngineTika sends a byte array to tika for processing into text
func EngineTika(b []byte) (string, error) {
	tikaurl := "http://tika:9998/tika"

	req, err := http.NewRequest("PUT", tikaurl, bytes.NewReader(b))
	req.Header.Set("Accept", "text/plain")
	req.Header.Set("User-Agent", "EarthCube_DataBot/1.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	// fmt.Println("Tika Response Status:", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	sw := stopwords.CleanString(string(body), "en", true) // remove stop words..   no reason for them in the search
	// remove anything not text and numbers
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Println(err)
	}
	processedString := reg.ReplaceAllString(sw, " ")

	// TODO remove duplicate words.. DONE..   but needs review
	return dedup(processedString), err
}

func dedup(input string) string {
	unique := []string{}

	words := strings.Split(input, " ")
	for _, word := range words {
		// If we alredy have this word, skip.
		if contains(unique, word) {
			continue
		}

		unique = append(unique, word)
	}

	return strings.Join(unique, " ")
}

func contains(strs []string, str string) bool {
	for _, s := range strs {
		if s == str {
			return true
		}
	}
	return false
}
