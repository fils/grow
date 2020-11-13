package sitemaps

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/fils/goobjectweb/internal/objectstore"
	"github.com/minio/minio-go"
	"github.com/snabb/sitemap"
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

	// TODO   // ? https://schema.org/CreateAction https://schema.org/docs/actions.html
	// POST a request JSON body to /api/sitemap
	// make PID
	// Make a JSON status document in /id/queue/[PID].json
	// send of go func (let it know the PID)
	// send 202 Accepted, location /id/queue/[PID] (any FET request to resource is 200 with info about job)
	// When go fun is done ned to make /id/queue/[PID]  303 with link to new resournce made,
	// rewrite the doc with a new status entry?

	//  use the presence of /id/queue/[PID]  to tell if this is a 200 report or 303 redirect
	// PID.json is the metadata for PID
	//  How to delete the queue (if to delete it) needs to be resolved.
}

// TODO..  this following function SUCKS..  it's a hackish proceedural flow that "works" but
// is fragile to changes..   based on this I see how it really should be done and need to
// update this soon...

func builder(bucket, prefix, domain string, mc *minio.Client) {

	fmt.Printf("Bucket: %s  Prefix: %s   Domain: %s\n", bucket, prefix, domain)

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

		var k []string

		if prefix == "" {
			k = strings.SplitAfterN(message.Key, "/", 2)
		} else {
			k = strings.SplitAfterN(message.Key, "/", 3)
		}

		x := message.LastModified.UTC()

		// fmt.Println(message.Key)
		// fmt.Println(k)
		// fmt.Println((prefix == ""))
		// fmt.Println(len(k))

		if prefix == "" {
			if strings.Contains(k[0], "website/") { // BUG..  if prefix == "" then these indexes are off..
				u := fmt.Sprintf("%s%s", domain, k[1])
				if strings.HasSuffix(u, ".html") {
					sm.Add(&sitemap.URL{Loc: u, LastMod: &x})
				}
			} else {
				u := fmt.Sprintf("%sid/%s%s", domain, k[0], k[1])
				if !strings.Contains(u, "assets/") {
					if strings.HasSuffix(u, ".jsonld") {
						sm.Add(&sitemap.URL{Loc: u, LastMod: &x})
					}
				}
			}
		} else {
			if strings.Contains(k[1], "website/") { // BUG..  if prefix == "" then these indexes are off..
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
		}

		if c > 40000 {
			// fmt.Println(c)
			c = 0
			saveto := ""
			if prefix == "" {
				saveto = fmt.Sprintf("website/sitemap_%d.xml", c2)
			} else {
				saveto = fmt.Sprintf("%s/website/sitemap_%d_%s.xml", prefix, c2, prefix)
			}

			if prefix == "" {
				a = append(a, strings.TrimPrefix(saveto, fmt.Sprintf("website/"))) //s = strings.TrimPrefix(s, "¡¡¡Hello, ")
			} else {
				a = append(a, strings.TrimPrefix(saveto, fmt.Sprintf("%s/website/", prefix))) //s = strings.TrimPrefix(s, "¡¡¡Hello, ")
			}
			// a = append(a, strings.TrimPrefix(saveto, fmt.Sprintf("%s/website/", prefix))) //s = strings.TrimPrefix(s, "¡¡¡Hello, ")

			c2 = c2 + 1
			var b bytes.Buffer
			foo := bufio.NewWriter(&b)
			sm.WriteTo(foo)
			foo.Flush()
			log.Printf("Write sitemap %s", saveto)
			_, err := objectstore.LoadToMinio(b.Bytes(), bucket, saveto, mc)
			if err != nil {
				log.Println(err)
			}
			sm = sitemap.New()
		}
	}

	// need to save the last few in the loop
	saveto := ""
	if prefix == "" {
		saveto = fmt.Sprintf("website/sitemap_%d.xml", c2)
	} else {
		saveto = fmt.Sprintf("%s/website/sitemap_%d_%s.xml", prefix, c2, prefix)
	}

	fmt.Printf("Save to index: %s \n", saveto)

	if prefix == "" {
		a = append(a, strings.TrimPrefix(saveto, fmt.Sprintf("website/"))) //s = strings.TrimPrefix(s, "¡¡¡Hello, ")
	} else {
		a = append(a, strings.TrimPrefix(saveto, fmt.Sprintf("%s/website/", prefix))) //s = strings.TrimPrefix(s, "¡¡¡Hello, ")
	}
	// a = append(a, strings.TrimPrefix(saveto, fmt.Sprintf("%s/website/", prefix))) //s = strings.TrimPrefix(s, "¡¡¡Hello, ")

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

	if prefix == "" {
		saveto = fmt.Sprintf("website/sitemap.xml")
	} else {
		saveto = fmt.Sprintf("%s/website/sitemap.xml", prefix)
	}

	var b2 bytes.Buffer
	foo2 := bufio.NewWriter(&b2)
	smi.WriteTo(foo2)
	foo2.Flush()
	_, err = objectstore.LoadToMinio(b2.Bytes(), bucket, saveto, mc)
	if err != nil {
		log.Println(err)
	}

}
