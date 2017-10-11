#!/bin/sh

curl https://monitoring.api.rackspacecloud.com/pki/agent/linux.asc | sudo apt-key add -

echo "deb [arch=amd64] http://stable.poller.packages.cloudmonitoring.rackspace.com/debian cloudmonitoring main" > /etc/apt/sources.list.d/cloudmonitoring.list

apt-get update

