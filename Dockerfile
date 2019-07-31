# Build cadence binaries
FROM golang:1.12.7-alpine AS builder

RUN apk add --update --no-cache ca-certificates make git curl mercurial bzr

# Install dep
ENV DEP_VERSION=0.5.4
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | INSTALL_DIRECTORY=/usr/local/bin DEP_RELEASE_TAG=v${DEP_VERSION} sh

WORKDIR /go/src/github.com/uber/cadence

COPY Gopkg.* ./
RUN dep ensure -v -vendor-only

COPY . .
RUN sed -i 's/dep-ensured//g' Makefile
RUN CGO_ENABLED=0 make copyright cadence-cassandra-tool cadence-sql-tool cadence cadence-server


# Download dockerize
FROM alpine:3.10 AS dockerize

RUN apk add --no-cache openssl

ENV DOCKERIZE_VERSION v0.6.1
RUN wget https://github.com/jwilder/dockerize/releases/download/$DOCKERIZE_VERSION/dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && tar -C /usr/local/bin -xzvf dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz \
    && rm dockerize-alpine-linux-amd64-$DOCKERIZE_VERSION.tar.gz


# Final image
FROM alpine:3.10

RUN apk add --update --no-cache ca-certificates tzdata bash curl

# set up nsswitch.conf for Go's "netgo" implementation
# https://github.com/gliderlabs/docker-alpine/issues/367#issuecomment-424546457
RUN [ ! -e /etc/nsswitch.conf ] && echo 'hosts: files dns' > /etc/nsswitch.conf

SHELL ["/bin/bash", "-c"]

ENV CADENCE_HOME /etc/cadence
RUN mkdir -p /etc/cadence

COPY --from=dockerize /usr/local/bin/dockerize /usr/local/bin
COPY --from=builder /go/src/github.com/uber/cadence/cadence-cassandra-tool /usr/local/bin
COPY --from=builder /go/src/github.com/uber/cadence/cadence-sql-tool /usr/local/bin
COPY --from=builder /go/src/github.com/uber/cadence/cadence /usr/local/bin
COPY --from=builder /go/src/github.com/uber/cadence/cadence-server /usr/local/bin
COPY --from=builder /go/src/github.com/uber/cadence/schema /etc/cadence/schema

COPY docker/entrypoint.sh /docker-entrypoint.sh
COPY docker/config /etc/cadence/config

WORKDIR /etc/cadence

ENV SERVICES="history,matching,frontend,worker"

EXPOSE 7933 7934 7935 7939
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD dockerize -template /etc/cadence/config/template.yaml:/etc/cadence/config/docker.yaml cadence-server --root $CADENCE_HOME --env docker start --services=$SERVICES
