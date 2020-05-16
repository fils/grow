package operations

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"oceanleadership.org/grow/internal/fileactions"
	"oceanleadership.org/grow/pkg/objservices/spatial"

	"github.com/minio/minio-go"
)

// TypeGeoJSON takes a DO of type JSON-LD and coverts it to GeoJSON.
// The function atempts to get teh JSON-LD object and
// then convert and send
func TypeGeoJSON(mc *minio.Client, w http.ResponseWriter, r *http.Request, bucket, object string) error {

	jldobj := strings.Replace(object, ".geojson", ".jsonld", 1)
	fo, err := mc.GetObject(bucket, jldobj, minio.GetObjectOptions{})
	if err != nil {
		return err
	}

	var b bytes.Buffer
	bw := bufio.NewWriter(&b)

	_, err = io.Copy(bw, fo)
	if err != nil {
		return err
	}

	gj, err := spatial.SDO2GeoJSON(string(b.Bytes()))
	if err != nil {
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
