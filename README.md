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

In the workspace, generate self-signed client/server certificates and keys. In the following examples, the files
will be stored under sub-directories of `data`. Create that directory, if needed:

```
mkdir -p data
```
 
First, be your own certificate authority by filling out the prompts when running:

```
docker run -it -v $(pwd)/data/ca:/ca itzg/cert-helper \
  init
```

NOTE: if you forget the pass phrase for the ca.pem, just remove `data/ca` and `data/server-certs` and re-run init.

With CA powers, create a server certificate/key replacing `localhost` and `127.0.0.1` for your setup:

```
docker run -it -v $(pwd)/data/ca:/ca -v $(pwd)/data/server-certs:/certs itzg/cert-helper \
  create -server -cn localhost -alt IP:127.0.0.1
```

When running the `serve` command, your CA certificate will be specified by setting the environment variable `DEV_CA`
to the location `data/ca/ca.pem`.

Before starting the endpoint, place any additional zones->agents->checks under `contrib/endpoint-agents` (if using the example config), then
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

### Generating mocks

[GoMock](https://github.com/golang/mock) is currently in use for unit testing where an interface needs to be mocked out to minimize scope of testing,
system impact, etc. To generate a mock of an interface located in this repository first install mockgen

```
go get github.com/golang/mock/mockgen
```

With `$GOPATH/bin` in your `PATH`, run 

```
mockgen -package=PKG -destination=PKG/obj_mock_test.go github.com/racker/rackspace-monitoring-poller/PKG Object
```

where `PKG` is the sub-package that contains `Object` to mock. The file itself is named with `obj` as the snakecase
normalization of `Object`.

Please note that due to https://github.com/golang/mock/issues/30 you may need to manually scrub the imports of 
the generated file.