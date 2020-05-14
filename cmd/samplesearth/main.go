package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"oceanleadership.org/grow/internal/fileobjects"
	"oceanleadership.org/grow/internal/samplesearth"

	"github.com/gorilla/mux"
	minio "github.com/minio/minio-go"
)

var s3addressVal, s3bucketVal, s3prefixVal, keyVal, secretVal string
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
	flag.StringVar(&s3prefixVal, "prefix", "website", "bucket prefix for the objects")
	flag.StringVar(&keyVal, "key", "config", "Object server key")
	flag.StringVar(&secretVal, "secret", "config", "Object server secret")
}

func main() {
	// parse environment vars
	s3addressVal = os.Getenv("S3ADDRESS")
	s3bucketVal = os.Getenv("S3BUCKET")
	s3prefixVal = os.Getenv("S3PREFIX")
	keyVal = os.Getenv("S3KEY")
	secretVal = os.Getenv("S3SECRET")

	// Parse the flags if any, will override the environment vars
	flag.Parse() // parse any command line flags...
	log.Printf("a: %s  b %s  p %s  k %s  s %s\n", s3addressVal, s3bucketVal, s3prefixVal, keyVal, secretVal)

	// Need to convert this to gocloud.dev bloc (https://gocloud.dev/howto/blob/)
	mc, err := minio.New(s3addressVal, keyVal, secretVal, false)
	if err != nil {
		log.Println(err)
	}

	// Samples Earth DOC route, bucket needs to now reflect the location of the DOC objects
	doc := mux.NewRouter()
	doc.PathPrefix("/id/").Handler(http.StripPrefix("/id/", minioHandler(mc, s3bucketVal, s3prefixVal, samplesearth.DO)))
	doc.NotFoundHandler = http.HandlerFunc(notFound)
	http.Handle("/id/", &MyServer{doc})

	// dr  default route
	dr := mux.NewRouter()
	if localVal {
		dr.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./local"))))
	} else {
		dr.PathPrefix("/").Handler(http.StripPrefix("/", minioHandler(mc, s3bucketVal, s3prefixVal, fileobjects.FileObjects)))
	}
	dr.NotFoundHandler = http.HandlerFunc(notFound)
	http.Handle("/", &MyServer{dr})

	// Start the server...
	log.Printf("About to listen on 8080. Go to http://127.0.0.1:8080/")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func notFound(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/404.html", 303)
}

func minioHandler(minioClient *minio.Client, bucket, prefix string, f func(minioClient *minio.Client, bucket, prefix string, w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { f(minioClient, bucket, prefix, w, r) })
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
