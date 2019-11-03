---
date: 2017-08-16
title: A look at the DNS habits of the top 100k websites
description: An analysis of the DNS providers and resilience of the top 100,000 websites.
area: data analysis
url: /2017/08/16/top-sites-dns-providers/

resources:
  - src: providers.png
    name: Graph showing popularity of DNS providers across sites grouped by position
    params:
      default: true
  - src: resilience.png
    name: Graph showing use of resilience techniques by site position
  - src: provider-pairings.png
    name: Chart showing frequency of pairings of top providers
---

I was thinking about switching DNS providers recently, and found myself
`whois`ing random domains and looking at their nameservers. One thing lead
to another and I ended up doing a survey of the nameservers of the top
100,000 sites according to Alexa.

### Most popular providers

The top providers by a large margin were, unsurprisingly, Cloudflare and AWS
Route 53. Between them they accounted for around 30% of the top 100k sites.

<!--more--> 

The top 10 providers overall were:

| Provider        | Country       | Sites |
| --------------- |---------------|------:|
| Cloudflare      | United States | 19%   |
| AWS Route53     | United States | 10%   |
| GoDaddy         | United States | 4%    |
| DNSPod          | China         | 3%    |
| Dyn             | United States | 2%    |
| Akamai          | United States | 2%    |
| DNS Made Easy   | United States | 2%    |
| Hi China        | China         | 1%    |
| UltraDNS        | United States | 1%    |
| Namecheap       | United States | 1%    |

You have to search fairly deep to find a provider that's not American or
Chinese: OVH (France), Gandi (France again) and RU Center (Russia) all come
in at around 0.5% of the top sites.

One thing I found particularly interesting was the relatively small number of
sites that use Google's hosted DNS service -- out of the 100,000 sites only
0.4% appear to use Google Cloud DNS. That's 25 times fewer than are using
Route 53.

#### Different strokes for different folks

This graph shows the relative frequency of some of the big providers for
sites in different positions in the top 100,000 list:

{{< img "Graph showing popularity of DNS providers across sites grouped by position" >}}

There are a few interesting transitions that can be seen here. The very large
sites tend to manage their own DNS, as can be seen with the large
'Self-hosted/other' number in the top 100 category. As you move down into the
top thousand, you get to sites that still have significant requirements but
don't quite have the need to run their own DNS infrastructure; here you can see
Akamai peak, and Cloudflare usage jump up an order of magnitude.

As you travel further down the list, DNS becomes a much more mundane affair
and you see 'premium' providers such as NS1, Dyn and Verisign drop off, and
commodity providers such as GoDaddy start to soar. Cloudflare remains a popular
option for these sites thanks, I imagine, to its generous free plan.

### Resilience

In October 2016, Dyn was subject to
[a large DDoS attack](https://en.wikipedia.org/wiki/2016_Dyn_cyberattack) that
cripped a significant number of major websites. There are two main ways that 
individual sites can mitigate such an attack: they can host DNS themselves (in
which case it's as vulnerable to a DDoS attack as the rest of their
infrastructure), or they can use multiple DNS providers effectively hedging
their bets.

There's one other potential issue that may affect DNS resilience: the
reliability of the TLD's nameservers. Shortly after the Dyn outage, the
majority of the nameservers for the `.io`, `.ac` and `.sh ` TLDs went down.
If your nameservers were under one of those TLDs, clients would again be unable
to reach them. The easiest way to reduce the risk of this happening is to have
namesevers under multiple TLDs.

As you would expect, the use of these techniques tend to be more common with
the higher ranking sites:

{{< img "Graph showing use of resilience techniques by site position" >}}

#### Most popular pairings

Of those sites that do use multiple providers, there are some fairly common
pairings:

{{< img "Chart showing frequency of pairings of top providers" >}}

Dyn is obviously frequently paired with a number of providers. In fact, of all
the top 100k sites using Dyn 40% also use a different provider. They're second
only to NS1, who despite having smaller absolute numbers, appear alongside one
of their competitors on 72% of the sites that use them.

NS1 also suffered from [DDoS attacks](https://nsone.statuspage.io/incidents/g9fkrhqr7wnv)
over the summer of 2016. It seems that after a major outage, customers wisely
tend to hedge their bets and introduce a backup provider.
