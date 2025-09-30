---
date: 2025-09-30
title: "Avoiding the Consequences of Dumb Laws with Tailscale"
permalink: /avoiding-the-consequences-of-dumb-laws-with-tailscale/
tags: [networking]
format: short
resources:
  - src: apps.png
    name: "Screenshot of the app section in the Tailscale admin console. It shows a table with two entries: 'reddit' and 'bluesky'. Each entry has a list of domain names like '*.reddit.com, *.reddit.it'."
    title: "App configuration in the Tailscale admin console"
---

More and more sites are implementing privacy-invading age checks or just
completely blocking the UK thanks to the [Online Safety Act](https://www.legislation.gov.uk/ukpga/2023/50/contents).

Protecting kids from _some_ content online is certainly a noble goal, but
the asinine guidance from Ofcom, threats of absolutely disproportionate fines,
and the stupidly broad categories of content have resulted in companies just
giving up or going through a tick-box exercise that offers very little
protection but lots of inconvenience and a complete invasion of privacy.

Instead of uploading my ID to some third party company, I've taken to proxying
my traffic through to a country that doesn't have such stupid laws. Thankfully,
Tailscale makes this really easy. I've discussed [how I use Tailscale](https://chameth.com/how-i-use-tailscale/)
before, but not really covered _app connectors_. I find Tailscale's description
of these pretty confusing, but they basically amount to automatic, DNS-based
subnet routing configurations (or, to put it another way, a per-website exit
node). You can safely ignore all references to 'SaaS apps' in their docs.

I create a custom app connector, and give it the domains to be included:

{%img "Screenshot of the app section in the Tailscale admin console. It shows a table with two entries: 'reddit' and 'bluesky'. Each entry has a list of domain names like '*.reddit.com, *.reddit.it'." %}

Tailscale then magically resolves those domains, and has the 'connector'
advertise routes for them. Any client that accepts routes will start sending
requests to the connector, which passes them onto the Internet at large. Any
other traffic is left alone, unlike when you use an exit node.

The special bit here is how you can specify wildcard domains. Tailscale proxies
the DNS requests from clients (so it can inject responses for nodes on your
tailnet), which means it can dynamically update the routes as you resolve new
domains. I tried to set this up more manually, and quickly came unstuck: despite
using the same DNS servers, my server and my desktop would get different responses
for the same query as it varied by geography. Trying to get the full set of
IPs (and keeping them updated) would have been a nightmare. Tailscale expanding
the wildcards nicely sidesteps all of that.

At first I was just proxying the traffic to one of my servers, but just today
I added a new connector for Imgur and found I was still blocked, just for
different reasons. They not only block my entire country but also a load
of known datacenter IP ranges. Hmph. I fixed this by hacking up a new side
project: [tsv](https://github.com/csmith/tsv). It's a simple Go app that accepts
traffic from the tailnet (advertising itself as both an app connector and an
exit node), and passes it on to another VPN.

There are lots of other ways you could accomplish this, but this makes it so
all my devices can still access services without any additional configuration.
As long as Tailscale is installed, the Internet will still work as it's meant
to, without all the nonsense. If I come across a site that doesn't work, adding
it is trivial: I just make a new app connector in Tailscale.

Obvious disclaimer: the laws in the UK are binding on the service providers,
not the end user. Doing this sort of thing in other countries might be illegal.
I don't know; do your own research! Also all of this is a workaround
for something that should be fixed at a legislative level, but I'm not holding
my breath.