BINARY := server
DOCKERVER :=`/bin/cat VERSION`
.DEFAULT_GOAL := linux

linux:
		cd cmd/$(BINARY) ; \
		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 env go build -o $(BINARY)

docker:
		docker build  --tag="fils/goobjectweb:$(DOCKERVER)"  --file=./build/Dockerfile .

dockerlatest:
		docker build  --tag="fils/goobjectweb:latest"  --file=./build/Dockerfile .

publish:  
		docker push fils/goobjectweb:$(DOCKERVER)
		docker push fils/goobjectweb:latest

tag:
	docker tag fils/goobjectweb:$(DOCKERVER) gcr.io/top-operand-112611/goobjectweb:$(DOCKERVER)

publishgcr:
	docker push gcr.io/top-operand-112611/goobjectweb:$(DOCKERVER)

togcr: linux docker tag publishgcr
