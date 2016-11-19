[![CircleCI](https://circleci.com/gh/racker/rackspace-monitoring-poller.svg?style=svg)](https://circleci.com/gh/racker/rackspace-monitoring-poller)
[![Go Documentation](https://godoc.org/github.com/racker/rackerspace-monitoring-poller?status.svg)](https://godoc.org/github.com/racker/rackspace-monitoring-poller)

## Prepare your workspace

In order to comply with Go packaging structure, be sure to clone this repo 
into the path `$GOPATH/src/github.com/racker/rackspace-monitoring-poller`

Get external dependencies before building/running:

```
glide install
go build
```

## Documentation

If you are adding new checks or hostinfo queries, viewing the godoc's will be helpful. 
The [main documentation](https://godoc.org/golang.org/x/tools/cmd/godoc) shows several ways to
run it, but the easiest is to run

    godoc -http=:6060
    
With that running, open your browser to http://localhost:6060/pkg/github.com/racker/rackspace-monitoring-poller/