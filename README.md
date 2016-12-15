[![Build Status](https://travis-ci.org/racker/rackspace-monitoring-poller.svg?branch=master)](https://travis-ci.org/racker/rackspace-monitoring-poller)
[![Go Documentation](https://godoc.org/github.com/racker/rackerspace-monitoring-poller?status.svg)](https://godoc.org/github.com/racker/rackspace-monitoring-poller)
[![Coverage Status](https://coveralls.io/repos/github/racker/rackspace-monitoring-poller/badge.svg?branch=master)](https://coveralls.io/github/racker/rackspace-monitoring-poller?branch=master)

## Prepare your workspace

In order to comply with Go packaging structure, be sure to clone this repo 
into the path `$GOPATH/src/github.com/racker/rackspace-monitoring-poller`

Get external dependencies before building/running:

```
glide install
go build
```

## Running Simple Endpoint Server for development

In the workspace, generate self signed certificate and private key:
 
```
openssl req -newkey rsa:2048 -nodes -keyout key.pem -x509 -days 365 -out cert.pem
```

First, place any additional zones->agents->checks under `contrib/endpoint-agents` (if using the example config), then
start the Endpoint server and Poller server.

In window #1:

    ./rackspace-monitoring-poller endpoint --config contrib/endpoint-config.json  --debug
    
In window #2:

    ./rackspace-monitoring-poller serve --config contrib/local-endpoint.cfg  --debug
    
## Development-time Documentation

If you are adding or modifying documentation comments, viewing the godoc's locally will be very helpful. 
The [godoc tool documentation](https://godoc.org/golang.org/x/tools/cmd/godoc) shows several ways to
run it, but the easiest is to run

    godoc -http=:6060
    
With that running, open your browser to [http://localhost:6060/pkg/github.com/racker/rackspace-monitoring-poller/]().

As you save file changes, just refresh the browser page to pick up the documentation changes.

