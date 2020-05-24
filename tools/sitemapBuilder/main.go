package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go"
	"github.com/snabb/sitemap"
)

// TODO..   look at https://github.com/ikeikeikeike/go-sitemap-generator  as well
// The ikei one is almost too much..  just make an [][]string for the sites.  Where
// the first array elements is the sitemap name.  Then build the sitemap index from
//that.   In my DO pattern
//  /website index any .html
//  /do index all objects that do NOT have .jsonld
//  /dp (if I keep this) index all objects that do NOT have .jsonld
//  /assets  ignore
//  all others..   index distinct base name (no extension)

type SiteMapIndex struct {
	XMLName  xml.Name  `xml:"sitemapindex"`
	Sitemaps []Sitemap `xml:"sitemap"`
}

type Sitemap struct {
	Loc     string `xml:"loc"`
	Lastmod string `xml:"lastmod"`
}

type URLSet struct {
	XMLName xml.Name  `xml:"urlset"`
	URLs    []URLNode `xml:"url"`
}

type URLNode struct {
	Loc        string  `xml:"loc"`
	Lastmod    string  `xml:"lastmod"`
	Changefreq string  `xml:"changefreq"`
	Priority   float64 `xml:"priority"`
}

var urlprefix, s3addressVal, s3bucketVal, s3prefixVal, keyVal, secretVal string
var htmlonly bool

func init() {
	log.SetFlags(log.Lshortfile)

	//flag.BoolVar(&htmlonly, "htmlonly", false, "Require sitemap entries end in .html")

	flag.StringVar(&urlprefix, "urlprefix", "https://example.org/", "URL prefix for the sitemap")
	flag.StringVar(&s3addressVal, "server", "0.0.0.0:0000", "Address of the object server with port")
	flag.StringVar(&s3bucketVal, "bucket", "website", "bucket which holds the web site objects")
	flag.StringVar(&s3prefixVal, "prefix", "website", "bucket prefix for the objects")
	flag.StringVar(&keyVal, "key", "config", "Object server key")
	flag.StringVar(&secretVal, "secret", "config", "Object server secret")
}

func main() {
	fmt.Println("Sitemap builder")

	// parse environment vars
	urlprefix = os.Getenv("URLPREFIX")
	s3addressVal = os.Getenv("S3ADDRESS")
	s3bucketVal = os.Getenv("S3BUCKET")
	s3prefixVal = os.Getenv("S3PREFIX")
	keyVal = os.Getenv("S3KEY")
	secretVal = os.Getenv("S3SECRET")

	flag.Parse() // parse any command line flags...
	log.Printf("uL %s a: %s  b %s  p %s  k %s  s %s\n", urlprefix, s3addressVal, s3bucketVal, s3prefixVal, keyVal, secretVal)

	start := time.Now()
	s3Maker(urlprefix, s3addressVal, s3bucketVal, s3prefixVal, keyVal, secretVal)
	elapsed := time.Since(start)
	log.Printf("s3Maker took %s", elapsed)

}

func s3Maker(prefix, server, bucket, objPrefix, key, secret string) {
	mc := minioConnection(server, "80", key, secret)

	// Create a done channel.
	doneCh := make(chan struct{})
	defer close(doneCh)
	sm := sitemap.New()
	recursive := true
	c := 0
	c2 := 0
	var a []string

	for message := range mc.ListObjectsV2(bucket, objPrefix, recursive, doneCh) {
		c = c + 1

		// fmt.Println(message.Key)
		k := strings.SplitAfterN(message.Key, "/", 3)
		x := message.LastModified.UTC()

		if strings.Contains(k[1], "website/") {
			u := fmt.Sprintf("%s%s", prefix, k[2])
			if strings.HasSuffix(u, ".html") {
				sm.Add(&sitemap.URL{Loc: u, LastMod: &x})
			}
		} else {
			u := fmt.Sprintf("%sid/%s%s", prefix, k[1], k[2])
			if !strings.Contains(u, "assets/") {
				sm.Add(&sitemap.URL{Loc: u, LastMod: &x})
			}
		}

		if c > 40000 {
			// fmt.Println(c)
			c = 0
			saveto := fmt.Sprintf("./output/sitemap_%d_%s.xml", c2, objPrefix)
			a = append(a, strings.TrimPrefix(saveto, "./output/"))
			c2 = c2 + 1
			f, err := os.Create(saveto)
			if err != nil {
				log.Println(err)
			}
			sm.WriteTo(f)
			sm = sitemap.New()
		}
	}

	// need to save the last few in the loop
	saveto := fmt.Sprintf("./output/sitemap_%d_%s.xml", c2, objPrefix)
	a = append(a, strings.TrimPrefix(saveto, "./output/"))
	f, err := os.Create(saveto)
	if err != nil {
		log.Println(err)
	}

	// var b bytes.Buffer
	// foo := bufio.NewWriter(&b)
	// sm.WriteTo(foo)
	// write to minio now...

	sm.WriteTo(f)

	smi := sitemap.NewSitemapIndex()
	for x := range a {
		smi.Add(&sitemap.URL{Loc: fmt.Sprintf("%s%s", prefix, a[x])})
	}

	saveto = fmt.Sprint("./output/sitemap.xml")
	f, err = os.Create(saveto)
	if err != nil {
		log.Println(err)
	}
	smi.WriteTo(f)
}

func minioConnection(minioVal, portVal, accessVal, secretVal string) *minio.Client {
	endpoint := fmt.Sprintf("%s:%s", minioVal, portVal)
	accessKeyID := accessVal
	secretAccessKey := secretVal
	useSSL := false
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Println(err)
	}
	return minioClient
}
