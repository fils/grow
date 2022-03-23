package main

import (
	"flag"
	"fmt"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"net/http"
	"os"
	"strconv"

	"github.com/fils/goobjectweb/internal/digitalobjects"
	log "github.com/sirupsen/logrus"

	"github.com/fils/goobjectweb/internal/fileobjects"

	"github.com/gorilla/mux"
	minio "github.com/minio/minio-go/v7"
)

var s3addressVal, s3bucketVal, s3prefixVal, domainVal, keyVal, secretVal string
var localVal, s3SSLVal bool

// MyServer is the Gorilla mux router structure
type MyServer struct {
	r *mux.Router
}

func init() {

	// Output to stdout instead of the default stderr. Can be any io.Writer, see below for File example

	// name the file with the date and time
	//const layout = "2006-01-02-15-04-05"
	//t := time.Now()
	//lf := fmt.Sprintf("grow-%s.log", t.Format(layout))

	lf := fmt.Sprint("grow.log")

	LogFile := lf // log to custom file
	logFile, err := os.OpenFile(LogFile, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
		return
	}

	log.SetFormatter(&log.JSONFormatter{}) // Log as JSON instead of the default ASCII formatter.
	log.SetReportCaller(true)              // include file name and line number
	log.SetOutput(logFile)

	//log.SetLevel(log.WarnLevel) // Only log the warning severity or above.

	flag.BoolVar(&localVal, "local", false, "Serve file local over object store, false by default")
	flag.BoolVar(&s3SSLVal, "ssl", false, "S3 access is SSL, false by default for docker network backend")
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
	s3SSLVal, err := strconv.ParseBool(os.Getenv("S3SSL"))
	if err != nil {
		log.Println("Error reading SSL booling flag")
	}

	// TODO move to viper config for this app  (pass tika URL)

	// Parse the flags if any, will override the environment vars
	flag.Parse() // parse any command line flags...

	log.Printf("a: %s  b %s  p %s d %s k %s  s %s  ssl %v \n", s3addressVal, s3bucketVal, s3prefixVal, domainVal, keyVal, secretVal, s3SSLVal)

	// Need to convert this to gocloud.dev bloc (https://gocloud.dev/howto/blob/)
	//mc, err := minio.New(s3addressVal, keyVal, secretVal, s3SSLVal)
	mc, err := minio.New(s3addressVal,
		&minio.Options{Creds: credentials.NewStaticV4(keyVal, secretVal, ""),
			Secure: s3SSLVal})
	if err != nil {
		log.Println(err)
	}

	// Handler doc:   addresses the /id/* request path
	doc := mux.NewRouter()
	doc.PathPrefix("/id/").Handler(http.StripPrefix("/id/", minioHandler(mc, s3bucketVal, s3prefixVal, domainVal, digitalobjects.DO)))
	//doc.PathPrefix("/id/").Handler(http.StripPrefix("/", minioHandler(mc, s3bucketVal, s3prefixVal, domainVal, digitalobjects.DO)))
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
