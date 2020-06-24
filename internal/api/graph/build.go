package graph

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/minio/minio-go"
	"github.com/piprate/json-gold/ld"
	"github.com/rs/xid"
)

// Build launches a go func to build a graph.
func Build(mc *minio.Client, bucket, prefix, domain string, w http.ResponseWriter, r *http.Request) {

	go builder(bucket, prefix, domain, mc)

	// TODO   // ? https://schema.org/CreateAction https://schema.org/docs/actions.html
	// POST a request JSON body to /api/sitemap
	// make PID
	// Make a JSON status document in /id/queue/[PID].json
	// send of go func (let it know the PID)
	// send 202 Accepted, location /id/queue/[PID] (any FET request to resource is 200 with info about job)
	// When go fun is done ned to make /id/queue/[PID]  303 with link to new resournce made,
	// rewrite the doc with a new status entry?

	//  use the presence of /id/queue/[PID]  to tell if this is a 200 report or 303 redirect
	// PID.json is the metadata for PID
	//  How to delete the queue (if to delete it) needs to be resolved.
}

func builder(bucket, prefix, domain string, mc *minio.Client) {

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
		for message := range mc.ListObjectsV2(bucket, prefix, recursive, doneCh) {
			// fmt.Println(message.Key)
			// k := strings.SplitAfterN(message.Key, "/", 3)
			// save for later  x := message.LastModified.UTC()

			if strings.HasSuffix(message.Key, ".jsonld") {
				fo, err := mc.GetObject(bucket, message.Key, minio.GetObjectOptions{})
				if err != nil {
					fmt.Printf("get object %s", err)
					// return "", err
				}

				var b bytes.Buffer
				bw := bufio.NewWriter(&b)

				_, err = io.Copy(bw, fo)
				if err != nil {
					fmt.Printf("iocopy %s", err)
				}

				// TODO
				// Process the bytes in b to RDF (with randomized blank nodes)

				// should be using gleaners
				proc := ld.NewJsonLdProcessor()
				options := ld.NewJsonLdOptions("")
				options.Format = "application/nquads"

				rdf, err := jld2nq(b.Bytes(), proc, options)
				if err != nil {
					fmt.Printf("jld2nq %s", err)
				}

				rdfubn := globalUniqueBNodes(rdf)

				pw.Write([]byte(rdfubn))

			}
		}
	}()

	go func() {
		defer lwg.Done()
		_, err := mc.PutObject("scratch", "requestgraph.nq", pr, -1, minio.PutObjectOptions{})
		if err != nil {
			log.Println(err)
		}
	}()

	lwg.Wait() // wait for the pipe read writes to finish
	pw.Close()
	pr.Close()

	fmt.Println("Builder call done")

}

// jld2nq converts JSON-LD documents to NQuads  (from Gleaner..  should leverage pkg in Gleaner for this)
func jld2nq(jsonld []byte, proc *ld.JsonLdProcessor, options *ld.JsonLdOptions) (string, error) {
	var myInterface interface{}
	err := json.Unmarshal(jsonld, &myInterface)
	if err != nil {
		return "", err
	}

	nq, err := proc.ToRDF(myInterface, options)
	if err != nil {
		return "", err
	}

	return nq.(string), err
}

//(from Gleaner..  should leverage pkg in Gleaner for this)
func globalUniqueBNodes(nq string) string {
	scanner := bufio.NewScanner(strings.NewReader(nq))
	// make a map here to hold our old to new map
	m := make(map[string]string)

	for scanner.Scan() {
		//fmt.Println(scanner.Text())
		// parse the line
		split := strings.Split(scanner.Text(), " ")
		sold := split[0]
		oold := split[2]

		if strings.HasPrefix(sold, "_:") { // we are a blank node
			// check map to see if we have this in our value already
			if _, ok := m[sold]; ok {
				// fmt.Printf("We had %s, already\n", sold)
			} else {
				guid := xid.New()
				snew := fmt.Sprintf("_:b%s", guid.String())
				m[sold] = snew
			}
		}

		// scan the object nodes too.. though we should find nothing here.. the above wouldn't
		// eventually find
		if strings.HasPrefix(oold, "_:") { // we are a blank node
			// check map to see if we have this in our value already
			if _, ok := m[oold]; ok {
				// fmt.Printf("We had %s, already\n", oold)
			} else {
				guid := xid.New()
				onew := fmt.Sprintf("_:b%s", guid.String())
				m[oold] = onew
			}
		}
		// triple := tripleBuilder(split[0], split[1], split[3])
		// fmt.Println(triple)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	//fmt.Println(m)

	filebytes := []byte(nq)

	for k, v := range m {
		// fmt.Printf("Replace %s with %v \n", k, v)
		filebytes = bytes.Replace(filebytes, []byte(k), []byte(v), -1)
	}

	return string(filebytes)
}
