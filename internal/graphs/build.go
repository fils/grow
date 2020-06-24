package graph

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/minio/minio-go"
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

	for message := range mc.ListObjectsV2(bucket, prefix, recursive, doneCh) {
		// fmt.Println(message.Key)
		k := strings.SplitAfterN(message.Key, "/", 3)
		// save for later  x := message.LastModified.UTC()

		if strings.HasSuffix(message, ".jsonld") {
			fmt.Println("Object needs processing")

			// convert object to RDF

			// if err == nil pipewrite to graph object
		}
	}
}
