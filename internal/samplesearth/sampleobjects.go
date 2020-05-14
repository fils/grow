package samplesearth

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"text/template"

	"github.com/minio/minio-go"
	"oceanleadership.org/grow/pkg/objservices/spatial"
)

// PageData is the struct for the template page
type PageData struct {
	JSONLD  string
	GeoJSON string
}

// IGSNDataGraph pulls the objects from the object store
func IGSNDataGraphOLD(mc *minio.Client, bucket, prefix string, w http.ResponseWriter, r *http.Request) {
	// TODO review this hack...
	// the browser rewrites /index.html to / and this of course fails to locate the object.
	// So for the one case of index.html (.htm ?) we need to do this hack
	key := fmt.Sprintf("%s", r.URL.Path)
	// if key == "" { // deal with browser index.html rewrites
	// 	key = "index.html"
	// }

	log.Println(key)

	fo, err := mc.GetObject("doclouds", "igsndatagraphs/"+key, minio.GetObjectOptions{})
	if err != nil {
		log.Println(err)
		// TODO need to return an error here...
	}

	acpt := r.Header.Get("Accept")

	if strings.Contains(acpt, "text/html") {
		fmt.Println("Client can understand html")
		w.Header().Set("Content-Type", "text/html")

		var b bytes.Buffer
		bw := bufio.NewWriter(&b)

		_, err = io.Copy(bw, fo)
		if err != nil {
			log.Println(err)
		}

		// tc is our JSON-LD.  We not want to perform an object service on it.
		gj, err := spatial.SDO2GeoJSON(string(b.Bytes()))
		if err != nil {
			log.Println(err)
		}

		pd := PageData{JSONLD: string(b.Bytes()), GeoJSON: gj}

		t := "assets/templates/id/template.html" // TODO ..   can I get the template from the object store too?
		// t := fmt.Sprintf("assets/templates/%s/template.html", filepath.Dir(r.URL.Path))

		ht, err := template.New("object template").ParseFiles(t) // open and parse a template text file
		if err != nil {
			log.Printf("template parse failed: %s", err)
		}

		err = ht.ExecuteTemplate(w, "T", pd) // substitute fields in the template 't', with values from 'user' and write it out to 'w' which implements io.Writer
		if err != nil {
			log.Printf("htemplate execution failed: %s", err)
		}

	} else {
		fmt.Println("Client says it will take what I give it")

		// m := fileactions.MimeByType(filepath.Ext(key))

		w.Header().Set("Content-Type", "application/ld+json") //  m)  // override for now until I update the loaded for samples earth to mod the object name with .jsonld
		n, err := io.Copy(w, fo)                              // todo need to stream write from s3 reader...  not copy
		log.Println(n)
		if err != nil {
			log.Println("Issue with writing file to http response")
			log.Println(r.URL.Path)
			log.Println(err)
		}
		// log.Printf("%s: %s \n", key, m)
	}

}
