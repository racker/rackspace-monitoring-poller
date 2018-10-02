FROM golang:1.11-alpine AS build

ARG DEP_VERSION=0.5.0

RUN apk update && apk upgrade && \
    apk add --no-cache --virtual .build-deps \
        bash git openssh make openssl && \
    rm -rf /var/cache/apk/*

ADD https://github.com/golang/dep/releases/download/v$DEP_VERSION/dep-linux-amd64 /usr/local/bin/dep
RUN chmod +x /usr/local/bin/dep

WORKDIR $GOPATH/src/github.com/racker/rackspace-monitoring-poller

COPY Gopkg.* $GOPATH/src/github.com/racker/rackspace-monitoring-poller/
RUN dep ensure -vendor-only
COPY . $GOPATH/src/github.com/racker/rackspace-monitoring-poller

RUN CGO_ENABLED=0 go build
RUN cp rackspace-monitoring-poller /usr/local/bin

FROM ubuntu

COPY --from=build /usr/local/bin/rackspace-monitoring-poller /usr/local/bin/
EXPOSE 55000

ENTRYPOINT ["/usr/local/bin/rackspace-monitoring-poller"]

CMD ["serve", "--debug"]
