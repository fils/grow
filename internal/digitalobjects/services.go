package digitalobjects

import (
	"fmt"
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
	objInfo, err := mc.StatObject(bucket, object, minio.StatObjectOptions{})
	if err != nil {
		log.Print(err)
	} else {
		log.Println(objInfo)
	}
	log.Printf("%s %s %s %s %s %s ", acpt, object, ext, mt, base, baseobj)

	fmt.Fprintf(w, "hello world")
}
