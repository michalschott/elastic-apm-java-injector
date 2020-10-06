elastic-apm-java-injector was born to automate process described [here](https://www.elastic.co/blog/using-elastic-apm-java-agent-on-kubernetes-k8s).

# Development
- install [kind](https://github.com/kubernetes-sigs/kind)
- create kind cluster (hack/kind_up.sh)
- generate and deploy certs (hack/gen_certs.sh)
- build (hack/build.sh)
- deploy (hack/deploy.sh)

Rinse and repeat with build and deploy steps.

# TO-DO:
- add some flags (ie loglevel)
- move mutating functions from main.go to dedicated pkg
- inject more variables, ideally configurable with annotations
- configurable image to inject
```
- name: ELASTIC_APM_SERVER_CERT
  valueFrom:
    secretKeyRef:
      name: apm-server-apm-server-cert
      key: tls.crt
- name: ELASTIC_APM_IGNORE_SERVER_CERT
  value: "true"
- name: ELASTIC_APM_SERVER_URL 
  value: "http://apm-server-apm-http:8200" 
- name: ELASTIC_APM_SERVICE_NAME 
  value: "petclinic" 
- name: ELASTIC_APM_APPLICATION_PACKAGES 
  value: "org.springframework.samples.petclinic" 
- name: ELASTIC_APM_ENVIRONMENT 
  value: test 
- name: ELASTIC_APM_LOG_LEVEL 
  value: DEBUG 
- name: ELASTIC_APM_SECRET_TOKEN 
  valueFrom: 
    secretKeyRef: 
      name: apm-server-apm-token 
      key: secret-token
```
- setup checks (golangci-lint, gosec)  // GH actions
- goreleaser // GH actions
- public docker image to use // GH actions