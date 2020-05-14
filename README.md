# Go Object Web

## about

This is a simple proof of concept program.  It is a Go based program
that is just a simple web server.  The only unique element is that I 
use minio (s3) as a storage for the files.

## future

The implementation could be improved and also I have not worked in 
any https or http 2.0 support.   I have code to support these as well
as Let's Encrypt support so I hope to work that in.  

I will also likely move from Minio client to the Go Cloud Dev client (ref: https://gocloud.dev/howto/blob/)
to allow easy access to Google, MS and AWS object stores from one
code base.  

The other goal would be ensure I can docker-ize this and deploy
to something like Google Cloud Run or Amazon ECS.  

## commands

I will soon work this up as a Docker system and use the viper config
pattern.  It will run by default looking for the config file which 
can be loaded into the docker image.  

I will also keep the command line option for testing and I need 
to activate the local file source option via flag too.

example command line with object store option 
```bash
go run cmd/server/main.go -address 192.168.0.1:1234  -key mykey -secret mysecret
```

## refs

* https://stackoverflow.com/questions/35245649/aws-s3-large-file-reverse-proxying-with-golangs-http-responsewriter

## Better name..

Go Resource Object Web Server (GROW Server)
