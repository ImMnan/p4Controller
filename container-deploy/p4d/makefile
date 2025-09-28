
.PHONY: all download build tag push

VERSION ?= 0.4.0
IMAGE = immnan/p4d

all: download build tag push

download:
	wget https://ftp.perforce.com/pub/perforce/r25.1/bin.linux26x86_64/p4d
	wget https://ftp.perforce.com/pub/perforce/r25.1/bin.linux26x86_64/p4

build:
	docker build -t p4d .

tag:
	docker tag p4d $(IMAGE):$(VERSION)
	docker tag p4d $(IMAGE):latest
push:
	docker push $(IMAGE):$(VERSION)
	docker push $(IMAGE):latest
	docker images
	

    