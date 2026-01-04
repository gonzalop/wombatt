ARG BUILD_FROM
FROM golang:alpine as build
WORKDIR /src
RUN apk add --no-cache git
RUN git clone https://github.com/gonzalop/wombatt /src/wombatt
WORKDIR /src/wombatt
RUN CGO_ENABLED=0 go build -o /wombatt -ldflags "-s -w" ./main.go

FROM golang:alpine
COPY --from=build /wombatt /usr/bin/wombatt
RUN apk add --no-cache ca-certificates
RUN addgroup -S wombatt && adduser -S wombatt -G wombatt
RUN addgroup wombatt dialout
USER wombatt
ENTRYPOINT ["/usr/bin/wombatt"]
