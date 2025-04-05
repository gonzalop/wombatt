# docker build -t wombatt --build-arg=ARCH=arm64 --build-arg=OS=linux --build-arg=TARGETPLATFORM=arm64 .
ARG BUILD_IMAGE=golang:1.24
ARG BASE=alpine
FROM --platform=$BUILDPLATFORM ${BUILD_IMAGE} AS build
SHELL [ "/bin/bash", "-ec" ]

COPY . /go/src

ARG TARGETPLATFORM
RUN ARCH=$(echo $TARGETPLATFORM |cut -f2 -d/);OS=$(echo $TARGETPLATFORM |cut -f1 -d/); \
    cd /go/src; CGO_ENABLED=0 GOOS=${OS} GOARCH=${ARCH} go build -ldflags "-s -w"

FROM --platform=${TARGETPLATFORM} ${BASE}
COPY --from=build /go/src/wombatt /wombatt
# 20 is the dialout group that gives access to serial ports.
RUN echo wombatt:x:1001:20::/:/bin/nologin >> /etc/passwd
USER wombatt:dialout
ENTRYPOINT ["/wombatt"]
