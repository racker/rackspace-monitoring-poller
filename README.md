[![Build Status](https://travis-ci.org/racker/rackspace-monitoring-poller.svg?branch=master)](https://travis-ci.org/racker/rackspace-monitoring-poller)
[![Go Documentation](https://godoc.org/github.com/racker/rackerspace-monitoring-poller?status.svg)](https://godoc.org/github.com/racker/rackspace-monitoring-poller)
[![Coverage Status](https://coveralls.io/repos/github/racker/rackspace-monitoring-poller/badge.svg?branch=master)](https://coveralls.io/github/racker/rackspace-monitoring-poller?branch=master)
[![GitHub release](https://img.shields.io/github/release/racker/rackspace-monitoring-poller.svg)](https://github.com/racker/rackspace-monitoring-poller/releases)

# Installing and Running a Poller Instance

_NOTE: These instructions are only a brief, informal procedure. Details such as locating the `monitoring_token`
will be described elsewhere._

From [the latest release](https://github.com/racker/rackspace-monitoring-poller/releases/latest), 
download either the "deb" package or one of the standalone binaries listed after the debian package.

## APT

**NOTE:** the following "apt repo" procedure is only supported on [upstart](http://upstart.ubuntu.com/) based systems
such as Ubuntu 14.04 Trusty.

* Update your packages to latest

```
apt-get update
```

* Add public key

```
curl https://monitoring.api.rackspacecloud.com/pki/agent/linux.asc | sudo apt-key add -
```

* Add cloudmonitoring repo

```
echo "deb [arch=amd64] http://stable.poller.packages.cloudmonitoring.rackspace.com/debian cloudmonitoring main" > /etc/apt/sources.list.d/cloudmonitoring.list
```

* Update packages to include the new repo

```
apt-get update
```

* Install the poller

```
apt-get install rackspace-monitoring-poller -y
```

Proceed to configuring your poller.


## DEB package installation directly

**NOTE:** there are two deb packages provided. The one with the `-systemd` qualifier should be installed on systems that
use [systemd](https://freedesktop.org/wiki/Software/systemd/) for the init system, such as Ubuntu 15.04 Vivid and greater. 
The other deb package should be installed on systems that use [upstart](http://upstart.ubuntu.com/).

Install the package with `dpkg -i rackspace-monitoring-poller_*.deb`. 

Proceed to configuring your poller.


## Configuration for debian install

Adjust these configuration files

* `/etc/rackspace-monitoring-poller.cfg`
  * Configure `monitoring_token` by uncommenting the line and replacing the `TOKEN` placeholder
  * Configure `monitoring_private_zones` by uncommenting the line and replacing the `ZONE` placeholder
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

# Configuration options

Parameter | Type | Description
----------|------|------------
monitoring_id | string | Optional. Specifies a user-provided id string that identifies this agent to the monitoring services.
monitoring_token | string | Required. This token gives the poller access to the monitoring services for an account. See [the section below](#locate-or-create-a-monitoring-token).
monitoring_private_zones | string | Required. A private monitoring zone that this poller supports. More than one poller can and should support a given monitoring zone. See [the section below](#allocate-a-private-zone).
monitoring_endpoints | comma-delimited sets of ip:port values | Optional. Provides a series of endpoint IP addresses for the agent to connect to instead of the default endpoint addresses.
monitoring_snet_region | string | Optional. This option tells the poller to connect to the agent endpoints over the Rackspace ServiceNet (instead of over the public Internet)
monitoring_proxy_url | url | Optional. Provides a URL string to a HTTP Proxy service that supports the CONNECT command.
prometheus_uri | url | Optional. A Prometheus gateway where internal poller operation metrics can be sent.
statsd_endpoint | host:port | Optional. The address of a UDP statsd receiver that will receive all of the same metrics sent into the Rackspace agent endpoint.

# Preparing your Rackspace Monitoring account

The Rackspace Monitoring Poller is primarily intended to be used as part of 
your overall [Rackspace Monitoring](https://developer.rackspace.com/docs/rackspace-monitoring/v1/getting-started/concepts/) setup. 
You will need to ensure the following items are prepared before you can effectively use one or more poller instances:

- [ ] [Enable private zones feature](#enable-private-zones-feature)
- [ ] [Locate or create a monitoring token](#locate-or-create-a-monitoring-token)
- [ ] [Allocate a private zone](#allocate-a-private-zone)
- [ ] [Creating remote checks to run in a private zone](#creating-remote-checks-to-run-in-a-private-zone)

The examples below show use of the [Rackspace Monitoring REST API](https://developer.rackspace.com/docs/rackspace-monitoring/v1/),
but you can also use the [Rackspace Monitoring CLI](https://support.rackspace.com/how-to/getting-started-with-rackspace-monitoring-cli/).

Using the REST API requires obtaining a token ID and knowing your tenant ID. 
[This page in the developer guide](https://developer.rackspace.com/docs/rackspace-monitoring/v1/getting-started/send-request-ovw/) explains this procedure.
The examples exclude the service access endpoint and path prefix for brevity. 
The actual calls should be invoked against `https://monitoring.api.rackspacecloud.com/v1.0/$TENANT_ID`

## Enable private zones feature

Private zone monitoring is currently a limited-availability feature, so 
please contact [Rackspace Support](https://www.rackspace.com/support) to request this feature to be enabled on your
account.

## Locate or create a monitoring token

In order for your poller instance to authenticate against your Rackspace Monitoring account you will either need to
obtain an existing token from your account or create a new one. _The latter is a good practice in order to delineate
the authentication of your private zone pollers independent of your on-node agents._

**NOTE:** the token is traditionally referred to as an **agent** token, but the concept is interchangeable with **monitoring** tokens. 

You can [list existing tokens](https://developer.rackspace.com/docs/rackspace-monitoring/v1/api-reference/agent-token-operations/#list-agent-tokens)
by invoking:

```text
GET /agent_tokens
```

and using the `token` from the appropriate entry in the response list.

You can also [create a new token](https://developer.rackspace.com/docs/rackspace-monitoring/v1/api-reference/agent-token-operations/#create-an-agent-token)
by invoking:

```text
POST /agent_tokens
```

You may choose any `label` that you would like.

## Allocate a private zone

A private zone is allocated by invoking

```text
POST /monitoring_zones
```

The JSON object payload contains the following attributes:

Name | Description | Validation
-----|-------------|-----------
label | A label of your choosing | (Required) Non-empty string
metadata | Arbitrary key/value pairs | (Optional) Object of string values, up to 256 keys. Values and keys must be less than 255 characters.

A response status of `201` indicates that the zone was successfully created. 
The `X-Object-ID` response header contains the ID that is used in the poller configuration file, above, and the check
creation, below. 
The full details of the newly created zone can be retrieving from the URL in the `Location` response header.

## Creating remote checks to run in a private zone

Creating a check for a private zone uses the existing API to [create a check](https://developer.rackspace.com/docs/rackspace-monitoring/v1/api-reference/check-operations/#create-a-check).
Be sure to pass the private zone ID in the `monitoring_zones_poll` array attribute.

The following check types are currently supported:
* [remote.tcp](https://developer.rackspace.com/docs/rackspace-monitoring/v1/tech-ref-info/check-type-reference/#remote-tcp)
* [remote.http](https://developer.rackspace.com/docs/rackspace-monitoring/v1/tech-ref-info/check-type-reference/#remote-http)
* [remote.ping](https://developer.rackspace.com/docs/rackspace-monitoring/v1/tech-ref-info/check-type-reference/#remote-ping)
