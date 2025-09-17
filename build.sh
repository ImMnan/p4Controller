#!/bin/sh

wget https://ftp.perforce.com/pub/perforce/r25.1/bin.linux26x86_64/p4d

wget https://ftp.perforce.com/pub/perforce/r25.1/bin.linux26x86_64/p4

echo "Tagging and pushing Docker image..."

docker build -t p4d .
docker tag p4d immnan/p4d:0.3.8
docker push immnan/p4d:0.3.8
docker push immnan/p4d:latest

echo "Listing Docker images..."
docker images