![GitHub Logo](./docs/images/growShield.png)

# Generic Resource Object Web (GROW)

## about

This is a simple proof of concept program.  
It is a Go based program that serves objects.  These objects 
come from an S3 API object store.  It can use the open source Minio 
S3 backend or any of the AWS, Google Cloud or Azure object storage 
services.   I've tested this mostly on Docker Swarm and Google Cloud Run.

> Note:  GROW is a very simplistic object server.  For a more sophisticate
> solution with far more features try Clouder. (https://clowder.ncsa.illinois.edu/)

GROW leverages the RDA Digital Object Cloud pattern and is a basic implementation 
of that pattern.

For more details visit [the about page.](./docs/about.md)



![GitHub Logo](./docs/images/objecChain.png)

## commands

GROW can be run from the command line for Docker.

example command line with object store option 
```bash
go run cmd/server/main.go -domain "https://example.org" -server 192.168.86.45:1234 
-bucket sites -prefix siteprefix -key mykey -secret mysecret
```

The elements are:

```
domain: Your sites domain (needed when generating sitemaps)

server: The address for the object storage such as Minio

bucket: The bucket your object tree is stored in

prefix:  Optional prefix for your object tree root

key:  Object store key

secret: Object store secret
```

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

