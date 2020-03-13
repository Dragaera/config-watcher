FROM golang:1.14-alpine AS GOLANG

RUN apk add git

WORKDIR /go/src/config-watcher

COPY . .
RUN go get -d -v ./...
RUN go build -v


FROM alpine:3.9

LABEL maintainer="Michael Senn <michael@morrolan.ch>"

COPY --from=GOLANG /go/src/config-watcher/config-watcher /usr/local/bin/
RUN chmod +x /usr/local/bin/config-watcher

ENTRYPOINT ["/usr/local/bin/config-watcher"]

