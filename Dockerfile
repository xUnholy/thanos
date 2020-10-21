FROM golang:1.15-alpine as builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT=""

ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=${TARGETOS} \
  GOARCH=${TARGETARCH} \
  GOARM=${TARGETVARIANT}

WORKDIR /go/src/github.com/thanos-io/thanos

RUN apk update && apk upgrade && apk add --no-cache alpine-sdk git make

COPY . .

RUN make build

FROM quay.io/prometheus/busybox:latest

LABEL maintainer="The Thanos Authors"

COPY --from=builder /go/bin/thanos /bin/thanos

ENTRYPOINT [ "/bin/thanos" ]
