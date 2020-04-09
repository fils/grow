package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	minio "github.com/minio/minio-go"
)

var s3addressVal, keyVal, secretVal string
var localVal bool

// MyServer is the Gorilla mux router structure
type MyServer struct {
	r *mux.Router
}

func init() {
	log.SetFlags(log.Lshortfile)
	// log.SetOutput(ioutil.Discard) // turn off all logging

	// Place holder for logging to buffer..   could use this later for
	// writing logs to Minio.  Perhaps useless if fronting with traefik.
	// var (
	//      buf    bytes.Buffer
	//      logger = log.New(&buf, "logger: ", log.Lshortfile)
	// )

	flag.BoolVar(&localVal, "local", false, "Server file local over object store, false by default")
	flag.StringVar(&s3addressVal, "server", "0.0.0.0:0000", "Address of the object server with port")
	flag.StringVar(&s3bucketVal, "bucket", "website", "bucket which holds the web site objects")
	flag.StringVar(&keyVal, "key", "config", "Object server key")
	flag.StringVar(&secretVal, "secret", "config", "Object server secret")
}

func main() {
	flag.Parse() // parse any command line flags...

	// Need to convert this to gocloud.dev bloc (https://gocloud.dev/howto/blob/)
	mc, err := minio.New(s3addressVal, keyVal, secretVal, false)
	if err != nil {
		log.Println(err)
	}

	vocroute := mux.NewRouter()
	if localVal {
		vocroute.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./local"))))
	} else {
		vocroute.PathPrefix("/").Handler(http.StripPrefix("/", minioHandler(mc, vocCore)))
	}
	vocroute.NotFoundHandler = http.HandlerFunc(notFound)
	http.Handle("/", &MyServer{vocroute})

	// Start the server...
	log.Printf("About to listen on 8080. Go to http://127.0.0.1:8080/")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func vocCore(mc *minio.Client, w http.ResponseWriter, r *http.Request) {
	key := fmt.Sprintf("%s", r.URL.Path)

	// TODO review this hack...
	// the browser rewrites /index.html to / and this of course fails to locate the object.
	// So for the one case of index.html (.htm ?) we need to do this hack
	if key == "" { // deal with browser index.html rewrites
		key = "index.html"
	}

	m := MimeByType(filepath.Ext(key))
	w.Header().Set("Content-Type", m)
	log.Printf("%s: %s \n", key, m)

	fo, err := mc.GetObject("website", key, minio.GetObjectOptions{})
	if err != nil {
		log.Println(err)
	}

	n, err := io.Copy(w, fo)
	log.Println(n)
	if err != nil {
		log.Println("Issue with writing file to http response")
		log.Println(err)
	}
}

func minioHandler(minioClient *minio.Client, f func(minioClient *minio.Client, w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { f(minioClient, w, r) })
}

// MimeByType matches file extensions to mimetype
func MimeByType(e string) string {
	t := mime.TypeByExtension(e)
	if t == "" {
		t = "application/octet-stream"
	}
	return t
}

func notFound(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/404.html", 303)
}

func (s *MyServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	rw.Header().Set("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	// Stop here if its Preflighted OPTIONS request
	if req.Method == "OPTIONS" {
		return
	}

	// Let Gorilla work
	s.r.ServeHTTP(rw, req)
}

func addDefaultHeaders(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fn(w, r)
	}
}
