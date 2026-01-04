FROM gonzalomono/wombatt:latest as source

FROM alpine:latest
COPY --from=source /wombatt /usr/bin/wombatt
RUN apk add --no-cache ca-certificates
RUN addgroup -S wombatt && adduser -S wombatt -G wombatt
RUN addgroup wombatt dialout
USER wombatt
ENTRYPOINT ["/usr/bin/wombatt"]
