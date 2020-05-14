package ufoknorg

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/minio/minio-go"
	"oceanleadership.org/grow/pkg/objservices/spatial"
)

// UFOKNPageData is the struct for the template page
type UFOKNPageData struct {
	JSONLD  string
	GeoJSON string
}

// TODO
// Need a generic /id approach here...

// DO pulls the objects from the object store
func DO(mc *minio.Client, bucket, prefix string, w http.ResponseWriter, r *http.Request) {
	// First step, see what we are asked for
	// by the extension and its mime.
	// that will set the object search.
	// 1) it's in the object store and is sent
	// 2) it's not in the object store and we have a render option
	// 3) we have no way to deal with that format
	// Really want this to be a "route" ..   check the open core code to see if I did that there....

	acpt := r.Header.Get("Accept")
	// ref: https://play.golang.org/p/S7xCsiKe8KE
	objpath := r.URL.Path
	objname := path.Base(r.URL.Path)
	objext := path.Ext(r.URL.Path)

	log.Printf("objpath: %s \nobjectname: %s \nobjext: %s\n", objpath, objname, objext)

	key := fmt.Sprintf("%s.jsonld", r.URL.Path) // the default for now is to look for .jsonld on this resource ID
	object := fmt.Sprintf("%s/%s", prefix, key)

	// TODO
	// stat the object like I did..
	// if found..  send with mimetype no templating
	// if not found look for registered functions
	//            if not found, format not supported
	//            if found..  process and send

	fo, err := mc.GetObject(bucket, object, minio.GetObjectOptions{})
	if err != nil {
		log.Println(err)
		// TODO need to return an error here...
	}

	if strings.Contains(acpt, "text/html") {
		fmt.Println("Client can understand html")
		w.Header().Set("Content-Type", "text/html")

		var b bytes.Buffer
		bw := bufio.NewWriter(&b)

		_, err = io.Copy(bw, fo)
		if err != nil {
			log.Println(err)
		}

		// Get the template from the site assets
		// t := "assets/ufokndam.html" // TODO ..   can I get the template from the object store too?

		t := fmt.Sprintf("assets/templates/%s/template.html", filepath.Dir(r.URL.Path))
		// Read the template into a text var and parse that.

		fmt.Printf("b %s  p  %s   t %s\n", bucket, prefix, t)

		to, err := mc.GetObject(bucket, fmt.Sprintf("%s/%s", prefix, t), minio.GetObjectOptions{})
		if err != nil {
			log.Println(err)
			// TODO need to return an error here...
		}

		var tb bytes.Buffer
		tbw := bufio.NewWriter(&tb)

		_, err = io.Copy(tbw, to)
		if err != nil {
			log.Println(err)
		}

		tc := string(tb.Bytes())

		// tc is our JSON-LD.  We not want to perform an object service on it.
		gj, err := spatial.SDO2GeoJSON(string(b.Bytes()))
		if err != nil {
			log.Println(err)
		}

		pd := UFOKNPageData{JSONLD: string(b.Bytes()), GeoJSON: gj}

		//ht, err := template.New("object template").ParseFiles(t) // open and parse a template text file
		ht, err := template.New("object template").Parse(tc) // open and parse a template text file
		if err != nil {
			log.Printf("template parse failed: %s", err)
		}

		err = ht.ExecuteTemplate(w, "T", pd) // substitute fields in the template 't', with values from 'user' and write it out to 'w' which implements io.Writer
		if err != nil {
			log.Printf("htemplate execution failed: %s", err)
		}

	} else {
		fmt.Println("Client says it will take what I give it..  good luck with that buddy...")

		// m := fileactions.MimeByType(filepath.Ext(key))

		w.Header().Set("Content-Type", "application/ld+json") //  m)  // override for now until I update the loaded for samples earth to mod the object name with .jsonld
		n, err := io.Copy(w, fo)                              // todo need to stream write from s3 reader...  not copy
		log.Println(n)
		if err != nil {
			log.Println("Issue with writing file to http response")
			log.Println(err)
		}
	}

}
