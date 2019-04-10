FROM golang:1.12-alpine as builder

RUN apk add --update --no-cache git gcc musl-dev make
RUN go get github.com/Pixep/crowlet/cmd/crowlet

WORKDIR /go/src/github.com/Pixep/crowlet
RUN make tests

WORKDIR /go/src/github.com/Pixep/crowlet/cmd/crowlet
RUN CGO_ENABLED=0 go build -a -ldflags '-extldflags "-static"' -o /opt/bin/crowlet .

FROM centurylink/ca-certs

COPY --from=builder /opt/bin/crowlet /opt/bin/crowlet

ENTRYPOINT ["/opt/bin/crowlet"]
