elastic-apm-java-injector was born to automate process described [here](https://www.elastic.co/blog/using-elastic-apm-java-agent-on-kubernetes-k8s).

ensure to label your desired namespace with `elastic-apm-java-injector: enabled`

# Development
- install [kind](https://github.com/kubernetes-sigs/kind)
- create kind cluster (hack/kind_up.sh)
- generate and deploy certs (hack/gen_certs.sh)
- build (hack/build.sh)
- deploy (hack/deploy.sh)

Rinse and repeat with build and deploy steps.

# TO-DO:
- ~~add some flags (ie loglevel)~~
- ~~move mutating functions from main.go to dedicated pkg~~
- ~~inject more variables~~
- ~~configurable image to inject~~
- setup checks (golangci-lint, gosec)  // GH actions
- goreleaser // GH actions
- public docker image to use // GH actions

For variable injection, since this webhook relies on `namespaceSelector.matchLabels` to operate, having various instances of webhook with different settings is desireable. Thus I took env vars config approach.