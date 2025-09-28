
.PHONY: all download build tag push

VERSION ?= 0.1.0
IMAGE = immnan/p4c

all: download build tag push

download:
	wget https://ftp.perforce.com/pub/perforce/r25.1/bin.linux26x86_64/p4

build:
	go build -o p4controller ../cmd/p4controller/main.go
	docker build -t p4c .

tag:
	docker tag p4c $(IMAGE):$(VERSION)

push:
	docker push $(IMAGE):$(VERSION)
	docker push $(IMAGE):latest
    docker images
	

    