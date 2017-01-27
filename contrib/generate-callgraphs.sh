#!/bin/sh

go get -u github.com/TrueFurby/go-callvis

packages=
packages="$packages poller"
packages="$packages endpoint"

for pkg in $packages; do
    echo "Generating docs/${pkg}_callgraph.*"

    $GOPATH/bin/go-callvis -focus ${pkg} github.com/racker/rackspace-monitoring-poller > docs/${pkg}_callgraph.dot

    if which dot > /dev/null; then dot -Tpng -o docs/${pkg}_callgraph.png docs/${pkg}_callgraph.dot ; fi
done