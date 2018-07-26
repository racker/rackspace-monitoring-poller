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

With this repository cloned into your `$GOPATH`, install required tools and populate the external dependencies 
before building/running:

```
make prep
```

Finally, you can build an instance of the `rackspace-monitoring-poller` executable using:

```
go build
```

## Building like CI does

If you would like to build like the CI build does it, then you can perform a `make build` invocation inside
of a Go container:

```bash
docker run --rm \
  -v ${PWD}:/go/src/github.com/racker/rackspace-monitoring-poller \
  -w /go/src/github.com/racker/rackspace-monitoring-poller \
  golang:1.10.2 \
  make build
```

Since it volume-attaches the workspace, the result of the build will be placed in the `build` sub-directory.

## Environment required for RPM packaging

The RPM packaging in the [Makefile](Makefile) requires specific environment variables be set:
* `PACKAGING=rpm`
* `DISTRIBUTION` set to the appropriate value, such as `centos` or `redhat`. The value needs to conform
  to the naming convention used in the [agent repo](http://stable.packages.cloudmonitoring.rackspace.com/)
* `RELEASE` set to the appropriate value relative to the `DISTRIBUTION`. For example, `centos` and `redhat` tend to
  use the single digit convetion like `7`. Again, the choice of value needs to align with the convention used in
  the [agent repo](http://stable.packages.cloudmonitoring.rackspace.com/).
* `GPG_KEYID` set to indicate the GPG key to use for signing the rpm and the yum repository metadata

Other than the above environment variables, the build environment needs `rpm-build` and `createrepo` packages
installed above and beyond the usual development packages of `git` and `make`. 
A [Dockerfile is available](contrib/docker-builder/Dockerfile.centos7) that comes with those packages.

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

## Publishing a pre-release

A pre-release is actually mostly the same process for the git and Github portion. We just don't typically publish
to the apt repository.

1. [In Travis CI](https://travis-ci.org/racker/rackspace-monitoring-poller/builds) make sure that your local clone
   is at the same commit that has successfully built.
2. Locate [the most recent release](https://github.com/racker/rackspace-monitoring-poller/releases/latest)
3. Determine the pre-release version by adding a 4th version slot to that latest release. 
   For example, a pre-release of `0.2.32` would actually be `0.2.31.1` since `0.2.31` is the version actually published 
   at that point. If more pre-releases are needed, just bump that 4th version slot.
4. Execute a `git tag -s <VERSION> -m <MSG>` to create a signed+annotated tag
5. Push that tag, such as `git push origin <VERSION>`
6. Wait for [Travis CI to finish the tagged build](https://travis-ci.org/racker/rackspace-monitoring-poller)
7. Assuming that succeeded, click 'Edit' on 
   [the release it published](https://github.com/racker/rackspace-monitoring-poller/releases)
   and enable the "This is a pre-release" option towards the bottom of the page.
8. Click "Update release" and you have now published and designated a pre-release of the poller
