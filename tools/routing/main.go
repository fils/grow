package main

import (
	"fmt"
	"path/filepath"
)

func main() {

	// paths
	// 1 send object direclty   DO exists with no extension & no accept html
	// 2 send object directly   DO exists with extension that is not .html
	// 3 render html page       DO exists with extension that is .html, requires DO.jsonld
	// 4 render html page       DO exists with no extension & accept html, requires DO.jsonld
	// 5 process data and send  DO exists with no or different extension that reuqest, but
	//                               we know how to render that view (ie, like .geojson)

	do := []string{"foo", "foo.html", "foo.geojson"}

	isobj := true //  fo stat
	acptHTML := true

	// is ext == .html or acptHTML is true we are gooing to send HTMl

	// there is an object with that extension (so send it)
	// that object exist (or just its .jsonld?) and was requested with .html or with accept html but no extension, the .jsonld is there so
	//        so take it and render it (render html)
	// object.X is requested, but does not exist.  object exists and a mathcing service route exist..   so prcoess by that route
	//        make a map[string]function and make .ext to a function   (use this for .html too?)

	for x := range do {
		ext := filepath.Ext(do[x])

		if ext == "" && isobj && acptHTML {
			fmt.Printf("Object found: %t   Accepts HTML: %t   Extension:   %s \n", isobj, acptHTML, ext)
		}

		if ext == ".html" && isobj && acptHTML {
			fmt.Printf("Object found: %t   Accepts HTML: %t   Extension:   %s \n", isobj, acptHTML, ext)
		}

		if ext == ".geojson" && isobj && acptHTML {
			fmt.Printf("Object found: %t   Accepts HTML: %t   Extension:   %s \n", isobj, acptHTML, ext)
		}

	}

}
