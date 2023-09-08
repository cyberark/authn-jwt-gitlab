FROM golang:1.19 as builder

WORKDIR /go/src/github.com/cyberark/authn-jwt-gitlab
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

FROM golang:1.19-alpine as builder-alpine

WORKDIR /go/src/github.com/cyberark/authn-jwt-gitlab
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

FROM ubuntu:lunar as ubuntu

COPY --from=builder /go/bin/authn-jwt-gitlab /authn-jwt-gitlab

RUN apt-get update && \
    apt-get install -y ca-certificates && \
    update-ca-certificates && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

FROM alpine:3.18 as alpine

COPY --from=builder-alpine /go/bin/authn-jwt-gitlab /authn-jwt-gitlab

RUN apk add --no-cache ca-certificates && \
    update-ca-certificates

CMD ["/authn-jwt-gitlab"]

FROM redhat/ubi9:9.2 as ubi

COPY --from=builder /go/bin/authn-jwt-gitlab /authn-jwt-gitlab

CMD ["/authn-jwt-gitlab"]