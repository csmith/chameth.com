---
date: 2017-12-17
title: DNS-over-TLS on the EdgeRouter Lite
description: Installing and configuring Unbound on an Edgerouter Lite to enable encrypted DNS requests.
tags: [security, sysadmin]
format: long
permalink: /dns-over-tls-on-edgerouter-lite/

resources:
  - src: edgerouter.jpg
    name: An EdgeRouter Lite

opengraph:
  image: /dns-over-tls-on-edgerouter-lite/edgerouter.jpg
---

{% figure "left" "An EdgeRouter Lite" %}

DNS-over-TLS is a fairly recent specification described in
[RFC7858](https://tools.ietf.org/html/rfc7858), which enables DNS clients to
communicate with servers over a TLS (encrypted) connection instead of requests
and responses being sent in plain text. I won't ramble on about why it's a good
thing that your ISP, government, or neighbour can't see your DNS requests...

I use an [EdgeRouter Lite](https://www.ubnt.com/edgemax/edgerouter-lite/) from
Ubiquiti Networks at home, and recently configured it to use DNS-over-TLS for
all DNS queries. Here's how I did it.

### Installing unbound

Out of the box, the ERL uses `dnsmasq` to service DNS requests from local
clients. To get DNS-over-TLS support I switched to using
[Unbound](https://unbound.net/), an open source DNS resolver with support
for many modern features such as DNSSEC and DNS-over-TLS.

<!--more-->

{% update "2017-05-31" %}
Before installing unbound, if you haven't done so before, you will need to enable the apt
repositories as described in the [Ubiquiti help center](https://help.ubnt.com/hc/en-us/articles/205202560-EdgeRouter-Add-other-Debian-packages-to-EdgeOS).
Thanks to [@ozaed](https://twitter.com/ozaed/status/960615650489233408) for the tip.
{% endupdate %}

Installing unbound on the ERL is a simple case of SSHing in, and then:

```text
sudo apt-get update
sudo apt-get install unbound
```

And then configuring the ERL to use the new local resolver for DNS requests,
turn off dnsmasq, and and tell DHCP clients to send DNS requests to it
(obviously substituting network names and subnets as appropriate):

```text
set system name-server 127.0.0.1
set service dhcp-server shared-network-name lan1 subnet 192.168.1.0/24 dns-server 192.168.1.1
set service dhcp-server use-dnsmasq disable
set service dns
```

At this point DNS should still work, but Unbound will still be sending requests
out in plain text.

### Configuration

The unbound configuration lives in `/etc/unbound/unbound.conf`, here's a basic
example that I use to enable DNS-over-TLS:

```yaml
# Unbound configuration file for Debian.
#
# See the unbound.conf(5) man page.
#
# See /usr/share/doc/unbound/examples/unbound.conf for a commented
# reference config file.

server:
    auto-trust-anchor-file: "/var/lib/unbound/root.key"
    verbosity: 1
    interface: 0.0.0.0
    interface: ::0
    port: 53
    do-ip4: yes
    do-ip6: yes
    do-udp: yes
    do-tcp: yes
    access-control: 192.168.0.0/16 allow
    access-control: 127.0.0.0/8 allow
    access-control: 10.0.0.0/8 allow
    root-hints: "/var/lib/unbound/root.hints"

    hide-identity: yes
    hide-version: yes
    harden-glue: yes
    harden-dnssec-stripped: yes

    cache-min-ttl: 900
    cache-max-ttl: 14400
    prefetch: yes
    rrset-roundrobin: yes
    ssl-upstream: yes
    use-caps-for-id: yes

    private-address: 192.168.0.0/16
    private-address: 172.16.0.0/12
    private-address: 10.0.0.0/8

    logfile: "/var/lib/unbound/unbound.log"
    verbosity: 0
    val-log-level: 3

forward-zone:
    name: "."
    forward-addr: 9.9.9.9@853
```

Notice the server directive `ssl-upstream`, and that the forward zone specifies
the [quad9](https://www.quad9.net/) resolver on its TLS port (853).

#### DNSSEC

To enable Unbound to validate DNSSEC signatures, we need to provide it with
some information about the root nameservers that we trust. First, download the
list of root nameservers to the `root-hints` file specified in the unbound
config:

```text
wget ftp://FTP.INTERNIC.NET/domain/named.cache -O /var/lib/unbound/root.hints
```

Then we need to add the root keys to the `auto-trust-anchor-file`. The trust
anchor at the time of writing is below, but you can get the latest values from
the [IANA](https://data.iana.org/root-anchors/).

```text
.       172800  IN      DNSKEY  257 3 8 AwEAAaz/tAm8yTn4Mfeh5eyI96WSVexTBAvkMgJzkKTOiW1vkIbzxeF3+/4RgWOq7HrxRixHlFlExOLAJr5emLvN7SWXgnLh4+B5xQlNVz8Og8kvArMtNROxVQu
.       172800  IN      DNSKEY  257 3 8 AwEAAagAIKlVZrpC6Ia7gEzahOR+9W29euxhJhVVLOyQbSEW0O8gcCjFFVQUTf6v58fLjwBd0YI0EzrAcQqBGCzh/RStIoO8g0NfnfL2MTJRkxoXbfDaUeVPQuY
```

### Redirecting unencrypted requests

I have a slew of devices on my network that, over time, I have configured to
use 8.8.8.8 as a DNS server. They're not going to care about the DHCP reply,
and I don't really feel like going around checking every weird and wonderful
internet-connected device in the house, so I decided to just intercept requests
to 8.8.8.8 and send them to Unbound. A simple NAT rule does the trick:

```text
set service nat rule 1 description "HonestDNS Redirect"
set service nat rule 1 destination address 8.8.8.8
set service nat rule 1 destination port 53
set service nat rule 1 inbound-interface eth0
set service nat rule 1 inside-address address 192.168.1.1
set service nat rule 1 inside-address port 53
set service nat rule 1 log disable
set service nat rule 1 protocol tcp_udp
set service nat rule 1 type destination
```

You could of course redirect any traffic to port 53, but that would prevent you
from explicitly querying any other DNS server. By just intercepting traffic to
8.8.8.8 I'm taking care of the vast majority of my statically configured
devices, and can still issue manual queries to other resolvers when needed.

### Validating

To check that everything is working, you can use `tcpdump` on the router to
inspect packets on the WAN interface directed at port 53:

```text
sudo tcpdump -Xi eth0 port 53
```

You should, hopefully, not see anything.
