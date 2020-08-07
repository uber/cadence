ARG TARGET=server

# Can be used in case a proxy is necessary
ARG GOPROXY


# Build tcheck binary
FROM golang:1.13.6-alpine AS tcheck

RUN apk add --update --no-cache ca-certificates git curl

ENV GO111MODULE=off

RUN curl https://glide.sh/get | sh

ENV TCHECK_VERSION=v1.1.0

RUN go get -d github.com/uber/tcheck
RUN cd /go/src/github.com/uber/tcheck && git checkout ${TCHECK_VERSION}

WORKDIR /go/src/github.com/uber/tcheck

RUN glide install

RUN go install


# Build Cadence binaries
FROM golang:1.13.6-stretch AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
		libxml2 build-essential \
	    && rm -rf /var/lib/apt/lists/*

WORKDIR /cadence

# Making sure that dependency is not touched
ENV GOFLAGS="-mod=readonly" 

RUN go get -d github.com/ibmdb/go_ibm_db

RUN cd /go/src/github.com/ibmdb/go_ibm_db/installer \
    && go run setup.go


ENV DB2HOME=/go/src/github.com/ibmdb/go_ibm_db/installer/clidriver
ENV CGO_CFLAGS=-I/go/src/github.com/ibmdb/go_ibm_db/installer/clidriver/include
ENV CGO_LDFLAGS=-L/go/src/github.com/ibmdb/go_ibm_db/installer/clidriver/lib
ENV LD_LIBRARY_PATH=/go/src/github.com/ibmdb/go_ibm_db/installer/clidriver/lib:$LD_LIBRARY_PATH

RUN ls -la ./

# Copy go mod dependencies and build cache
COPY go.* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 make copyright cadence-cassandra-tool cadence-sql-tool cadence cadence-server


# Download dockerize
FROM alpine:3.11 AS dockerize

RUN apk add --no-cache openssl

ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && echo "**** fix for host id mapping error ****" \
    && chown root:root /usr/local/bin/dockerize


# Base image
FROM buildpack-deps:stretch-scm AS base

RUN apt-get update && apt-get install -y --no-install-recommends \
		netcat \
        libxml2 \
	    && rm -rf /var/lib/apt/lists/*

SHELL ["/bin/bash", "-c"]


# Cadence server
FROM base AS cadence-server

ENV CADENCE_HOME /etc/cadence
RUN mkdir -p /etc/cadence

COPY --from=tcheck /go/bin/tcheck /usr/local/bin
COPY --from=dockerize /usr/local/bin/dockerize /usr/local/bin
COPY --from=builder /go/src /go/src
COPY --from=builder /cadence/cadence-cassandra-tool /usr/local/bin
COPY --from=builder /cadence/cadence-sql-tool /usr/local/bin
COPY --from=builder /cadence/cadence /usr/local/bin
COPY --from=builder /cadence/cadence-server /usr/local/bin
COPY --from=builder /cadence/schema /etc/cadence/schema

COPY docker/entrypoint.sh /docker-entrypoint.sh
COPY config/dynamicconfig /etc/cadence/config/dynamicconfig
COPY docker/config_template.yaml /etc/cadence/config
COPY docker/start-cadence.sh /start-cadence.sh

WORKDIR /etc/cadence

ENV SERVICES="history,matching,frontend,worker"
ENV LD_LIBRARY_PATH=/go/src/github.com/ibmdb/go_ibm_db/installer/clidriver/lib:$LD_LIBRARY_PATH

EXPOSE 7933 7934 7935 7939
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD /start-cadence.sh


# All-in-one Cadence server
FROM cadence-server AS cadence-auto-setup

COPY docker/start.sh /start.sh

CMD /start.sh


# Cadence CLI
FROM base AS cadence-cli

COPY --from=tcheck /go/bin/tcheck /usr/local/bin
COPY --from=builder /cadence/cadence /usr/local/bin

ENTRYPOINT ["cadence"]


# Final image
FROM cadence-${TARGET}
