---
date: 2016-06-17
title: Why you should be using HTTPS
description: There's no good reason for sites to avoid HTTPS any more, and lots of reasons they should be actively encouraging it.
area: security
permalink: /why-you-should-be-using-https/

resources:
  - src: https-everywhere.jpg
    name: The EFF's HTTPS Everywhere logo
    params:
      default: true
---

{% figure "left" "The EFF's HTTPS Everywhere logo" %}

One of my favourite hobbyhorses recently has been the use of HTTPS, or lack thereof. HTTPS is the
thing that makes the little padlock appear in your browser, and has existed for over 20 years.
In the past, that little padlock was the exclusive preserve of banks and other 'high security'
establishments; over time its use has gradually expanded to most (but not all) websites
that handle user information, and the time is now right for it to become ubiquitous.

### Why use HTTPS?

There are numerous advantages to using HTTPS, both for the users of a website and for the
operator:

#### Privacy

The most obvious advantage is that HTTPS gives your users additional privacy. An insecure (HTTP)
request can potentially be read by anyone on the same network, or the network operators, or anyone
who happens to operate a network along the path between the user and the server.

Users on shared WiFi networks (such as those in coffee shops, hotels, or offices) are particularly
vulnerable to passive sniffing by anyone else on that network. If the network is open (as is
frequently the case) then anyone in radio range can see exactly what the user is up to.

<!--more-->

#### Integrity

HTTPS also helps to maintain the integrity of your site. With a plain HTTP request, there's nothing
to stop anyone in between the server and the user from modifying the content of the request or the
response. This is a frequent tactic used by annoying WiFi gateways (such as the ones [you'd find in
a hotel](http://justinsomnia.org/2012/04/hotel-wifi-javascript-injection/)), dubious ISPs who want
to serve you extra adverts, or just plain old nefarious attackers.

If you're trying to convey some kind of information to users (and if you aren't, why exactly are
you running a website again?) it seems beneficial to both you and them if the information arrives
as you intended, rather than in a modified form due to someone or something tampering with it.

#### Security

If your website has any kind of authentication, or session identifiers, it becomes extremely
vulnerable to an attacker monitoring the traffic and stealing the credentials. This was
starkly demonstrated in 2010 when [Firesheep](https://en.wikipedia.org/wiki/Firesheep) was
released. This tool allowed anyone to quickly and automatically hijack social media accounts of
anyone on the same network who was using HTTP to access them.

Even if your login pages are served over HTTPS, if you send a single session ID cookie over HTTP
(such as a page you decided wasn't particularly 'important') then an attacker can probably spoof
the user's session and gain full access to their account.  Again, in the case of open WiFi networks
that could be anyone in radio range.

#### Search engine rankings

Some search engines use HTTPS as a signal in their ranking algorithms. [Google announced in
2004](https://security.googleblog.com/2014/08/https-as-ranking-signal_6.html) that it was using
the presence of HTTPS as a small positive signal, but that it may strengthen that signal over time
as more and more websites switch to using a secure transport. It's not unthinkable that at some
point in the future there will be HTTPS-only search engines.

### But... But... But...

There are lots of excuses for not implementing HTTPS. Most of them are either misguided or outdated.

#### It's too expensive and/or complicated

In the past, getting HTTPS certificates was a pain. A number of free suppliers have existed for
a while but the process for getting their certificates wasn't particularly straight forward, and
many imposed arbitrary restrictions on the certificate parameters. Even once you had the
certificate, you had to fiddle about with your HTTP server configuration to make it work, remember
to manually get a new certificate when the old one expired, and lots of other annoying busywork.

With the arrival of [Let's Encrypt](https://letsencrypt.org/), all that changed. You can retrieve
and deploy a free HTTPS certificate with two or three commands. Renewal can be handled completely
automatically with a single command executed by cron.

#### There's no point; nothing on my site is sensitive

You might not think your content warrants privacy, but can you speak for everyone who accesses it?
Even content that seems mundane to you — such as travel advice, or technical writing — could be
used to build up a profile of a user. If an attacker is monitoring traffic in a coffee shop and
sees a user looking at travel advice and weather forecasts for a foreign country, he could use that
information to plan a burglary knowing that the user will be away. Similarly, some content which
is perfectly mundane to you may actually be very sensitive in other countries with repressive
governments. HTTPS makes it much harder for these people to snoop on traffic.

From another angle, if you're offering any kind of information, instructions, or especially file
downloads, there's a severe risk to users if the content is modified on its way to them. An evil
sysadmin could rewrite your travel advice to suggest visiting the local drug dealer's hangout, or
replace your download with a malware-infested version.

#### HTTPS is slower, uses more resources, etc

Back in 1995 this might have been a valid argument. Enabling HTTPS on a modern server will make
an almost negligible difference to performance. If you also enable HTTP/2 (which most
implementations only support over HTTPS), it's likely to actually use fewer resources, and result
in a faster, smoother experience for your users. HTTP/2 was designed to work with HTTPS, and
designed with modern requirements and networking techniques in mind.

CloudFlare have an [excellent demonstration](https://www.cloudflare.com/http2/) of the benefits of
HTTP/2, and it can show speed improvements of 2-3x in a typical environment. On top of being faster,
HTTP/2 uses fewer connections which results in less resource overhead on both the server and the
client.

### So what are you waiting for?

If you run a website and aren't using HTTPS, [give it a try](https://certbot.eff.org/).
