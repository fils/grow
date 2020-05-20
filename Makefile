BINARY := server
GROUP := general
DOCKERVER :=`/bin/cat VERSION`
.DEFAULT_GOAL := linux

linux:
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

togcr: linux docker tag publishgcr
tohub: linux docker dockerlatest publish
