FROM busybox:1.33.0

ADD build/elastic-apm-java-injector /

ENTRYPOINT ["/elastic-apm-java-injector"]
