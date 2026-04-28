ARG BUILD_ARCH=amd64
ARG WOMBATT_IMAGE=docker.io/gonzalomono/wombatt:v0.6.0

FROM ghcr.io/home-assistant/${BUILD_ARCH}-base:latest

# Re-declare ARG so it's available after FROM
ARG WOMBATT_IMAGE
COPY --from=${WOMBATT_IMAGE} /wombatt /usr/bin/wombatt

RUN apk add --no-cache ca-certificates jq

# Create wombatt user and add to dialout group for hardware access
RUN (grep -q ^dialout: /etc/group || addgroup -g 20 -S dialout) && \
    adduser -D -u 1001 -G dialout -s /bin/nologin wombatt

COPY run.sh /
RUN chmod a+x /run.sh

USER wombatt
ENTRYPOINT [ "/run.sh" ]
