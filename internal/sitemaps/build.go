package sitemaps

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/minio/minio-go"
	"github.com/snabb/sitemap"
	"oceanleadership.org/grow/internal/objectstore"
)

// SiteMapIndex is t he index of sitemaps
type SiteMapIndex struct {
	XMLName  xml.Name  `xml:"sitemapindex"`
	Sitemaps []Sitemap `xml:"sitemap"`
}

// Sitemap struct
type Sitemap struct {
	Loc     string `xml:"loc"`
	Lastmod string `xml:"lastmod"`
}

// URLSet array of URLNodes
type URLSet struct {
	XMLName xml.Name  `xml:"urlset"`
	URLs    []URLNode `xml:"url"`
}

// URLNode is the resource entry in the sitemap
type URLNode struct {
	Loc        string  `xml:"loc"`
	Lastmod    string  `xml:"lastmod"`
	Changefreq string  `xml:"changefreq"`
	Priority   float64 `xml:"priority"`
}

// put back an upper level function that call the sitemap builder as a
// go func   https://www.reddit.com/r/golang/comments/9504fe/how_to_handle_go_routines_with_rest_api/

// Build launches a go func to build.   Needs to return a
func Build(mc *minio.Client, bucket, prefix, domain string, w http.ResponseWriter, r *http.Request) {

	go builder(bucket, prefix, domain, mc)

	// TODO need to write back a 202 Accepted
}

func builder(bucket, prefix, domain string, mc *minio.Client) {

	// Create a done channel.
	doneCh := make(chan struct{})
	defer close(doneCh)
	sm := sitemap.New()
	recursive := true
	c := 0
	c2 := 0
	var a []string

	for message := range mc.ListObjectsV2(bucket, prefix, recursive, doneCh) {
		c = c + 1

		// fmt.Println(message.Key)
		k := strings.SplitAfterN(message.Key, "/", 3)
		x := message.LastModified.UTC()

		if strings.Contains(k[1], "website/") {
			u := fmt.Sprintf("%s%s", domain, k[2])
			if strings.HasSuffix(u, ".html") {
				sm.Add(&sitemap.URL{Loc: u, LastMod: &x})
			}
		} else {
			u := fmt.Sprintf("%sid/%s%s", domain, k[1], k[2])
			if !strings.Contains(u, "assets/") {
				sm.Add(&sitemap.URL{Loc: u, LastMod: &x})
			}
		}

		if c > 40000 {
			// fmt.Println(c)
			c = 0
			saveto := fmt.Sprintf("%s/website/sitemap_%d_%s.xml", prefix, c2, prefix)
			a = append(a, strings.TrimPrefix(saveto, fmt.Sprintf("%s/website/", prefix))) //s = strings.TrimPrefix(s, "¡¡¡Hello, ")
			c2 = c2 + 1
			var b bytes.Buffer
			foo := bufio.NewWriter(&b)
			sm.WriteTo(foo)
			foo.Flush()
			log.Println("Write sitemap to minio")
			_, err := objectstore.LoadToMinio(b.Bytes(), bucket, saveto, mc)
			if err != nil {
				log.Println(err)
			}
			sm = sitemap.New()
		}
	}

	// need to save the last few in the loop
	saveto := fmt.Sprintf("%s/website/sitemap_%d_%s.xml", prefix, c2, prefix)
	a = append(a, strings.TrimPrefix(saveto, fmt.Sprintf("%s/website/", prefix))) //s = strings.TrimPrefix(s, "¡¡¡Hello, ")
	var b bytes.Buffer
	foo := bufio.NewWriter(&b)
	sm.WriteTo(foo)
	foo.Flush()
	_, err := objectstore.LoadToMinio(b.Bytes(), bucket, saveto, mc)
	if err != nil {
		log.Println(err)
	}

	smi := sitemap.NewSitemapIndex()
	log.Println(a)
	for x := range a {
		smi.Add(&sitemap.URL{Loc: fmt.Sprintf("%s%s", domain, a[x])})
	}

	saveto = fmt.Sprintf("%s/website/sitemap.xml", prefix)
	var b2 bytes.Buffer
	foo2 := bufio.NewWriter(&b2)
	smi.WriteTo(foo2)
	foo2.Flush()
	_, err = objectstore.LoadToMinio(b2.Bytes(), bucket, saveto, mc)
	if err != nil {
		log.Println(err)
	}

}
