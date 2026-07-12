ARG BUILD_ARCH=amd64
ARG BUILD_FROM=ghcr.io/home-assistant/${BUILD_ARCH}-base:latest
ARG BUILD_VERSION=0.7.0

FROM docker.io/gonzalomono/wombatt:v${BUILD_VERSION} AS source
FROM ${BUILD_FROM}

COPY --from=source /wombatt /usr/bin/wombatt

RUN apk add --no-cache ca-certificates jq

# Create wombatt user and add to dialout group for hardware access
RUN (grep -q ^dialout: /etc/group || addgroup -g 20 -S dialout) && \
    adduser -D -u 1001 -G dialout -s /bin/nologin wombatt

COPY run.sh /
RUN chmod a+x /run.sh

ENTRYPOINT [ "/run.sh" ]
