# docker build -t wombatt --build-arg=ARCH=arm64 --build-arg=OS=linux --build-arg=TARGETPLATFORM=arm64 .
ARG BUILD_IMAGE=golang:1.25-alpine
ARG BASE=alpine

FROM --platform=$BUILDPLATFORM ${BUILD_IMAGE} AS build
COPY . /go/src

# Ensure dialout group exists and create user native to the build platform
RUN apk add --no-cache git && \
    (grep -q ^dialout: /etc/group || addgroup -g 20 -S dialout) && \
    echo "wombatt:x:1001:20::/:/bin/nologin" >> /etc/passwd

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG VERSION
ARG DATE
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    cd /go/src && \
    export GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT#v} && \
    CGO_ENABLED=0 go build -ldflags "-s -w -X wombatt/cmd.Version=${VERSION} -X wombatt/cmd.BuildDate=${DATE}" -o wombatt

FROM --platform=${TARGETPLATFORM} ${BASE}
COPY --from=build /etc/group /etc/group
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /go/src/wombatt /wombatt
USER wombatt:dialout
ENTRYPOINT ["/wombatt"]
