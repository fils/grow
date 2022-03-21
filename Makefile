BINARY := server
GROUP := general
DOCKERVER :=`/bin/cat VERSION`
.DEFAULT_GOAL := server

# server for the dynamic site  (routing api and id)
server:
		cd cmd/server ; \
		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 env go build -o $(BINARY)

docker:
		docker build  --tag="fils/grow-$(GROUP):$(DOCKERVER)"  --file=./build/Dockerfile.yml .

dockerlatest:
		docker build  --tag="fils/grow-$(GROUP):latest"  --file=./build/Dockerfile.yml .

publish:  
		docker push fils/grow-$(GROUP):$(DOCKERVER)
		docker push fils/grow-$(GROUP):latest

tag:
	docker tag fils/grow-$(GROUP):$(DOCKERVER) gcr.io/top-operand-112611/grow-$(GROUP):$(DOCKERVER)

publishgcr:
	docker push gcr.io/top-operand-112611/grow-$(GROUP):$(DOCKERVER)

togcr: server docker tag publishgcr
tohub: server docker dockerlatest publish

# server for the static site approach (ala S3 website)

server_static:
		cd cmd/staticsite ; \
		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 env go build -o $(BINARY)_static

docker_static:
		docker build  --tag="fils/grow_static-$(GROUP):$(DOCKERVER)"  --file=./build/Dockerfile_static.yml .

dockerlatest_static:
		docker build  --tag="fils/grow_static-$(GROUP):latest"  --file=./build/Dockerfile_static.yml .

publish_static:  
		docker push fils/grow_static-$(GROUP):$(DOCKERVER)
		docker push fils/grow_static-$(GROUP):latest
