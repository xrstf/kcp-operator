# This Dockerfile creates a container image with kcp's source code in it, so
# we can deploy it as a Pod inside a kind cluster and then run the kcp e2e
# tests against the kcp-operator managed environment.

FROM docker.io/golang:1.26.2

RUN apt update && apt install -y jq

WORKDIR /apps/kcp

# get kcp, compile the kcp binaries for the few tests
# that start a private kcp server (like the cache replication tests)
RUN git clone --depth 1 https://github.com/kcp-dev/kcp . && \
    make build

ENV KCP_BINARIES_DIR=/apps/kcp/bin

# sanity check
RUN "$KCP_BINARIES_DIR/kcp" --version

COPY entrypoint.sh /usr/local/bin/entrypoint.sh

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
