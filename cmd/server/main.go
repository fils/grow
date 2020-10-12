package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/fils/goobjectweb/internal/api/graph"
	"github.com/fils/goobjectweb/internal/api/sitemaps"
	"github.com/fils/goobjectweb/internal/digitalobjects"
	"github.com/fils/goobjectweb/internal/fileobjects"

	"github.com/gorilla/mux"
	minio "github.com/minio/minio-go"
)

var s3addressVal, s3bucketVal, s3prefixVal, domainVal, keyVal, secretVal string
var localVal bool

// MyServer is the Gorilla mux router structure
type MyServer struct {
	r *mux.Router
}

func init() {
	log.SetFlags(log.Lshortfile)

	flag.BoolVar(&localVal, "local", false, "Server file local over object store, false by default")
	flag.StringVar(&s3addressVal, "server", "0.0.0.0:0000", "Address of the object server with port")
	flag.StringVar(&s3bucketVal, "bucket", "website", "bucket which holds the web site objects")
	flag.StringVar(&s3prefixVal, "prefix", "website", "bucket prefix for the objects")
	flag.StringVar(&domainVal, "domain", "example.org", "domain of our served web site")
	flag.StringVar(&keyVal, "key", "config", "Object server key")
	flag.StringVar(&secretVal, "secret", "config", "Object server secret")
}

func main() {
	// parse environment vars
	s3addressVal = os.Getenv("S3ADDRESS")
	s3bucketVal = os.Getenv("S3BUCKET")
	s3prefixVal = os.Getenv("S3PREFIX")
	domainVal = os.Getenv("DOMAIN")
	keyVal = os.Getenv("S3KEY")
	secretVal = os.Getenv("S3SECRET")

	// Parse the flags if any, will override the environment vars
	flag.Parse() // parse any command line flags...
	log.Printf("a: %s  b %s  p %s d %s k %s  s %s\n", s3addressVal, s3bucketVal, s3prefixVal, domainVal, keyVal, secretVal)

	// Need to convert this to gocloud.dev bloc (https://gocloud.dev/howto/blob/)
	mc, err := minio.New(s3addressVal, keyVal, secretVal, false)
	if err != nil {
		log.Println(err)
	}

	// Handler sm:   builds sitemaps
	sm := mux.NewRouter()
	sm.PathPrefix("/api/sitemap").Handler(http.StripPrefix("/api/", minioHandler(mc, s3bucketVal, s3prefixVal, domainVal, sitemaps.Build)))
	sm.PathPrefix("/api/graph").Handler(http.StripPrefix("/api/", minioHandler(mc, s3bucketVal, s3prefixVal, domainVal, graph.Build)))

	sm.NotFoundHandler = http.HandlerFunc(notFound)
	http.Handle("/api/", &MyServer{sm})

	// Handler doc:   addresses the /id/* request path
	doc := mux.NewRouter()
	doc.PathPrefix("/id/").Handler(http.StripPrefix("/id/", minioHandler(mc, s3bucketVal, s3prefixVal, domainVal, digitalobjects.DO)))
	doc.NotFoundHandler = http.HandlerFunc(notFound)
	http.Handle("/id/", &MyServer{doc})

	// Handler dr:   addresses the / request path
	dr := mux.NewRouter()
	if localVal {
		dr.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./local"))))
	} else {
		dr.PathPrefix("/").Handler(http.StripPrefix("/", minioHandler(mc, s3bucketVal, s3prefixVal, domainVal, fileobjects.FileObjects)))
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

func minioHandler(minioClient *minio.Client, bucket, prefix, domain string, f func(minioClient *minio.Client, bucket, prefix, domain string, w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { f(minioClient, bucket, prefix, domain, w, r) })
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
