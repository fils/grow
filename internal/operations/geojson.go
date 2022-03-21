package operations

import (
	"bufio"
	"bytes"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/fils/goobjectweb/internal/fileactions"
	"github.com/fils/goobjectweb/pkg/objservices/spatial"

	"github.com/minio/minio-go"
)

// TypeGeoJSON takes a DO of type JSON-LD and coverts it to GeoJSON.
// The function attempts to get teh JSON-LD object and
// then convert and send
func TypeGeoJSON(mc *minio.Client, w http.ResponseWriter, r *http.Request, bucket, object string) error {

	jldobj := strings.Replace(object, ".geojson", ".jsonld", 1)
	fo, err := mc.GetObject(bucket, jldobj, minio.GetObjectOptions{})
	if err != nil {
		log.Println("Error reading object in TypeGeoJSON")
		return err
	}

	var b bytes.Buffer
	bw := bufio.NewWriter(&b)

	_, err = io.Copy(bw, fo)
	if err != nil {
		log.Println("Error copying object in TypeGeoJSON")
		return err
	}

	gj, err := spatial.SDO2GeoJSON(string(b.Bytes()))
	if err != nil {
		log.Println("Spatial Call error in TypeGeoJSON")
		return err
	}
	sr := strings.NewReader(gj)

	// log.Println(filepath.Ext(object))
	mime := fileactions.MimeByType(filepath.Ext(object))

	// m := fileactions.MimeByType(filepath.Ext(key))
	w.Header().Set("Content-Type", mime) //  m)  // override for now until I update the loaded for samples earth to mod the object name with .jsonld

	_, err = io.Copy(w, sr) // todo need to stream write from s3 reader...  not copy

	return err
}
