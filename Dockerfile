FROM golang:1.17 as builder

ARG VERSION

RUN go install github.com/korylprince/fileenv@v1.1.0
RUN go install "github.com/korylprince/ls-relay-cert@$VERSION"

RUN apt-get update && apt-get install -y \
    curl \
    build-essential \
    autoconf \
    libxml2-dev \
    libssl-dev \
    libz-dev \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir /build && cd /build && \
    curl -L -O https://github.com/downloads/mackyle/xar/xar-1.6.1.tar.gz && \
    tar xzf xar-1.6.1.tar.gz && \
    cd xar-1.6.1 && \
    sed -i 's/OpenSSL_add_all_ciphers/CRYPTO_free/g' configure.ac && \
    ./autogen.sh && make


FROM ubuntu:20.04

RUN apt-get update && apt-get install -y \
    curl \
    libxml2 \
    libssl1.1 \
    zlib1g \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /go/bin/fileenv /
COPY --from=builder /go/bin/ls-relay-cert /
COPY --from=builder /build/xar-1.6.1/src/xar /usr/bin/

CMD ["/fileenv", "/ls-relay-cert"]
