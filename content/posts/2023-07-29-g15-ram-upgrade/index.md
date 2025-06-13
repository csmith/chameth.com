---
date: 2023-07-29
title: Upgrading the RAM in a Dell G15 laptop
permalink: /g15-ram-upgrade/
tags: [hardware, computer]
resources:
- src: g15.png
  name: An open Dell G15 laptop
  title: The Dell G15
  params:
    default: true
- src: motherboard.jpg
  name: The G15 motherboard, with a large "DDR5 8G/16G Only" label, and a smaller "DIMM B DDR5 8G/16G" label next to a SODIMM slot
  title: The G15 motherboard adamantly proclaiming that it doesn't want 32GB SODIMMs
opengraph:
  title: "Upgrading the RAM in a Dell G15 laptop · Chameth.com"
  type: article
  url: /g15-ram-upgrade/
  image: /g15-ram-upgrade/motherboard.jpg
---

{% figure "left" "An open Dell G15 laptop" %}

I currently use a Dell G15 laptop for work. It has served me well for a little
over a year, but recently it has been struggling a little with my day-to-day
workload. It came with 32GB of RAM — the highest possible specification at the
time[^1] — but that is apparently no longer enough for me.

For a recent project, I was working on a Rust library used in an Android app.
That meant running the usual glut of Android tools (Android Studio, an emulator
and at least one Gradle daemon) alongside a normal IDE (IntelliJ IDEA). Throw
in a web browser and a couple of electron apps, and I often managed to
use all 32GB.

When you start swapping memory out to an encrypted disk — even an SSD — it
doesn't make for great performance. At first, I tried to work around this
by enabling the Linux out-of-memory (OOM) killer, but it turns out that it's not
too good with Electron apps: it will kill the large browser process, but then
the small Electron wrapper will just respawn it.

<!--more-->

### Can it be upgraded or not?

The obvious solution to not having enough RAM is to add more RAM. A quick look
in the manual showed this might not be possible, though. The manual includes
the following "Memory specifications" table:

| Description                     | Values                                                                                                                                                                                                 |
|---------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Memory slots                    | Two SODIMM slots                                                                                                                                                                                       |
| Memory type                     | DDR5                                                                                                                                                                                                   |
| Memory speed                    | 4800                                                                                                                                                                                                   |
| Maximum memory configuration    | 32GB                                                                                                                                                                                                   |
| Minimum memory configuration    | 8GB                                                                                                                                                                                                    |
| Memory size per slot            | 8GB or 16GB                                                                                                                                                                                            |
| Memory configurations supported | <ul><li>8 GB, 1 x 8 GB, DDR5, 4800 MHz</li><li>16 GB, 1 x 16 GB, DDR5, 4800 MHz</li><li>16 GB, 2 x 8 GB, DDR5, 4800 MHz, dual-channel</li><li>32 GB, 2 x 16 GB, DDR5, 4800 MHz, dual-channel</li></ul> |

That unambiguously says that an upgrade from 32GB is not possible. I gave up.

Later though, I was complaining about memory issues to a friend, and he pointed
out a Dell forum thread where a couple of people claim to have successfully
installed dual-channel 32GB modules. Since the alternative was getting an
entire new PC after only a year, I decided to give it a go.

### The upgrade attempt

I ordered a pair of Crucial 32GB DDR5-4800 SODIMMs, and after they turned up
dismantled the laptop. The G15 comes apart pretty normally: there are uncovered
screws on the bottom holding the lower part of the case on. With those
removed and some gentle prying, it pops off, and you get access to the battery,
GPU and motherboard.

The first thing I saw was this:

{% img "The G15 motherboard, with a large \"DDR5 8G/16G Only\" label, and a smaller \"DIMM B DDR5 8G/16G\" label next to a SODIMM slot" %}

Not one but two labels that indicate it will only accept 8GB or 16GB modules.
Oh well, what's the worst that can happen?

### Oops?

I dutifully installed the new modules, reconnected the battery and put the
case back together. I pressed the power button, and… nothing. None of the
usual garish lights that immediately come on, no screen output, just a dead
laptop. After reading some more of the user manual, I found that there is a tiny
status LED on the side next to the ethernet port. Forcing the laptop to power
off and back on again, the status LED blinked a distress code at me: 2 amber
blinks, 4 white blinks. The manual says that is — unsurprisingly — a memory
fault.

I figured at this point that the manual and labels on the motherboard were
probably right. I took the laptop apart again, reinstalled the original 2x16GB
modules, reassembled it, and pressed the power button. It didn't boot. I don't
spend a lot of time fiddling inside computers, but I've done it enough that
I'm reasonably confident I can't entirely break a computer while swapping some
RAM modules. I took to Googling[^2], and found an interesting article that said
Dell laptops don't like to boot after RAM changes unless you clear the CMOS by
popping out the battery for 15 minutes.

I opened the laptop up, and looked around for the CMOS battery. There wasn't
one. Turns out they don't exist any more. I left the main battery disconnected
for a while to see if it would help, and it didn't.

### Unexpected success

I started to get worried: if I couldn't fix this, I wouldn't be able to
work until I got a new PC, and that wasn't really in my budget at the minute.
I sat reading old forum threads and help guides[^3], none of which were actually
useful. Out of nowhere, though, the laptop booted up.

Naturally, I immediately shut the laptop down again, opened it up, and switched
back to the new RAM modules. Then I turned it on again and sat waiting. After
about 15 minutes of it looking totally dead, it turned on and showed a BIOS
warning about the hardware configuration being changed. It then booted perfectly
normally, and all 64GB of RAM was visible and usable.

My theory is that the forum threads were right: Dell laptops are funny about
RAM upgrades. But somehow in removing the physical CMOS battery, they've kept
the same "you have to wait 15 minutes" behaviour just without any indication
that's what's happening. Regardless, I now have enough RAM even for the
greediest of IDEs and Electron apps.

[^1]: Bizarrely, the maximum spec has _decreased_ to 16GB since then.

[^2]: In the generic sense. I use [Kagi](https://kagi.com/) these days.

[^3]: On my phone because, y'know, the laptop was busted.