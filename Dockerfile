FROM golang:1.23.7 AS builder

WORKDIR /go/src/github.com/cyberark/authn-jwt-gitlab
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

COPY VERSION /go/src/github.com/cyberark/authn-jwt-gitlab/VERSION

FROM golang:1.23.7-alpine AS builder-alpine

WORKDIR /go/src/github.com/cyberark/authn-jwt-gitlab
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

COPY VERSION /go/src/github.com/cyberark/authn-jwt-gitlab/VERSION

FROM ubuntu:24.04 AS ubuntu

RUN groupadd -r appuser && useradd -r -g appuser appuser

COPY --from=builder /go/bin/authn-jwt-gitlab /authn-jwt-gitlab
COPY --from=builder /go/src/github.com/cyberark/authn-jwt-gitlab/VERSION /VERSION

RUN apt-get update && \
    apt-get install -y ca-certificates && \
    update-ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

USER appuser

FROM alpine:3.21.3 AS alpine

RUN addgroup -S appuser && adduser -S appuser -G appuser

RUN apk update && \
    apk upgrade --no-cache && \
    apk add --no-cache libcrypto3 libssl3 && \
    apk add --no-cache ca-certificates && \
    update-ca-certificates

COPY --from=builder-alpine /go/bin/authn-jwt-gitlab /authn-jwt-gitlab
COPY --from=builder-alpine /go/src/github.com/cyberark/authn-jwt-gitlab/VERSION /VERSION

USER appuser

CMD ["/authn-jwt-gitlab"]

FROM redhat/ubi9:9.5-1742918310 AS ubi

RUN yum -y update && \
    yum clean all && \
    rm -rf /var/cache/yum

RUN groupadd -r appuser && useradd -r -g appuser appuser

COPY --from=builder /go/bin/authn-jwt-gitlab /authn-jwt-gitlab
COPY --from=builder /go/src/github.com/cyberark/authn-jwt-gitlab/VERSION /VERSION

USER appuser

CMD ["/authn-jwt-gitlab"]
