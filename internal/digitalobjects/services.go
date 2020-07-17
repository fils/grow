package digitalobjects

import (
	"fmt"
	"io/ioutil"

	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go"
)

// Service send an object to a service and routes the results back
func Service(mc *minio.Client, bucket, prefix, domain string, w http.ResponseWriter, r *http.Request) {

	// at this time I am collecting a lot..   we may not need all of this
	object := fmt.Sprintf("%s/%s", prefix, r.URL.Path)
	base := filepath.Base(r.URL.Path)
	ext := filepath.Ext(r.URL.Path)
	baseobj := strings.TrimSuffix(base, ext)
	mt := mime.TypeByExtension(ext)
	acpt := r.Header.Get("Accept")

	// hack
	// split off the serivce ID from the object
	newobject := strings.TrimSuffix(object, "/"+base)
	log.Println(newobject)

	objInfo, err := mc.StatObject(bucket, newobject, minio.StatObjectOptions{})
	if err != nil {
		log.Print(err)
	} else {
		log.Println(objInfo)
	}
	log.Printf("%s %s %s %s %s %s ", acpt, object, ext, mt, base, baseobj)

	// TODO..  read the object

	fo, err := mc.GetObject(bucket, newobject, minio.GetObjectOptions{})
	if err != nil {
		log.Println(err)
	}
	/*
		// send as body
		var tb bytes.Buffer
		tbw := bufio.NewWriter(&tb)

		_, err = io.Copy(tbw, fo)
		if err != nil {
			log.Println(err)
		}
	*/

	// Just a test POST call for now
	var client http.Client // why do I make this here..  can I use 1 client?  move up in the loop
	urlloc := "https://us-central1-top-operand-112611.cloudfunctions.net/function-1"
	req, err := http.NewRequest("POST", urlloc, fo)
	if err != nil {
		log.Printf("#error on %s : %s  ", urlloc, err) // print an message containing the index (won't keep order)
	}
	req.Header.Set("User-Agent", "EarthCube_DataBot/1.0")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("#error on %s : %s  ", urlloc, err) // print an message containing the index (won't keep order)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
	}

	fmt.Fprintf(w, string(body))
}
