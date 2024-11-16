FROM golang:1.23-alpine AS builder

RUN apk add --update --no-cache git gcc musl-dev make

ARG MODULE_PATH=${GOPATH}/src/github.com/Pixep/crowlet

COPY . $MODULE_PATH
WORKDIR $MODULE_PATH
RUN make build-static \
 && mkdir -p /opt/bin \
 && mv ./crowlet /opt/bin/crowlet

FROM golang:1.23-alpine

COPY --from=builder /opt/bin/crowlet /opt/bin/crowlet

ENTRYPOINT ["/opt/bin/crowlet"]
