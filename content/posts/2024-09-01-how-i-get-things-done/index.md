---
date: 2024-09-01
title: How I Get Things Done
tags: [productivity, personal]
format: long
permalink: /how-i-get-things-done/
---

I always quite like reading about how other people _do things_. What software
or hardware they use, or how they manage reminders, todo lists, and so on.
I've never actually written about how I do any of that, though. So here it is!

### Productivity

In the past I've fallen victim to the idea of there being One True Productivity
System that would solve all my problems and make me amazing at getting things
done. The title of this post is a nod to Getting Things Done, which I've read
and tried to religiously follow in the past, but it's just not for me. One thing
that did actually stick from it, though, was the idea of "open loops":

> Anything that does not belong where it is, the way it is, is an “open loop,”
> which will be pulling on your attention if it’s not appropriately managed.

If I want to do something and it's not recorded in a way I trust, it
weighs on me a little. Those little weights all add up, and just make life
uncomfortable. I deal with those in a few ways:

- Inbox Zero-ish: if an e-mail needs me to do something, it sits in my inbox
  until it's done. When it's done, it gets archived.
- Notes: random things I try to remember I put into my notes, so I can search
  for them later.
- Budget: anything monetary I just adjust in my budget or create a category for.
- Todo list: any other kind of task I want/need to do either now or at some
  point in the future goes into Todoist.

<!--more-->

#### Todoist

In the past I've gone through a lot of different todo systems, once again
hunting for one that will fix all my problems etc[^1]. In the end I always ended
up coming back to [Todoist](https://todoist.com/), and have stuck with it for a
decent period now.

In Todoist, I have everything split up into projects. Work projects are one
colour and personal projects are another. Anything that needs more than a
handful of individual tasks gets a project, which is archived when it's done.
Any one-off things that don't neatly fall into projects go into "Random junk"
project.

As I'm writing this, I have 116 items in 22 projects. A good deal of those are
things that I might want to do at some point, not things that I actively need
to worry about. If I want to get one of them done, I schedule it so it shows
up in the "today" view. I have the Todoist desktop app open automatically when
I login to my computer, and it gets positioned on the right hand side of my
ultra-wide monitor, so I can see the "today" view and add/remove things
immediately at any point.

I'm not precious about tasks becoming overdue, or unscheduling things I'd
previously scheduled. Priorities change, ideas that sounded good a week ago
might no longer sound good, and so on.

I do a few things to help me manage tasks and keep track of things:

- I have a monthly recurring task to look through everything that's not
  scheduled, just so I keep a rough idea of what's there, and can update or
  get rid of it as needed.
- I wrote [todoistager](https://github.com/csmith/todoistager) and have it
  running on a server to automatically apply tags to tasks based on their rough
  age: `weeks`, `months` or `years`. This lets me identify things that are
  lingering and reconsider them[^2]. 
- I automate adding tasks for things that demand my attention, like incoming
  GitHub pull requests. The Todoist API is nice and straight forward, and
  hooking it up to things is trivial.

### Other desktop software

For notes I use [Logseq](https://logseq.com/), and I just put whatever I need
to write under today's note and rely on search or tags to find it again when
I need to. I note basically anything that I might like to use again in the
future. One of my most common tags is `#blub`, a term I came across in
the excellent article "[In defense of blub studies](https://www.benkuhn.net/blub/)".
Blub is random ultra-specific knowledge; it might not be eminently useful
to know on its own, but over time it can pay dividends.

I've previously tried a bunch of note taking apps, and methodologies, but like
with todo lists I end up just wanting something that _just works_ and stays out
of the way. Logseq stores notes in Markdown, and I have a cronjob to commit and
push my changes to a private git repository.

Like Todoist, Logseq launches when I log into my computer, and sits in the same
500px-or-so region on the right of my monitor. This allows me to switch from
Todoist to take or refer to notes when needed. To do this positioning, I use
[devilspie2](https://www.nongnu.org/devilspie2/). I also use this to position
Discord and my IRC client side-by-side on a small portable monitor that sits
underneath my ultrawide.

More mundane software: I use [IntejjiJ IDEA](https://www.jetbrains.com/idea/)
for almost all development and complex text editing[^3]. The only exception is
Android development where [Android Studio](https://developer.android.com/studio)
rules the roost. For lighter text editing, [Sublime Text](https://www.sublimetext.com/)
is my choice.

I spend a fair bit of the time at the command line. I use [kitty](https://sw.kovidgoyal.net/kitty/)
as a terminal emulator, with the [zsh](https://www.zsh.org/) shell and 
[Oh My Zsh](https://ohmyz.sh/) for some nice plugins and enhancements. One thing
I couldn't live without is [autojump](https://github.com/wting/autojump), which
allows quickly jumping to any directory you've previously visited using a
substring of the name.

Web browsing is all [Firefox](https://www.mozilla.org/en-GB/firefox/new/),
with [Bitwarden](https://bitwarden.com/) for password management and two pinned
[Fastmail](https://www.fastmail.com/) tabs, one for personal e-mail and one for
work. I don't think there's much else to note there.

Oh, I use [Arch](https://archlinux.org/), by the way.

### Infrastructure

This is the hodge-podge of things that support my computer usage but sit outside
the actual computer. In no particular order…

All my DNS queries go through [NextDNS](https://nextdns.io/) using
DNS-over-HTTPS. NextDNS lets you add custom responses, block certain things
from resolving, and subscribe to pre-made anti-tracking or anti-advertising
lists. It gives you far more control over everything than most other services
I've looked at.

Every device I control runs [Tailscale](https://tailscale.com/) to enable remote
access. I also use Tailscale's SSH authentication feature for my servers,
allowing quick and easy access from my phone where I don't have or want normal
SSH keys.

One exciting service that combines these two is [golink](https://github.com/tailscale/golink).
It was created by Tailscale, presumably based on the `go` service used extensively
inside Google. I run a version on my tailnet, and have NextDNS configured to
resolve `go` to the tailscale IP of the service. That lets me type "go/whatever"
on any machine I set up with NextDNS and Tailscale, including mobile devices.
I use go links for all sorts of things:

- Shortcuts to tools I use commonly: `go/board` gets updated to my current
  client's kanban board or similar tracker, whether it's in Jira, GitHub, Asana
  or whatever; `go/meet` goes to Google Meet with the user parameter set
  so I'm logged in with my work account not my personal account.
- In place of bookmarks: `go/flex` goes to the
  [CSS-Tricks flexbox guide](https://css-tricks.com/snippets/css/a-guide-to-flexbox/)[^4],
  `go/cad` goes to the [OnShape](https://www.onshape.com/en/) login page,
  `go/hmrc` goes to the [HMRC online services](https://www.gov.uk/log-in-register-hmrc-online-services)
  page, and so on
- A few dynamic things: `go/github/chameth.com` will go to the `csmith/chameth.com`
  project on GitHub; `go/ref/pipico` will go to the [Pi Pico page](https://ref.c5h.io/view/pipico)
  on my electronics reference wiki.

For my servers I have a minimal [Ansible](https://www.ansible.com/) setup that
does some basic configuration and deploys SSH keys to the right places. Almost
everything runs inside Docker, and I use my own software for HTTP proxying:
[Dotege](https://github.com/csmith/dotege) uses tags I apply to containers to
build a config for [Centauri](https://github.com/csmith/centauri) which handles
reverse proxying, using [Let's Encrypt](https://letsencrypt.org/) for TLS certs.
At the minute I have two servers: a dedicated server that runs basically all my
services (including this website), and a little VPS for monitoring and backups.
Both are from [Hetzner](https://www.hetzner.com/).

I wrote my own basic monitoring software called [Goplum](https://github.com/csmith/goplum)
to make sure things that should be running on my servers actually are. If it
fails checks it sends a notification to my phone using [Pushover](https://pushover.net/),
and a message to a channel on my private IRC server (running [Ergo](https://github.com/ergochat/ergo/)).

### Hardware

For the last two years I've used a laptop as my main computer[^5]. It's a Dell G15
Special Edition. After some upgrades its specs are:

- Intel i7-12700H processor
- NVIDIA GeForce RTX 3060 graphics card
- 2x32GB DDR5 RAM
- 2TB Gen4 SSD

At home it sits behind my [Alienware AW3423DWF](https://www.dell.com/en-uk/shop/alienware-34-curved-qd-oled-gaming-monitor-aw3423dwf/apd/210-bfrq/monitors-monitor-accessories)
monitor. This is a Quantum Dot OLED monitor, and I bought it almost immediately
upon seeing someone else use it at a LAN event. As did three other people that
were sat near us. It just looks **so** good. I wasn't sure I'd get on with
a widescreen monitor, and made sure it supported picture-by-picture before
buying it, but have never actually used that. I generally run whatever I'm
working on at nearly-fullscreen, then have Todoist or Logseq on the far right.
Games go fullscreen over everything, of course.

Above the widescreen I have a [curved lightbar](https://www.amazon.co.uk/gp/product/B09FHMPFW1/)
that I find helps reduce eyestrain and makes it easier to focus on the monitor.
And on top of _that_ I have a [Logitech Brio webcam](https://www.logitech.com/en-gb/products/webcams/brio-4k-hdr-webcam.960-001106.html)
which gives a great picture with lots of control over pan/zoom/etc for meetings.
I have a little set of systemd scripts that detect when the external webcam is
connected and disable the laptop's built-in webcam, and vice-versa when it's
disconnected; that means whenever I jump into a video call in any app there's
only one input and it's always the right one, which is nice.

Below the main monitor and its friends, I have a generic 1080p portable monitor
connected to the laptop via USB-C. This comes with me when I travel, and serves
as a Discord/IRC monitor when at home. It's handy for throwing other things on
as well, like if I need to reference a design while writing code, and don't
want to keep alt+tabbing.

I have two [Durgod K320](https://www.durgod.com/product/k320-space-gray/)
keyboards: one with Cherry MX Silent Red switches for when I'm at home and don't
want to annoy people with the sounds, and one with Cherry MX Blues for when I'm
travelling and can make a bit more noise in exchange for a much nicer typing
experience. They're both tenkeyless so they can handily fit in a backpack.

I use a [Zowie EC1-C](https://zowie.benq.eu/en-uk/mouse/ec1-c.html) wired mouse,
which does the job well. I got fed up with mice from more mainstream manufacturers
failing: my previous mouse was a Razer Viper that lasted 7 months before it
started losing clicks. All my audio goes through a
[HyperX Cloud Alpha Wireless](https://uk.hyperx.com/products/hyperx-cloud-alpha-wireless),
which has the most insanely good battery life of any wireless headset I've ever
owned or seen. Their claim of 300 hours is not just marketing!

In non-computer hardware, I have a [Bambu P1S](https://uk.store.bambulab.com/products/p1s) 3D printer
that I've mentioned in my blog before. For printing in fewer dimensions I have an
[Epson EcoTank ET-1810](https://www.epson.co.uk/en_GB/products/printers/inkjet/consumer/p/30174),
which is surprisingly reasonable for an inkjet printer, and the ink costs aren't as astronomical
as normal cartridges. Plus, you can't decide to DRM ink when it's loaded raw into the
printer…

I have a small Zigbee network set up mainly to control the lights and the old
3D printer I no longer use (the P1S has a sensible standby mode so you don't
have to yank its power away to make the fans stop spinning!). That's run through
a Raspberry Pi running [Zigbee2Mqtt](https://www.zigbee2mqtt.io/) and a custom
app I wrote to monitor events on MQTT and generate the appropriate responses
(or just log the data).

My daily-driver phone is an [iPhone 15 Pro](https://www.apple.com/uk/iphone-15-pro/),
but I have a bunch of Android devices kicking around for work purposes. I read
nearly every day on a [Kindle Paperwhite Signature Edition](https://www.amazon.co.uk/gp/product/B08N2QK2TG/).

### Money and work

I use [You Need A Budget](https://www.ynab.com/) for figuring out how much money
I can spend on what, and making sure I have enough put aside for bills and taxes
and so on. I use it for both business and personal accounts together, but I mark
all the business categories and accounts with a 💼 emoji, and have some scripts
that use the API to make sure the money split makes sense[^6].

Every recurring monthly and annual payment has a category, as do all upcoming
events I plan to attend, and all the wonderful taxes I have to pay. I also have
categories for hobbies and other expenditure. I think all of that is pretty
typical for a normal YNAB setup.

I also use YNAB for tracking anything I might like to buy at some point: it
gets a category in a "Wishlist" group, and I'll occasionally budget to buy
things from there. I try (and often fail) to use that as a way to put off
impulse purchasing. As my income is often very inconsistent, I try to set
aside money so that I can "spend" a bit every week on things in the wishlist
group. That feels a lot nicer to me than just budgeting everything in one go
as soon as I get paid, and gives me more time to consider whether I _really_
want those things[^7].

For work, I do all the more technical accountancy and tax filing and so on
through [FreeAgent](https://www.freeagent.com/). For my fairly straight-forward tax
situation, I can basically manage everything myself using FreeAgent without
paying for an accountant. That does mean occasionally spending a bunch of time
reading tax manuals, but I'd rather fully understand the rules than just follow
advice anyway.

[^1]: The problem with implementing productivity systems or switching todo apps
or anything of that ilk is that it can _feel_ like you're being productive by
doing so. It takes a while to realise that you're not, and actually getting on
with things is better.

[^2]: Or just accept that I'll maybe do them at some point. I kept the labels
vague to try and not make it feel judgmental.

[^3]: Including writing this blog post right now.

[^4]: One day I will get all the `justify-` and `align-` options right without
looking them up, but until then…

[^5]: This is very normal for developers, but much less common for computer
gamers. I'll probably go back to a desktop next time I upgrade, as I'm
travelling a lot less.

[^6]: I can't just arbitrarily move money from the business to myself, it has
to be given as salary or a divided etc, which has tax implications. If I budget
"business money" for personal things, then it actually has to be properly
transferred and accounted for before being spent (as I'm in effect budgeting
a future paycheck). It would be cleaner to keep the accounts completely
separate, but I find this way works well for me as I can be a bit more flexible.

[^7]: It feels a little like I'm giving myself pocket money, which is a bit
weird, but it's a lot better than having no "income" for several months when
client invoices fall weirdly or I'm between contracts.