FROM golang:1.9-alpine3.7 as builder

COPY . /go/src/github.com/flaccid/sitemap-crawler

WORKDIR /go/src/github.com/flaccid/sitemap-crawler/cmd/smapcrawl

RUN apk add --update --no-cache git gcc musl-dev && \
    go get ./... && \
    CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o /opt/bin/smapcrawl .

FROM centurylink/ca-certs

COPY --from=builder /opt/bin/smapcrawl /opt/bin/smapcrawl

ENTRYPOINT ["/opt/bin/smapcrawl"]
