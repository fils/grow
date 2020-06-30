![GitHub Logo](./docs/images/growShield.png)

# Generic Resource Object Web (GROW)

## about

This is a simple proof of concept program.  
It is a Go based program that serves objects.  These objects 
come from an S3 API object store.  It can use the open source Minio 
S3 backend or any of the AWS, Google Cloud or Azure object storage 
services.   I've tested this mostly on Docker Swarm and Google Cloud Run.

> Note:  GROW is a very simplist object server.  For a more sophisticate
> solution with far more features try Clouder. (https://clowder.ncsa.illinois.edu/)

GROW leverages the RDA Digital Object Cloud pattern and is a basic implementation 
of that pattern.   


![GitHub Logo](./docs/images/objectChain.png)

## future

I will likely move from Minio client to the Go Cloud Dev client (ref: https://gocloud.dev/howto/blob/)
to allow easy access to Google, MS and AWS object stores from one
code base.  

A basic set of APIs exist that can be invoked to perform two functions:

* Build a sitemap of the objects (leverages sitemap indexes for > 50K object counts)
* Build an RDF graph based on converting stored JSON-LD objects into a single NQuads RDF file.  

I'll add in object store triggers to call web hooks to peform workflows on addedd objects.

## commands

GROW can be run from the command line for Docker.

example command line with object store option 
```bash
go run cmd/server/main.go -domain "https://example.org" -server 192.168.86.45:1234 
-bucket sites -prefix siteprefix -key mykey -secret mysecret
```

The elements are:

domain: Your sites domain (needed when generating sitemaps)

server: The address for the object storage such as Minio

bucket: The bucket your object tree is stored in

prefix:  Optional prefix for your object tree root

key:  Object store key

secret: Object store secret


On Docker this would look like:

```Docker
    image: fils/grow-general:latest
    environment:
      - S3ADDRESS=s3system:9000
      - S3BUCKET=sites
      - S3PREFIX=siteprefix
      - DOMAIN=https://examples.org/
      - S3KEY=mykey
      - S3SECRET=mysecret
    labels:
    ...
    networks:
      - traefik_default

```

