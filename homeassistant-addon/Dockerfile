ARG BUILD_FROM
FROM ${BUILD_FROM} as build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /wombatt -ldflags "-s -w" ./main.go

FROM alpine:latest
COPY --from=build /wombatt /usr/bin/wombatt
RUN apk add --no-cache ca-certificates
RUN addgroup -S wombatt && adduser -S wombatt -G wombatt
RUN addgroup wombatt dialout
USER wombatt
ENTRYPOINT ["/usr/bin/wombatt"]
