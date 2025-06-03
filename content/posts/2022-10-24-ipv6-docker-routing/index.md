---
date: 2022-10-24
title: Adventures in IPv6 routing in Docker
description: Guest starring musl, as the evil villain
tags: [docker, networks, research]
permalink: /ipv6-docker-routing/
---

One of the biggest flaws in Docker's design is that it wasn't created with IPv6 in mind. Out of the box Docker assigns
each container a private IPv4 address, and they won't be able to reach IPv6-only services. While incoming connections
might work, the containers won't know the correct remote IP address which can cause problems for some applications.
This situation is obviously suboptimal in the current day and age. It's a bit like not supporting HTTPS on a website --
you might not have any issues because of it immediately, but you're fighting against the currents of progress and are
making life worse for your users.

Thankfully, it's now relatively easy to make Docker behave a lot nicer. The
[docker-ipv6nat](https://github.com/robbertkl/docker-ipv6nat) project has been around since 2016, and uses an IPv6
overlay network and some iptables magic to route traffic to and from containers in a sensible fashion. It uses NAT
to emulate the behaviour Docker employs for IPv4 traffic; while using NAT with IPv6 is an anathema, I think it makes
sense for containers. You could give each container a publicly routable IPv6 address, but that brings with it a lot
of headaches: you're basically going to be forced to implement service discovery and some kind of DNS management to
deal with the fact that your containers will be popping up on randomly assigned IP addresses. That is completely
overkill for people running a small number of services on one or two physical boxes; and if it's not overkill for you
then you're probably already looking at more complicated orchestration solutions like Kubernetes.

More recently, similar functionality has been built into the Docker daemon itself. You can now
[edit the config file to enable ipv6](https://docs.docker.com/config/daemon/ipv6/) and each container will be
assigned an address in the range specified when it uses the default bridge network. This gives more-or-less the same
functionality of docker-ipv6nat -- you lose a little flexibility as you can't disable IPv6 on the default bridge, but
that's a very worthy trade for having the functionality built-in.

So far this all seems very simple. Hardly worthy of being called an "adventure". Enter stage left: the wicked witch of
destination address selection...

<!--more-->

### Destination address selection and you

When a computer program tries to connect to an address such as `google.com`, it first resolves it in to an IP address.
That's DNS 101, but what happens if the address resolves to multiple IP addresses? For example, `google.com` resolves
to both `142.250.74.206` and `2a00:1450:4001:82b::200e`. You might just assume there's a simple
"prefer IPv4 / prefer IPv6" toggle somewhere that decides, but it's actually a lot more complicated. With IPv6, devices
are likely to have many addresses - a link-local address, a unique local address, a normal public address, various
privacy addresses, and so on. To deal with this, a number of RFCs define a series of rules that most implementations
follow. These are called the _destination address selection rules_. [RFC6724](https://www.rfc-editor.org/rfc/rfc6724)
gives the rules as:

1. Avoid unusable destinations
2. Prefer matching scope
3. Avoid deprecated addresses
4. Prefer home addresses
5. Prefer matching label
6. Prefer higher precedence
7. Prefer native transport
8. Prefer smaller scope
9. Use the longest matching prefix
10. Otherwise, leave the order unchanged

Applying these rules will re-order the list of IP addresses such that (in theory) the most likely one to work will be
first. Most of them are fairly niche rules; the ones that do the heavy lifting are rules 5 and 6, which rely on a policy
table to make their decisions. The policy table 'SHOULD' be configurable by system administrators to allow them to tweak
how traffic is routed. In the absence of an admin-provided policy table, the RFC gives the following defaults:

| Prefix          | Precedence | Label | Notes                                  |
|-----------------|------------|-------|----------------------------------------|
| `::1/128`       | 50         | 0     | Loopback address                       |
| `::/0`          | 40         | 1     | Any IPv6 address                       |
| `::ffff:0:0/96` | 35         | 4     | IPv4 addresses mapped as v6 addresses  |
| `2002::/16`     | 30         | 2     | 6-to-4 gateways                        |
| `2001::/32`     | 5          | 5     | Toredo tunnels                         |
| `fc00::/7`      | 3          | 13    | Unique Local Addresses (ULAs)          |
| `::/96`         | 1          | 3     | IPv4 compatible addresses (deprecated) |
| `fec0::/10`     | 1          | 11    | Site-local addresses (deprecated)      |
| `3ffe::/16`     | 1          | 12    | 6bone (deprecated)                     |

The policy table is a bit complex, but you can see that normal IPv6 addresses are preferred (have a higher precedence)
over IPv4 addresses, which are preferred over the various tunnels, local addresses, and various deprecated ranges.

We can see the result of this when I run `ping google.com` on a box that has a native IPv6 connection as well as a
native IPv4 connection:

```text
PING google.com(fra07s29-in-x200e.1e100.net (2a00:1450:4001:802::200e)) 56 data bytes
64 bytes from fra24s01-in-x0e.1e100.net (2a00:1450:4001:802::200e): icmp_seq=1 ttl=119 time=5.11 ms
64 bytes from fra24s01-in-x0e.1e100.net (2a00:1450:4001:802::200e): icmp_seq=2 ttl=119 time=5.28 ms
64 bytes from fra07s29-in-x200e.1e100.net (2a00:1450:4001:802::200e): icmp_seq=3 ttl=119 time=5.31 ms
^C
--- google.com ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 2003ms
rtt min/avg/max/mdev = 5.114/5.235/5.311/0.086 ms
```

The address selection in this instance has been a result of rule 6: the IPv6 address has a higher precedence (40) than
the IPv4 address (35).

However, if I run the same command in an IPv6-enabled Ubuntu container then it seems to prefer the IPv4 address:

```text
PING google.com (172.217.16.206) 56(84) bytes of data.
64 bytes from fra16s08-in-f206.1e100.net (172.217.16.206): icmp_seq=1 ttl=59 time=4.86 ms
64 bytes from fra16s08-in-f14.1e100.net (172.217.16.206): icmp_seq=2 ttl=59 time=4.86 ms
64 bytes from fra16s08-in-f14.1e100.net (172.217.16.206): icmp_seq=3 ttl=59 time=5.03 ms
^C
--- google.com ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 2003ms
rtt min/avg/max/mdev = 4.860/4.916/5.027/0.078 ms
```

If I force `ping` to use IPv6 with the `-6` flag then it works the same as on the host, but when left to its own devices
it prefers IPv4. What's going on there? The key difference between the Docker container and the host is that the host's
network interfaces have public addresses, but the container has a private IPv4 address (`172.19.0.7`) and a private
IPv6 address (`fd00:dead:beef::7`). My first reaction to this was to think "Ah yes, fc00::/7 has a lower precedence
than native IPv4, that makes sense", but that's not quite right. These are the _destination_ address selection rules;
rule 6 doesn't care about the source addresses. This is actually rule 5 at work: the label of source IPv6 address is
`13`, but the label of the destination address is `1`; meanwhile both the source and destination IPv4 addresses are `4`.
This sorts the IPv4 address before the IPv6 one, and rule 6 becomes irrelevant.

### Adjusting the policy table

Thankfully, the RFC says the policy table should be configurable by system administrators, so those of us who are doing
unorthodox things like NAT'ing IPv6 can customise the behaviour to fit our weird environments. The configuration is
done via the [/etc/gai.conf](https://man.archlinux.org/man/gai.conf.5.en) file ('gai' standing for `getaddrinfo`, the
function in the standard library responsible for dealing with all these rules). The `gai.conf` file by default (if it
exists) will likely just contain comments and examples; if there are no uncommented "label" or "precedence" lines then
the library will use its built-in defaults based on the RFC requirements.

To make our container happy, we need to cause `fd00::/8` to have the same label as a public IPv6 address. To do this we
can uncomment the default labels in `gai.conf` and then add a single extra line, like so:

```diff
 label ::1/128       0
 label ::/0          1
 label 2002::/16     2
 label ::/96         3
 label ::ffff:0:0/96 4
 label fec0::/10     5
 label fc00::/7      6
 label 2001:0::/32   7
+label fd00::/8      1
```

Because `fd00::/8` has a longer prefix than `fc00::/7` it will match our addresses and give them a label of `1`, the
same as a public IPv6 address would get. This makes rule 5 leave it alone, and the default precedence table used by
the standard library will put the IPv6 address above the IPv4 address. (They don't bother giving `fc00::/7` a separate
precedence as listed in the RFC because rule 5 would have already de-prioritised those addresses, as we discovered.)

Making this small change to the `gai.conf` file in our Ubuntu container makes it start preferring the IPv6 address
for `google.com`:

```text
PING google.com(fra16s65-in-x0e.1e100.net (2a00:1450:4001:806::200e)) 56 data bytes
64 bytes from fra15s29-in-x0e.1e100.net (2a00:1450:4001:806::200e): icmp_seq=1 ttl=118 time=5.00 ms
64 bytes from fra15s29-in-x0e.1e100.net (2a00:1450:4001:806::200e): icmp_seq=2 ttl=118 time=5.10 ms
64 bytes from fra16s65-in-x0e.1e100.net (2a00:1450:4001:806::200e): icmp_seq=3 ttl=118 time=5.08 ms
^C
--- google.com ping statistics ---
3 packets transmitted, 3 received, 0% packet loss, time 2004ms
rtt min/avg/max/mdev = 5.002/5.058/5.098/0.040 ms
```

A fairly complicated problem, but a simple fix!

### But wait, there's more...

Unfortunately, this fix won't work on containers that use alpine. And that is a _lot_ of containers. Alpine uses
[musl](https://musl.libc.org/) as its standard library, rather than the much more common
[glibc](https://www.gnu.org/software/libc/). There is very little documentation on the subject, but if you browse
the source code for musl you will see that it doesn't implement any way at all to configure the policy tables. In
[network/lookup_name.c](https://git.musl-libc.org/cgit/musl/tree/src/network/lookup_name.c?id=63402be229facae2d0de9c5943a6ed25246fd021)
we can see the sorting logic:

```c
	/* The following implements a subset of RFC 3484/6724 destination
	 * address selection by generating a single 31-bit sort key for
	 * each address. Rules 3, 4, and 7 are omitted for having
	 * excessive runtime and code size cost and dubious benefit.
	 * So far the label/precedence table cannot be customized. */
	for (i=0; i<cnt; i++) {
		int family = buf[i].family;
		int key = 0;
		struct sockaddr_in6 sa6 = { 0 }, da6 = {
			.sin6_family = AF_INET6,
			.sin6_scope_id = buf[i].scopeid,
			.sin6_port = 65535
		};
		struct sockaddr_in sa4 = { 0 }, da4 = {
			.sin_family = AF_INET,
			.sin_port = 65535
		};
		void *sa, *da;
		socklen_t salen, dalen;
		if (family == AF_INET6) {
			memcpy(da6.sin6_addr.s6_addr, buf[i].addr, 16);
			da = &da6; dalen = sizeof da6;
			sa = &sa6; salen = sizeof sa6;
		} else {
			memcpy(sa6.sin6_addr.s6_addr,
				"\0\0\0\0\0\0\0\0\0\0\xff\xff", 12);
			memcpy(da6.sin6_addr.s6_addr+12, buf[i].addr, 4);
			memcpy(da6.sin6_addr.s6_addr,
				"\0\0\0\0\0\0\0\0\0\0\xff\xff", 12);
			memcpy(da6.sin6_addr.s6_addr+12, buf[i].addr, 4);
			memcpy(&da4.sin_addr, buf[i].addr, 4);
			da = &da4; dalen = sizeof da4;
			sa = &sa4; salen = sizeof sa4;
		}
		const struct policy *dpolicy = policyof(&da6.sin6_addr);
		int dscope = scopeof(&da6.sin6_addr);
		int dlabel = dpolicy->label;
		int dprec = dpolicy->prec;
		int prefixlen = 0;
		int fd = socket(family, SOCK_DGRAM|SOCK_CLOEXEC, IPPROTO_UDP);
		if (fd >= 0) {
			if (!connect(fd, da, dalen)) {
				key |= DAS_USABLE;
				if (!getsockname(fd, sa, &salen)) {
					if (family == AF_INET) memcpy(
						sa6.sin6_addr.s6_addr+12,
						&sa4.sin_addr, 4);
					if (dscope == scopeof(&sa6.sin6_addr))
						key |= DAS_MATCHINGSCOPE;
					if (dlabel == labelof(&sa6.sin6_addr))
						key |= DAS_MATCHINGLABEL;
					prefixlen = prefixmatch(&sa6.sin6_addr,
						&da6.sin6_addr);
				}
			}
			close(fd);
		}
		key |= dprec << DAS_PREC_SHIFT;
		key |= (15-dscope) << DAS_SCOPE_SHIFT;
		key |= prefixlen << DAS_PREFIX_SHIFT;
		key |= (MAXADDRS-i) << DAS_ORDER_SHIFT;
		buf[i].sortkey = key;
	}
	qsort(buf, cnt, sizeof *buf, addrcmp);
```

We can see if the address labels match, the sort key is adjusted using the `DAS_MATCHINGLABEL` constant. But where
do the labels come from? For that we need to investigate the `policyof` func:

```c
static const struct policy {
	unsigned char addr[16];
	unsigned char len, mask;
	unsigned char prec, label;
} defpolicy[] = {
	{ "\0\0\0\0\0\0\0\0\0\0\0\0\0\0\0\1", 15, 0xff, 50, 0 },
	{ "\0\0\0\0\0\0\0\0\0\0\xff\xff", 11, 0xff, 35, 4 },
	{ "\x20\2", 1, 0xff, 30, 2 },
	{ "\x20\1", 3, 0xff, 5, 5 },
	{ "\xfc", 0, 0xfe, 3, 13 },
#if 0
	/* These are deprecated and/or returned to the address
	 * pool, so despite the RFC, treating them as special
	 * is probably wrong. */
	{ "", 11, 0xff, 1, 3 },
	{ "\xfe\xc0", 1, 0xc0, 1, 11 },
	{ "\x3f\xfe", 1, 0xff, 1, 12 },
#endif
	/* Last rule must match all addresses to stop loop. */
	{ "", 0, 0, 40, 1 },
};

static const struct policy *policyof(const struct in6_addr *a)
{
	int i;
	for (i=0; ; i++) {
		if (memcmp(a->s6_addr, defpolicy[i].addr, defpolicy[i].len))
			continue;
		if ((a->s6_addr[defpolicy[i].len] & defpolicy[i].mask)
		    != defpolicy[i].addr[defpolicy[i].len])
			continue;
		return defpolicy+i;
	}
}
```

So the `policyof` func simply uses the `defpolicy` array to find the policy that applies to an address. This `defpolicy`
array contains a number of hardcoded entries which correspond exactly to the ones provided in the RFC. The one causing
us problems is `{ "\xfc", 0, 0xfe, 3, 13 }` which matches `fc00::/7` addresses.

Unfortunately, without recompiling musl from source there's not much we can do to address this directly. The only way
we can make the containers work as desired is to give them a different address range. Ideally this would be a range
that you control and that is otherwise not used, but there's a good chance you won't have such a range. One possible
alternative is the `2001:db8::/32` range which is reserved for documentation purposes. This doesn't feature in any of
the policy tables, so is treated like a normal public IPv6 address. It doesn't feel great to misuse a reserved range
like that, but it's probably the least of all evils, at least until musl allows configuring the policy table.