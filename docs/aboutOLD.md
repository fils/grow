# GROW

## What is GROW

GROW stands for "Generic Resource on the Web", but it's realy just a web server that leverages objects storage.   There is actually nothing special about that, many other tools and environments do that as well.  So what made GROW worth writing?

A few things:

* GROW was an attempt to leverage the RDA Digital Object Cloud pattern in a server focused on digital objects like a research organization would use.
* GROW also included a pattern for implementing structured data on the web approaches.  In particular including JSON-LD data graphs in landing pages associated with a digital object.  
* I also wanted a way to express "affordances for resources" .  Here, I will use the work "affordances" to mean operations the data could satisfy.  So, if the object could present itself a sa map, I wanted to be able to do that.  CSV, good...  PDF, fine... Punch Card image ...  I guess.  The main idea was..  content negotiation.  
* I also wanted these negotiated views to be optionally dynamic or static.  So I could either have code function that compute and send the results, or have those results pre-computed and available in the object store.  

So, you can likely see I also wanted to address approaches that supported as much as possible FAIR data principles as well.

Principle of Project
Throughout the work I tried to make sure that the approaches I took could be replicated in other environments. So during this presentation you may say "I could just do this in X".  That's great, that was part of the plan.  It really should be part of anyone's plan.  Indeed, GROW partly grew out of discussions with BCO-DMO and was an attempt to better understand their approaches by implementing them in code.  GROW also come from experiences working with several NSF organizations who published structured data on the net as part of EarthCube P418/p419 work.

Data by Convention
A pattern that evolved out of this work was what could be described as "data by convention".   Analogous to the "Convention over Configuration" approaches in coding, which attempted to reduce the amount of configuration a framework required by establishing a convention that everyone follows.   I've never actually liked software frameworks much (at all) in programming, so I am slightly (highly) annoyed I came up with one for the data store.  I'm happy to be talked off this ledge if people have ideas.  

That said, here is the story.  The main driver for the convention was that I wanted a way to map URI space to object space.

A URL like:
https://mbalpha.org/aboutus.html 
needed to map to an object (and potentially css and js resources to support it.

Additionally, a URL like
https://mbalpha.org/id/artifact/gaurdianofforever
need to map to a digital data object.  

I needed a way to ensure that URLs meant to be part of the web, were separate from data object URLs.  Mostly this is due to negotiated views and special computed workflows I will detail later.  I needed a default namespace that separated the two, and I settled on a "convention" I have used for some time.  The /id/ URL path prefix.

Anything without /id/ at the start of the path, is a resource intended for the web.  Any URL that represents a digital object, will start its path with an /id/.  Note, you can configure this prefex (so..  configuration over convention I guess)  ;)

Object Store structure.  
So how does this map into object storage?  Let's start with a base bucket in our object store.  We will call it demo.  It could be any object store by the way.  The code needs an S3 compliant object store, so; AWS, Minio (my choice), Google Cloud Storage (tested with this too), Azure, Wasabi, others...

A URL like:   
https://mbalpha.org/aboutus.html 

maps to:  /demo/website/aboutus.html  
any relative css like ./css/main.css would be in:
/demo/website/css/main.css 

Additionally, a URL like 
https://mbalpha.org/id/artifact/gaurdianofforever 

maps to: /deme/artifact/timportal

You will note I didn't map ove the ID name.  This was due to the fact I wanted people to be able to set their own prefix and also a few other routing thoughts.   We will see if that was wise or just an annoyance.

There is one last special "convent" object prefix and that is "assets".  I reserved that name for objects the server might need to satisfy an operation.  The driving example was for template files to address server side rendering of pages.  I'll detail this int he object request routing.  (which is next)

Object Request Routing.  
The routing of objects in "website" is easy.  Anything without a /id/ prefix hsould map to its path in website.  The only special case is the URL like https://mbalpha.org/visiting pattern which was must match to a request for index.html in the directory /visiting.

Simpler web routing aside we can now turn to the convent set up for object routing.  The base of the object name can be anything.  I tend to use hashes for the file for various reasons, but that is obviously not required.  We are restricted to name that align with the limits on URI and the host object storage.  This might vary. 

Ok, the pattern

id/artifact/guardianofforever 

is an object.  Following the RDA Digital object pattern we need a metadata object for it.  GROW is built around this metadata being a JSON-LD document.  JSON-LD is a serialization of the RDF data model and as such support all the vocabulary and work that can be represented in RDF.  A vast topic that is out of scope.  

That said then.  Lets look at two requests for this this object.

The simple..  explicate format.
Any explicatly suffexed request will look for that object directlry.  

id/artifact/guardianofforever.html  ->  html page on the Gaurdian
id/artifact/guardianofforever.josnld -> JSON-LD about the Gaurdian
id/artifact/guardianofforever.wkt -> Stellar coordinate for Gaurdian

suffex overrides headers?  (that can't be right...)
That aside, the concept is not that hard.  ;) 

Negotiated requests.
Negotiated requests are those requests for content that state their content type in the request header.  The would be on the object base name only.

req: accept (text/html) on id/artifact/guardianofforever
res: id/artifact/guardianofforever.html content

req: accept (application/ld+json) on id/artifact/guardianofforever
res: id/artifact/guardianofforever.jsonld content

Again..  not that hard.  


Materialize and Computed content responses 
One area I wanted to experiment with in GROW (and again, this is not unique to GROW) is dealing with requests for object serializations that may or may not be materialized.

What does that mean in English?

Grow is configured such that a rquest for say .geojson on a
resource (either by content negotiation or suffex mapping) gets resolved by 1 of 3 ways.

Request:
id/artifact/guardianofforever.kml 
or 
req: accept (text/kml) on id/artifact/guardianofforever
res: id/artifact/guardianofforever.kml content

So..  GROW would go looking for the object

artifact/guardianofforever.kml

However, we may not have a pre-computed KML for the Gaurdian. You would expect GROW to now return an error 415 (unsupported media type).  Note, 404 (not found) would be returned if NO version at all of the base object name "artifact/guardianofforever" exists. 

For GROW, however, there is one more option before that.  While the object store is checked first, a fall back is to look to see if a function has been coded in GROW that allows that view to be computed.  Why do this, wouldn't it always be slower?

* If I have 900K objects, I might only store views that are very commonly requested or require a lot of computation.  I might not want to store every possible combination of a files encoding.
*  Less common and/or very quickly computed views I might choose to compute and send.
*  I can also mix and match, GROW checks for the existance of a computed view first, even if the computed function exists. 
*  Also, this pattern allows me to leverage external functions that I might be able to incorporate.  (Looking at how to do this in a "web architectural" approach now).


So, GROW can serve simple documents but for the special case of "digital objects" it looks for views that are computed first and then looks to see if it can generated that format.  Failing all that it return 415 (or other errors of course for internal error (500), or if the base object doesn't exist at all (404)).  

Special Case "Landing Page"
The last topic is the special case fo the "landing page".  
In the above examples we noted the request for .html and the subsequent response.  It would be completely feasible to pre-render the HTML (ala Hugo or other such means) and place that HTML in the object store to be picked up first based on the previous description.  

However, GROW does also provide for using HTML templates (GROW is written in Go, so these are standard Go templates).  
... details here ...

Why do this?
It means that users of Grow can follow just the RDA Digital Object Cloud pattern mentioned earlier.  Putting a digital object (of any format) into the object store, along wtih a coorsponding .JSON-LD file menas that object can be shared and a basic landing page generated for it without any work.  Also, all computed views that can be supported by the .JSON-LD content are also available. 

This was a key goal.  To allow the data provider to focus on the data and the descriptive metadata for it.  The key elements they are authoratative for, and have GROW do as much out of box as it can while still allowing them to add other materializede views later.  

So in this case a template.html file is looked for in the "assets" directory at the same path in there as the resource has.  So 

id/artifact/guardianofforever.html
means, first we look for that object...  not finding it already existing.... we look for 

assets/artifact/template.html 
and
id/artifact/guardianofforever.jsonld

These are combined to make:
id/artifact/guardianofforever.html

with the .JSON-LD placed into the head of the document.  Out of scope for this document, we then can leverage web components to present human usable views into that JSON-LD like maps, tables, plots, etc. 

Routing exmaples:




Q: Why not just use AWS Lambda functions for web sites...  this has been done?
A: Agreed and it would be easy(ish) to implement this in that enviornment.  I'm not sure all the routes and moving parts are as easy to maintain in AWS land as in satic binary land.   It is just a pattern though, so yes you could.  However, I wanted something I could deploy locally as a binary or simple OCI container in Docker or Podman.  Or, deploy into something like Google Cloud Run.  I wanted options,

Q: Does this scale?
A: Everything is containerized and stateless, so it should scale easily in k8s or docker environments.   It's been deployed to Google Cloud run too (essentially managed k8s) and so there (like other clouds) you can set thresholds that when reached start the scale out process (and scale back.  
 
Q: Why don't you use 303 redirection to not conflate URIs and URLs of resoruces?
A: I know..  I thought of that at the start..   but who is doing that?   It seems wonderful in "logic" space but a pain in the ass in the user space where my clients now need to be 303 aware for loading resources in notebooks etc.   I choose to make one camp happy over another since it seemed unlikely either could be equally happy.  

