package fileobjects

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"oceanleadership.org/grow/internal/fileactions"

	"github.com/minio/minio-go"
)

// FileObjects pulls the objects from the object store
func FileObjects(mc *minio.Client, bucket, prefix string, w http.ResponseWriter, r *http.Request) {
	var object string
	key := fmt.Sprintf("%s", r.URL.Path)

	// TODO review this hack...
	// the browser rewrites /index.html to / and this of course fails to locate the object.
	// So for the one case of index.html (.htm ?) we need to do this hack
	if key == "" { // deal with browser index.html rewrites and sub directories!!!
		key = "index.html"
	}

	m := fileactions.MimeByType(filepath.Ext(key))
	w.Header().Set("Content-Type", m)
	// log.Printf("%s: %s \n", key, m)

	object = fmt.Sprintf("%s/website/%s", prefix, key)

	// log.Println(object)

	// check our object is there first....
	_, err := mc.StatObject(bucket, object, minio.StatObjectOptions{})
	if err != nil {
		log.Println("Error on object access")
		log.Println(err)
		http.Error(w, "object not found", 404)
		return
	}

	fo, err := mc.GetObject(bucket, object, minio.GetObjectOptions{})
	if err != nil {
		log.Printf("Error getobjt: %s  \n err: %s", object, err)
		//http.Error(w, "object not found",404)
		return
		// if object is .hml then 404 on this.
	}

	_, err = io.Copy(w, fo) // todo need to stream write from s3 reader...  not copy
	// log.Println(n)
	if err != nil {
		log.Println("Issue with writing file to http response")
		log.Println(err)
		// Is this an internal server error
	}
}
