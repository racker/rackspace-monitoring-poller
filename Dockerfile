FROM golang:1.9.2-alpine

ARG DEP_VERSION=0.5.0

RUN apk update && apk upgrade && \
    apk add --no-cache --virtual .build-deps \
        bash git openssh make openssl && \
    rm -rf /var/cache/apk/*

ADD https://github.com/golang/dep/releases/download/v$DEP_VERSION/dep-linux-amd64 /usr/local/bin/dep
RUN chmod +x /usr/local/bin/dep

WORKDIR $GOPATH/src/github.com/racker/rackspace-monitoring-poller

RUN openssl req \
    -new \
    -newkey rsa:2048 \
    -nodes \
    -keyout key.pem \
    -x509 \
    -days 365 \
    -out cert.pem \
    -subj "/C=US/ST=Texas/L=Rackspace/O=Dis/CN=www.example.com"

COPY Gopkg.* $GOPATH/src/github.com/racker/rackspace-monitoring-poller/
RUN dep ensure -vendor-only
COPY . $GOPATH/src/github.com/racker/rackspace-monitoring-poller
RUN ls -l
RUN go build

EXPOSE 55000

ENTRYPOINT ["./rackspace-monitoring-poller"]

CMD ["serve", "--config", "contrib/local-docker-endpoint.cfg", "--insecure", "--debug"]
