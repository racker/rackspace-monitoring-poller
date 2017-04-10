[![Build Status](https://travis-ci.org/racker/rackspace-monitoring-poller.svg?branch=master)](https://travis-ci.org/racker/rackspace-monitoring-poller)
[![Go Documentation](https://godoc.org/github.com/racker/rackerspace-monitoring-poller?status.svg)](https://godoc.org/github.com/racker/rackspace-monitoring-poller)
[![Coverage Status](https://coveralls.io/repos/github/racker/rackspace-monitoring-poller/badge.svg?branch=master)](https://coveralls.io/github/racker/rackspace-monitoring-poller?branch=master)
[![GitHub release](https://img.shields.io/github/release/racker/rackspace-monitoring-poller.svg)](https://github.com/racker/rackspace-monitoring-poller/releases)

# Installing and Running a Poller Instance

_NOTE: These instructions are only a brief, informal procedure. Details such as locating the `monitoring_token`
will be described elsewhere._

From [the latest release](https://github.com/racker/rackspace-monitoring-poller/releases/latest), 
download either the "deb" package or one of the standalone binaries listed after the debian package.

## DEB package

Install the package with `dpkg -i rackspace-monitoring-poller_*.deb`. 

Adjust these configuration files

* `/etc/rackspace-monitoring-poller.cfg`
  * Configure `monitoring_token` and `monitoring_private_zones`
* `/etc/default/rackspace-monitoring-poller`
  * Enable the service by setting `ENABLED=true`

Start the poller service using:

```bash
sudo initctl start rackspace-monitoring-poller 
```

## Standalone Binary

Use `chmod +x` to mark the binary executable. For convenience, you can rename it to `rackspace-monitoring-poller`,
which is what the following instructions will reference.

Using [this sample](contrib/remote.cfg) place the poller's configuration file in a location of your choosing or
the default location of `/etc/rackspace-monitoring-poller.cfg`.

Start the poller with the default configuration file location

```bash
./rackspace-monitoring-poller serve
```

or specify a custom confiuration file location:

```bash
./rackspace-monitoring-poller serve --config custom.cfg
```

# Development Notes

## Prepare your workspace

If you haven't already, setup your `$GOPATH` in the typical way [described here](https://golang.org/doc/code.html#GOPATH).

Similarly, if you haven't already installed [Glide, package management for Go](https://glide.sh/), then visit their
page and follow the "Get Glide" step.

In order to comply with Go's packaging structure, be sure to clone this repo
into the path `$GOPATH/src/github.com/racker/rackspace-monitoring-poller`, such as:

```
mkdir -p $GOPATH/src/github.com/racker
cd $GOPATH/src/github.com/racker
git clone https://github.com/racker/rackspace-monitoring-poller.git
```

With this repository cloned into your `$GOPATH`, populate the external dependencies before building/running:

```
glide install
```

or setup/upgrade Glide and install dependencies with make:

```
make prep
```

Finally, you can build an instance of the `rackspace-monitoring-poller` executable using:

```
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

#### For docker, you can simply:

```
docker build -t racker-poller .
docker run -ti --rm -v $(pwd)/contrib:/config racker-poller
```

#### If running locally, then

In window #1:

    ./rackspace-monitoring-poller endpoint --config contrib/endpoint-config.json  --debug

In window #2:

    export DEV_CA=data/ca/ca.pem
    ./rackspace-monitoring-poller serve --config contrib/local-endpoint.cfg  --debug

## Development-time Documentation

If you are adding or modifying documentation comments, viewing the godoc's locally will be very helpful.
The [godoc tool documentation](https://godoc.org/golang.org/x/tools/cmd/godoc) shows several ways to
run it, but the easiest is to run

    godoc -http=:6060

With that running, open your browser to [http://localhost:6060/pkg/github.com/racker/rackspace-monitoring-poller/]().

As you save file changes, just refresh the browser page to pick up the documentation changes.

### Generating mocks

To regenerate mocks automatically, run:

```
make generate-mocks
```

[GoMock](https://github.com/golang/mock) is currently in use for unit testing where an interface needs to be mocked out to minimize scope of testing,
system impact, etc. To generate a mock of an interface located in this repository first install mockgen

```
go get github.com/golang/mock/mockgen
```

With `$GOPATH/bin` in your `PATH`, run

```
mockgen -source={Pkg}/{InterfaceFile}.go -package={Pkg} -destination={Pkg}/{InterfaceFile}_mock_test.go
```

where `Pkg` is the sub-package that contains one or more interfaces in `{InterfaceFile}.go` to mock.
It will write mocks of those interfaces to an adjacent file `{InterfaceFile}_mock_test.go` where those
mocks will also reside in the package `Pkg`.

An example use for the interfaces in `poller/poller.go` is:

```
mockgen -source=poller/poller.go -package=poller -destination=poller/poller_mock_test.go
```

Don't forget to re-run `mockgen` when a mocked interface is altered since test-time compilation errors will result
from incomplete interface implementations.

Please note that due to https://github.com/golang/mock/issues/30 you may need to manually scrub the imports of
the generated file.
