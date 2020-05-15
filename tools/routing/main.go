package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func main() {

	// TODO
	// make a function to build geojson
	// make a map[".geojson"]function    that points to it...
	// registeredFunction  // is key in map

	// STEP 1 Check if valid object request first
	// Object checkign is a bit of a waste..   there are cases where this is try but
	// no render path is ultimately available.  Any real reason for this step?
	fmt.Printf("\n -------- step 1: Check for object\n")
	oid := []string{"foo", "foo.jsonld", "foo.html", "foo.geojson"}
	for x := range oid {
		ext := filepath.Ext(oid[x])
		fmt.Printf("\n--- For %s \n", oid[x])
		if ext == "" {
			fmt.Printf("Check for: %s or %s%s\n", oid[x], oid[x], ".jsonld")
			// on failure return http error
		} else {
			fmt.Printf("Check for: %s or %s\n", oid[x], strings.TrimSuffix(oid[x], ext))
			// on failure return http error
		}
	}

	// STEP 2 resolve if accepts html or not  (later not html could also resolve to map of functions )
	fmt.Printf("\n -------- step 2  HTML requested, \nlook for object to render html via template (need JSON-LD object for that)\n")
	oid = []string{"foo", "foo.jsonld", "foo.html", "foo.geojson"}
	acptHTML := false
	if acptHTML {
		for x := range oid {
			ext := filepath.Ext(oid[x])
			fmt.Printf("\n--- For %s \n", oid[x])
			if ext == "" || ext == ".jsonld" || ext == ".html" {
				s := strings.TrimSuffix(oid[x], ext)
				fmt.Printf("Check for: %s%s  If found, locate template and render\n", s, ".jsonld")
			} else {
				fmt.Printf("I don't know how to render HTMl for that extension, will check function map\n")
			}
		}
	} else {
		for x := range oid {
			fmt.Printf("\n--- For %s \n", oid[x])
			fmt.Printf("Send: %s as %s \n", oid[x], "resolved mime type by ext or object metadata")
			// if XYZ.ext doeesn't exist, checl .ext in function map.  if found try that.
			// functions will have "required object instances" like .jsonld or others.
		}
	}

}
