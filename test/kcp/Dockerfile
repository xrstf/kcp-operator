# This Dockerfile creates a container image with kcp's source code in it, so
# we can deploy it as a Pod inside a kind cluster and then run the kcp e2e
# tests against the kcp-operator managed environment.

FROM docker.io/golang:1.26.2

WORKDIR /apps/kcp
RUN git clone --depth 1 https://github.com/kcp-dev/kcp . && \
    go build -v ./test/...

COPY entrypoint.sh /usr/local/bin/entrypoint.sh

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
