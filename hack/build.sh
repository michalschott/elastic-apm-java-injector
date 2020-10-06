#!/usr/bin/env sh

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -o ../build/elastic-apm-java-injector -ldflags='-s -w' ../cmd/elastic-apm-java-injector/main.go
docker build  -t local/elastic-apm-java-injector:latest -f ../Dockerfile .. 