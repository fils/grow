package jld

import (
	"net/http"

	"github.com/piprate/json-gold/ld"
)

// ContextMapping holds the JSON-LD mappings for cached context
type ContextMapping struct {
	Prefix string
	File   string
}

// Proc build the JSON-LD processer and sets the options object
// to use in framing, processing and all JSON-LD actions
// TODO   we create this all the time..  stupidly..  Generate these pointers
// and pass them around, don't keep making it over and over
// Ref:  https://schema.org/docs/howwework.html and https://schema.org/docs/jsonldcontext.json
// func Proc(v1 *viper.Viper) (*ld.JsonLdProcessor, *ld.JsonLdOptions) { // TODO make a booklean
func ProcOpts() (*ld.JsonLdProcessor, *ld.JsonLdOptions) { // TODO make a booklean

	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")

	client := &http.Client{}
	nl := ld.NewDefaultDocumentLoader(client)

	//var s []ContextMapping
	// err := v1.UnmarshalKey("contextmaps", &s)
	// if err != nil {
	// 	log.Println(err)
	// }

	m := make(map[string]string)
	m["http://schema.org/"] = "./context/jsonldcontext.jsonld"
	m["https://schema.org/"] = "./context/jsonldcontext.jsonld"

	// for i := range s {
	// 	if fileExists(s[i].File) {
	// 		m[s[i].Prefix] = s[i].File // m["http://schema.org/"] = "/context/file"  // loaded as part of the build?  or from s3?

	// 	} else {
	// 		log.Printf("ERROR: context file location %s is wrong, this is a critical error", s[i].File)
	// 	}
	// }

	// Read mapping from config file
	cdl := ld.NewCachingDocumentLoader(nl)
	cdl.PreloadWithMapping(m)
	options.DocumentLoader = cdl

	// Set a default format..  let this be set later...
	options.Format = "application/nquads"

	return proc, options
}
