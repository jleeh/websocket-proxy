FROM golang:1.16-alpine as builder

RUN apk add --no-cache ca-certificates git

WORKDIR /go/src/github.com/jleeh/websocket-proxy
ADD . .
RUN cd cli && go build -o websocket-proxy

# Copy executable to a fresh container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /go/src/github.com/jleeh/websocket-proxy/cli/websocket-proxy /usr/local/bin/

EXPOSE 8080
ENTRYPOINT ["websocket-proxy"]
