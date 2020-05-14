package main

import (
	"fmt"
	"log"
	"time"

	cache "github.com/patrickmn/go-cache"
	"github.com/yterajima/go-sitemap"
)

func main() {
	start := time.Now()
	smap, err := sitemap.Get("http://samples.earth/sitemap.xml", nil)
	if err != nil {
		fmt.Println(err)
	}

	elapsed := time.Since(start)
	log.Printf("Download took %s", elapsed)

	// Print URL in sitemap.xml
	// for _, URL := range smap.URL {
	//	fmt.Printf("%s : %s ", URL.Loc, URL.LastMod)
	// }

	start2 := time.Now()
	c := cache.New(5*time.Minute, 10*time.Minute)

	for k, _ := range smap.URL {
		// if smap.URL[k].Loc == "http://samples.earth/id/bjv5omqu6s75joelhni0" {
		//	fmt.Println("Found it!")
		// }
		c.Set(smap.URL[k].Loc, smap.URL[k].LastMod, cache.NoExpiration)
	}

	elapsed = time.Since(start2)
	log.Printf("Cache loading took %s", elapsed)

	// now do 1 million searches
	start3 := time.Now()
	const ISOTime = "1978-07-25 17:25:59 -0500 CDT"
	for k, _ := range smap.URL {
		if x, found := c.Get(smap.URL[k].Loc); found {
			t, _ := time.Parse(time.RFC3339, x.(string))
			// if err != nil {
			//	log.Println(err)
			//}
			fmt.Println(t) // 1999-12-31 00:00:00 +0000 UTC
		}
	}
	elapsed = time.Since(start3)
	log.Printf("Look ups took %s", elapsed)

}
