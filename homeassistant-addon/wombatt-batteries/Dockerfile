ARG BUILD_FROM

FROM gonzalomono/wombatt:v0.5.1 as source

FROM $BUILD_FROM
COPY --from=source /wombatt /usr/bin/wombatt
COPY run.sh /
RUN chmod a+x /run.sh
RUN apk add --no-cache ca-certificates jq

ENTRYPOINT [ "/run.sh" ]
