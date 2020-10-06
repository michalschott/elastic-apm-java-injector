#!/usr/bin/env sh

CA_BUNDLE=$(kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.certificate-authority-data}')

kind load docker-image local/elastic-apm-java-injector:latest

sed "s/CA_BUNDLE/$CA_BUNDLE/" deploy.yaml | kubectl apply -f-

kubectl delete rs -l app=elastic-apm-java-injector

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  labels:
    elastic-apm-java-injector: enabled
  name: test
EOF

cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test
  namespace: test
  labels:
    app: test
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - args:
        - while true; do sleep 30; done;
        command:
        - /bin/sh
        - -c
        - --
        env:
        image: alpine:3.12
        imagePullPolicy: IfNotPresent
        name: alpine
EOF