package digitalobjects

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/minio/minio-go"
	"oceanleadership.org/grow/internal/fileactions"
	"oceanleadership.org/grow/internal/operations"
)

// UFOKNPageData is the struct for the template page
type UFOKNPageData struct {
	JSONLD  string
	PID     string
	GeoJSON string
}

// DO pulls the objects from the object store.  At present this function
// container the routing logic.
func DO(mc *minio.Client, bucket, prefix, domain string, w http.ResponseWriter, r *http.Request) {
	// collect the information we need for routing resolution
	// at this time I am collecting a lot..   we may not need all of this
	object := fmt.Sprintf("%s/%s", prefix, r.URL.Path)
	base := filepath.Base(r.URL.Path)
	ext := filepath.Ext(r.URL.Path)
	baseobj := strings.TrimSuffix(base, ext)
	mt := mime.TypeByExtension(ext)
	acpt := r.Header.Get("Accept")
	objInfo, err := mc.StatObject(bucket, object, minio.StatObjectOptions{})
	if err != nil {
		log.Print(err)
	} else {
		log.Println(objInfo)
	}

	// WARNING Hackish hard coding for testing....
	// see if baseobj maps to a service ID
	log.Printf("/assets/services/%s.jsonld", baseobj)
	serviceInfo, err := mc.StatObject(bucket, fmt.Sprintf("%s/assets/services/%s.jsonld", prefix, baseobj), minio.StatObjectOptions{})
	if err != nil {
		log.Print(err)
	} else {
		log.Println(serviceInfo)
	}

	log.Printf("%s %s %s %s %s %s ", acpt, object, ext, mt, base, baseobj)

	// First deal with requests that are looking for HTML via the ACCEPT header
	// assume they want a landing page represenation of the DO
	if strings.Contains(acpt, "text/html") {
		if ext == "" || ext == ".jsonld" || ext == ".html" {
			s := strings.TrimSuffix(r.URL.Path, ext)
			object := fmt.Sprintf("%s/%s.jsonld", prefix, s) // if prefex is nill?
			log.Printf("b: %s o: %s p:%s ", bucket, object, prefix)
			err := sendHTML(mc, w, r, bucket, object, prefix)
			if err != nil {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)
			}
		} else {
			log.Printf("Requested HTML, but I can not resolve oject to that media type at this time\n")
			http.Error(w, http.StatusText(http.StatusUnsupportedMediaType),
				http.StatusUnsupportedMediaType)
		}
	} else {
		// If we are in this ELSE we are NOT looking for  HTML (by HEADER)
		//  at this point, so we might be sending an object
		// or attempting to render a representation of one.
		// 1) check if the object exist and send it.
		// 2) if does not exist, check if the ext matches a render version

		// See if the object exists
		_, err := mc.StatObject(bucket, object, minio.StatObjectOptions{})
		// fmt.Println(objInfo)

		// If the OBJECT DOES NOT seem to exist
		if err != nil {
			// TODO Eval switch statementi vs if else if

			// TODO service call branch test
			// need (base object name (no ext), ext and mimetype we have, requested mimetype,  map of external services and mime they op on)

			// affordance check (look for services that do something..   likely reported in a HEAD call for the like)

			// What do we need to know?
			// in /id/do/p1/p2/p3
			// is p3 a serviceID?  this is the base...  which is either a service id or the end of an object prefix string
			// or is p3 the end of the object name with several prefixes (prefixi?)

			// invoke(serviceObject, targetObject) // optionsla JSON:API package?
			// services are in /assets/services  need a func to parse and select from these...
			// json-ld flatten  -> look for type action -> get entrypoint and grab the template, method and type

			// here down is OLD logic prior to the dynamic routing
			if strings.Contains(ext, ".geojson") {
				err := operations.TypeGeoJSON(mc, w, r, bucket, object)
				if err != nil {
					log.Println(err)
					http.Error(w, http.StatusText(http.StatusNotFound),
						http.StatusNotFound)
				}
			} else if strings.Contains(ext, "") { // a bit of a hack to see if a .jsonld exists and send it as a default
				jldobj := fmt.Sprintf("%s.jsonld", object) // I don't like this..  if we can not resolve explicitly, don't assume
				err := sendObject(mc, w, r, bucket, jldobj)
				if err != nil {
					log.Println(err)
					http.Error(w, http.StatusText(http.StatusNotFound),
						http.StatusNotFound)
				}
			} else {
				log.Println(err)
				http.Error(w, http.StatusText(http.StatusNotFound),
					http.StatusNotFound)
			}
		} else { // Thje object DOES exist..   send it..  :)
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
	fmt.Println("Client can understand html")
	w.Header().Set("Content-Type", "text/html")

	fo, err := mc.GetObject(bucket, object, minio.GetObjectOptions{})
	if err != nil {
		return err
	}

	var b bytes.Buffer
	bw := bufio.NewWriter(&b)

	_, err = io.Copy(bw, fo)
	if err != nil {
		return err
	}

	// Get the template from the site assets
	t := fmt.Sprintf("assets/templates/%s/template.html", filepath.Dir(r.URL.Path))
	log.Println(filepath.Dir(r.URL.Path))
	to, err := mc.GetObject(bucket, fmt.Sprintf("%s/%s", prefix, t), minio.GetObjectOptions{})
	if err != nil {
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

	fo, err := mc.GetObject(bucket, object, minio.GetObjectOptions{})
	if err != nil {
		return err
	}

	ext := filepath.Ext(object)

	log.Println(ext)
	mime := fileactions.MimeByType(ext)

	// m := fileactions.MimeByType(filepath.Ext(key))
	w.Header().Set("Content-Type", mime) //  m)  // override for now until I update the loaded for samples earth to mod the object name with .jsonld

	_, err = io.Copy(w, fo) // todo need to stream write from s3 reader...  not copy

	return err
}
