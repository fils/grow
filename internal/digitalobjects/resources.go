package digitalobjects

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"

	minio "github.com/minio/minio-go/v7"

	"github.com/fils/goobjectweb/internal/fileactions"
	"github.com/fils/goobjectweb/internal/operations"
)

// UFOKNPageData is the struct for the template page
type UFOKNPageData struct {
	JSONLD  string
	PID     string
	GeoJSON string
}

// DO pulls the objects from the object store
func DO(mc *minio.Client, bucket, prefix, domain string, w http.ResponseWriter, r *http.Request) {

	acptHTML := strings.Contains(r.Header.Get("Accept"), "text/html")

	// TODO need to look for accept header and is appliction/ld+json then send the JSON-LD
	//var contentTypeHeader = r.Header["Content-Type"]
	//if contains(contentTypeHeader, "application/ld+json") || contains(contentTypeHeader, "application/json") || fileExtensionIsJson(urlloc) {

	// TODO add in the elseif here to check for .zip (in both sections)
	// then route as I do for geojson  (look to make this generic at this time?)

	if acptHTML {
		ext := filepath.Ext(r.URL.Path)
		if ext == "" || ext == ".jsonld" || ext == ".html" {
			s := strings.TrimSuffix(r.URL.Path, ext)

			// TODO ocd update March 21  CSDCO
			// ocdprod o: website/csdco/do/bf9f63664bc66179ba65f8990ab1dd9b6e804c25dbc80cd547c3a786eac6d4fe.jsonld p:website "
			fmt.Println(prefix)
			newprefix := strings.Replace(prefix, "website", "", 1)
			fmt.Println(newprefix)
			prefix = newprefix

			object := fmt.Sprintf("%s/%s.jsonld", prefix, s) // if prefix is nill?
			log.Printf("b: %s o: %s p:%s ", bucket, object, prefix)
			err := sendHTML(mc, w, r, bucket, object, prefix)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)
			}
		} else if ext == ".zip" {
			s := strings.TrimSuffix(r.URL.Path, ext)
			object := fmt.Sprintf("%s/%s.zip", prefix, s) // if prefix is nill?
			log.Printf("b: %s o: %s p:%s ", bucket, object, prefix)
			operations.DownloadPkg(mc, w, r, bucket, object)
		} else {
			log.Printf("Unsupported media type request, in the future I will check function map\n")
			http.Error(w, http.StatusText(http.StatusUnsupportedMediaType),
				http.StatusUnsupportedMediaType)
		}
	} else {
		// We are not HTML at this point, so we might be sending an object
		// or attempting to render a representation of one.
		// 1) check if the object exist and send it.
		// 2) if does not exist, check if the ext matches a render version

		// TODO ocd update March 21  CSDCO
		// ocdprod o: website/csdco/do/bf9f63664bc66179ba65f8990ab1dd9b6e804c25dbc80cd547c3a786eac6d4fe.jsonld p:website "
		fmt.Println(prefix)
		newprefix := strings.Replace(prefix, "website", "", 1)
		fmt.Println(newprefix)
		prefix = newprefix

		//object := fmt.Sprintf("%s/%s", prefix, r.URL.Path)
		object := fmt.Sprintf("%s", r.URL.Path)

		_, err := mc.StatObject(context.Background(), bucket, object, minio.StatObjectOptions{})
		// fmt.Println(objInfo)

		if err != nil {
			// we don't see this object by the provided object name, so
			// let's see if the extension/mimetype can be rendered
			ext := filepath.Ext(object)

			// TODO recode the if else if to a switch statement?
			// 	switch ext {
			// 	case ".geojson":
			// 		err := operations.TypeGeoJSON(mc, w, r, bucket, object)
			// 		if err != nil {
			// 			log.Println(err)
			// 			http.Error(w, http.StatusText(http.StatusNotFound),
			// 				http.StatusNotFound)
			// 		}
			// case "":

			//log.Printf("Extension: %s", ext)

			if strings.Contains(ext, ".geojson") {
				err := operations.TypeGeoJSON(mc, w, r, bucket, object)
				if err != nil {
					log.Println(err)
					http.Error(w, http.StatusText(http.StatusNotFound),
						http.StatusNotFound)
				}
			} else if strings.Contains(ext, ".zip") {
				err := operations.DownloadPkg(mc, w, r, bucket, object)
				if err != nil {
					log.Println(err)
					http.Error(w, http.StatusText(http.StatusNotFound),
						http.StatusNotFound)
				}
			} else if strings.Contains(ext, "") { // a bit of a hack to see if a .jsonld exists
				jldobj := fmt.Sprintf("%s.jsonld", object)
				fmt.Println(jldobj)
				err := sendObject(mc, w, r, bucket, jldobj)
				if err != nil {
					log.Printf("Error: %v    Object: %s", err, jldobj)
					http.Error(w, http.StatusText(http.StatusNotFound),
						http.StatusNotFound)
				}
			} else {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusNotFound),
					http.StatusNotFound)
			}
		} else {
			err = sendObject(mc, w, r, bucket, object)
			if err != nil {
				log.Println(err)
				// not found is unlikely given the stat call above..  this
				// likely is only released via internal server error
				http.Error(w, http.StatusText(http.StatusNotFound),
					http.StatusNotFound)
			}
		}
	}

}

func sendHTML(mc *minio.Client, w http.ResponseWriter, r *http.Request, bucket, object, prefix string) error {
	log.Println("Client can understand html")
	w.Header().Set("Content-Type", "text/html")

	log.Println(object)

	fo, err := mc.GetObject(context.Background(), bucket, object, minio.GetObjectOptions{})
	if err != nil {
		log.Println("Failed to get object")
		w.WriteHeader(http.StatusNotFound)
		return err
	}

	var b bytes.Buffer
	bw := bufio.NewWriter(&b)

	_, err = io.Copy(bw, fo)
	if err != nil {
		log.Println("Failed io.Copy")
		log.Println(err)
		return err
	}

	// Get the template from the site assets
	t := fmt.Sprintf("assets/templates/%s/template.html", filepath.Dir(r.URL.Path))
	log.Println(filepath.Dir(r.URL.Path))
	to, err := mc.GetObject(context.Background(), bucket, fmt.Sprintf("%s/%s", prefix, t), minio.GetObjectOptions{})
	if err != nil {
		log.Println("Failed to open template")
		return err
	}

	var tb bytes.Buffer
	tbw := bufio.NewWriter(&tb)

	_, err = io.Copy(tbw, to)
	if err != nil {
		return err
	}

	tc := string(tb.Bytes())

	log.Println(object)
	pd := UFOKNPageData{JSONLD: string(b.Bytes()), PID: object}

	//ht, err := template.New("object template").ParseFiles(t) // open and parse a template text file
	ht, err := template.New("object template").Parse(tc) // open and parse a template text file
	if err != nil {
		return err
	}

	err = ht.ExecuteTemplate(w, "T", pd) // substitute fields in the template 't', with values from 'user' and write it out to 'w' which implements io.Writer
	if err != nil {
		return err
	}

	return err
}

func sendObject(mc *minio.Client, w http.ResponseWriter, r *http.Request, bucket, object string) error {
	fo, err := mc.GetObject(context.Background(), bucket, object, minio.GetObjectOptions{})
	if err != nil {
		return err
	}

	ext := filepath.Ext(object)
	mime := fileactions.MimeByType(ext)

	w.Header().Set("Content-Type", mime) //  m)  // override for now until I update the loaded for samples earth to mod the object name with .jsonld
	_, err = io.Copy(w, fo)              // todo need to stream write from s3 reader...  not copy

	return err
}
