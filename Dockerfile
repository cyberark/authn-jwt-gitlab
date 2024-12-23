FROM golang:1.22.7 as builder

WORKDIR /go/src/github.com/cyberark/authn-jwt-gitlab
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

FROM golang:1.22.7-alpine as builder-alpine

WORKDIR /go/src/github.com/cyberark/authn-jwt-gitlab
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

FROM ubuntu:22.04 as ubuntu

RUN groupadd -r appuser && useradd -r -g appuser appuser

COPY --from=builder /go/bin/authn-jwt-gitlab /authn-jwt-gitlab

RUN apt-get update && \
    apt-get install -y ca-certificates && \
    update-ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

USER appuser

FROM alpine:3.18 as alpine

RUN addgroup -S appuser && adduser -S appuser -G appuser

RUN apk update && \
    apk upgrade --no-cache && \
    apk add --no-cache libcrypto3 libssl3 && \
    apk add --no-cache ca-certificates && \
    update-ca-certificates

COPY --from=builder-alpine /go/bin/authn-jwt-gitlab /authn-jwt-gitlab

USER appuser

CMD ["/authn-jwt-gitlab"]

FROM redhat/ubi9:9.3 as ubi

RUN yum -y update && \
    yum clean all && \
    rm -rf /var/cache/yum

RUN groupadd -r appuser && useradd -r -g appuser appuser

COPY --from=builder /go/bin/authn-jwt-gitlab /authn-jwt-gitlab

USER appuser

CMD ["/authn-jwt-gitlab"]
