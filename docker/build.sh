#!/bin/sh

cd ..
docker build . -t ghcr.io/chikyukido/nande:latest
docker push ghcr.io/chikyukido/nande:latest
