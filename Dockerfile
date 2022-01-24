ARG TARGET=server

# Can be used in case a proxy is necessary
ARG GOPROXY

# Build tcheck binary
FROM golang:1.17-alpine3.15 AS tcheck

WORKDIR /go/src/github.com/uber/tcheck

COPY go.* ./
RUN go build -mod=readonly -o /go/bin/tcheck github.com/uber/tcheck

# Build Cadence binaries
FROM golang:1.17-alpine3.13 AS builder

ARG RELEASE_VERSION

RUN apk add --update --no-cache ca-certificates make git curl mercurial unzip

WORKDIR /cadence

# Making sure that dependency is not touched
ENV GOFLAGS="-mod=readonly"

# Copy go mod dependencies and build cache
COPY go.* ./
RUN go mod download

COPY . .
RUN rm -fr .bin .build

ENV CADENCE_RELEASE_VERSION=$RELEASE_VERSION

# bypass codegen, use committed files.  must be run separately, before building things.
RUN make .fake-codegen
RUN CGO_ENABLED=0 make copyright cadence-cassandra-tool cadence-sql-tool cadence cadence-server cadence-bench cadence-canary


# Download dockerize
FROM alpine:3.15 AS dockerize

RUN apk add --no-cache openssl

ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && echo "**** fix for host id mapping error ****" \
    && chown root:root /usr/local/bin/dockerize


# Alpine base image
FROM alpine:3.11 AS alpine

RUN apk add --update --no-cache ca-certificates tzdata bash curl

# set up nsswitch.conf for Go's "netgo" implementation
# https://github.com/gliderlabs/docker-alpine/issues/367#issuecomment-424546457
RUN test ! -e /etc/nsswitch.conf && echo 'hosts: files dns' > /etc/nsswitch.conf

SHELL ["/bin/bash", "-c"]


# Cadence server
FROM alpine AS cadence-server

ENV CADENCE_HOME /etc/cadence
RUN mkdir -p /etc/cadence

COPY --from=tcheck /go/bin/tcheck /usr/local/bin
COPY --from=dockerize /usr/local/bin/dockerize /usr/local/bin
COPY --from=builder /cadence/cadence-cassandra-tool /usr/local/bin
COPY --from=builder /cadence/cadence-sql-tool /usr/local/bin
COPY --from=builder /cadence/cadence /usr/local/bin
COPY --from=builder /cadence/cadence-server /usr/local/bin
COPY --from=builder /cadence/schema /etc/cadence/schema

COPY docker/entrypoint.sh /docker-entrypoint.sh
COPY config/dynamicconfig /etc/cadence/config/dynamicconfig
COPY config/credentials /etc/cadence/config/credentials
COPY docker/config_template.yaml /etc/cadence/config
COPY docker/start-cadence.sh /start-cadence.sh

WORKDIR /etc/cadence

ENV SERVICES="history,matching,frontend,worker"

EXPOSE 7933 7934 7935 7939
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD /start-cadence.sh


# All-in-one Cadence server
FROM cadence-server AS cadence-auto-setup

RUN apk add --update --no-cache ca-certificates py3-pip mysql-client
RUN pip3 install cqlsh && cqlsh --version

COPY docker/start.sh /start.sh

CMD /start.sh


# Cadence CLI
FROM alpine AS cadence-cli

COPY --from=tcheck /go/bin/tcheck /usr/local/bin
COPY --from=builder /cadence/cadence /usr/local/bin

ENTRYPOINT ["cadence"]

# Cadence Canary
FROM alpine AS cadence-canary

COPY --from=builder /cadence/cadence-canary /usr/local/bin
COPY --from=builder /cadence/cadence /usr/local/bin

CMD ["/usr/local/bin/cadence-canary", "--root", "/etc/cadence-canary", "start"]

# Cadence Bench
FROM alpine AS cadence-bench

COPY --from=builder /cadence/cadence-bench /usr/local/bin
COPY --from=builder /cadence/cadence /usr/local/bin

CMD ["/usr/local/bin/cadence-bench", "--root", "/etc/cadence-bench", "start"]

# Final image
FROM cadence-${TARGET}
