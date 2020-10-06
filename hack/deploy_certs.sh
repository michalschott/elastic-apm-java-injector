#!/usr/bin/env sh

kubectl create secret generic elastic-apm-java-injector-certs --from-file=elastic-apm-java-injector.pem=hack/elastic-apm-java-injector.pem --from-file=elastic-apm-java-injector.key=hack/elastic-apm-java-injector.key
