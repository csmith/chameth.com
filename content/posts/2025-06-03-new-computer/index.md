---
date: 2025-06-03
title: Building a new Computer
permalink: /building-a-new-computer/
tags: [hardware, computer, personal]
resources:
  - src: finished.jpg
    name: A small form factor PC sat on a desk, with a pen propped up in front of it for scale. It's about 1.5 pens tall. The case is silver, with a wooden panel at the bottom.
    title: The finished PC. Cat pen for scale.
  - src: build.jpg
    name: The computer fully built, but with the sides and top of the case off, and wires protruding everywhere.
    title: The case with sides and top off, and wires everywhere in desperate need of management.
  - src: case.jpg
    name: An open flight case, with the PC slotted in surrounded by foam. A keyboard sits below it, and various accessories on the side with 3D-printed supports.
    title: Travel mode, engage!
opengraph:
  title: "Building a new Computer · Chameth.com"
  type: article
  url: /building-a-new-computer/
  image: /building-a-new-computer/finished.jpg
---

{% figure "right" "A small form factor PC sat on a desk, with a pen propped up in front of it for scale. It's about 1.5 pens tall. The case is silver, with a wooden panel at the bottom." %}

I recently[^1] built a new computer, after exclusively using a laptop for three
years. It's also the first time I've departed from the usual combo of an Intel
CPU and Nvidia GPU.

While the form factor of a laptop did make it amazingly handy for travelling
and attending LAN events, it was starting to show its age and there was
basically no sane upgrade path. The main problem was its 3060 mobile graphics
card, which was _okay_ for the first few years and then slowly descended into
_painful_. At the time, the (fairly disappointing) 50-series had just been
released, but hadn't yet made it into laptops. Nvidia had stopped production
on the 40-series beforehand, so there also weren't any compelling options there,
either.

That's not to say there were no laptops at all that I could have upgraded to.
There were. Just not really any that ticked all of my perfectly normal boxes
like "run Linux", "have at least 64GB of ram", and "run modern games well". 
Having to upgrade the entire system just because the graphics card was
showing its age was a bit of a drag, too. So back to the land of desktop
computers I went!

To try to preserve some of that convenience, I opted for a small form factor
(SFF) case. The computer and all the accessories can fit inside a
hand-luggage-sized flight case. More on that later, though.

### Components, Choices, and Cramming

So what, exactly, is in this computer? The case is a Fractal Design Terra,
which is a lovely little 10.4L case[^2]. My last desktop PC was a normal-sized
Fractal Design case, which I liked a lot, and it seems like they've upped their
game since then. The Terra is both pleasant to look at (just look at that
wooden panel!), and a breeze to work in. The sides and top are fully and easily
removable, giving you great access to everything inside[^3]. I expected an SFF
build to be fiddly, but the case made it feel about the same as a normal
full-sized build[^4].

<!--more-->

{% figure "left" "The computer fully built, but with the sides and top of the case off, and wires protruding everywhere." %}

Another nice feature of the case is the movable spine. I think this is
relatively common in SFF cases, but it was new to me. You mount the graphics
card on one side of a metal divider, and everything else on the other. The
divider (or spine) can be moved side-to-side depending on what exactly you're
putting in the case. 

As I mentioned earlier, this was my first departure from the Intel + Nvidia
bandwagon. Nvidia's 50 series was underwhelming, and Intel managed to put out
a generation of chips that literally fried themselves. It seemed like a good
time to try out AMD's offering! I went for the newly released Radeon RX 9070XT
graphics card, and a Ryzen 7 9800X3D processor[^5]. The processor has a
Thermalright AXP90 X47 Full heatsink on top of it, with a Noctua NF-A9 fan
on top of that.

The motherboard is a Gigabyte B850I Aorus Pro, and it hosts 2 32GB Corsair
Vengeance RAM modules, and a 2TB Crucial T705 NVMe drive. Powering it all is
a Cooler Master V750 SFX. The PSU is modular to reduce the number of wires that
need to be fitted in, but they're still pretty big and bulky. I'm tempted by
custom cables but they're rather pricey, and you can't actually see inside
with the sides closed…

With the diminutive size of the case and all the components crammed in,
there's actually no room left over for case fans. The PSU, GPU and CPU all have
fans attached, but that's it. There are stories online of people having heat
issues, but thankfully everything seems pretty happy in my setup. The components
all draw in air from the sides, and the hot air passively vents out the top. If
I keep the CPU under maximum load for a while then it starts to thermal throttle
slightly, but the only time I've actually managed that is when compiling a
kernel. Why was I compiling a kernel, you ask? Well, we'll come to that.

### But does it Linux?

I didn't go into this build _completely_ blind. I did a bit of reading and
chatting, and it sounded like AMD drivers were pretty good on Linux these
days. I didn't check specifically if any of the components were supported.

The 9070XT was released on the 6th of March. I built my PC on the 26th
of March. The first version of Mesa (the 3D graphics lib) that was generally
stable with 9070XT cards landed in Arch on the 20th of March. If I did the build
just one week earlier, I would've had a much less fun experience.

It still wasn't quite perfectly smooth, though. When playing games I'd
occasionally have issues. In some cases the screen just went blank and came
back after switching windows, but in others the graphics drivers couldn't
recover and the PC needed rebooting. Fortunately the drivers are open source,
and they have a public [issue tracker](https://gitlab.freedesktop.org/drm/amd)
hosted by the Freedesktop project. Searching through the issues, I found they'd
already been reported (by another Arch user, of course), and there were various
suggestions for work-arounds, as well as lots of work trying to reproduce the
issue.

Within a week, an AMD developer had posted a set of patches that might address
the issues, and asked people to test them. That's fortunately pretty easy on
Arch: you just clone the repository containing the `PKGBUILD` file, drop in
the patches, and update the checksums for them. Then you start compiling,
realise you didn't tell it to use more than one CPU core, abort it, and start
again but this time running on all cylinders. Soon enough I was running a
patched 6.14 kernel and the problems stopped completely. Others reported
similar outcomes, and the patches worked their way into the kernel proper.

I spent a few weeks having to compile a custom kernel every time I accidentally
updated it (I should've just stopped `pacman` from touching the kernel for a
bit…), then one day the patches didn't apply because they'd been merged
upstream. Since then everything has been perfect.

A friend commented that "this wouldn't happen with Nvidia". They meant it
negatively, but I wholeheartedly agree! Maybe Nvidia have fewer launch bugs
than AMD, but you _absolutely_ wouldn't be able to talk to their driver
engineers on a public bug tracker, and get access to test patches that fix
your issue. I'll happily take a few bugs in exchange for a company that operates
openly in a way consistent with the rest of the Linux community.

### But does it travel?

{% figure "right" "An open flight case, with the PC slotted in surrounded by foam. A keyboard sits below it, and various accessories on the side with 3D-printed supports." %}

So the computer works fine at home, but how do I get it to other places? I just
shoved a laptop in a backpack[^6], but you can't quite do that with a PC even
one this size. I'd seen a few people at LANs who packed everything in flight
cases, but they always seemed a bit unwieldy. Doing some research, though, I
found that I'd just be able to squeeze everything into a flight case sized to
be taken onto a plane as cabin baggage.

Peli (or Pelican in the US) are the most well known maker of flight cases, but
I chose to go with Nanuk instead. Spec-wise they're fairly similar. Nanuk are
a bit cheaper. The killer difference for me, though, is that Nanuk publish
STEP files of their cases. As someone with a 3D printer, giving me CAD models
of things so that I can easily mod them is definitely a way to win my favour.

After lots of measuring and thinking and modelling I came up with an arrangement
where everything could fit snugly into the case. The case has cutouts for the
wheels and handles, which constrain what can go where. I ended up wanting the
computer at the handle-end of the case to make best use of space, but that means
it needs something to take its weight when the case is vertical. The solution
was, of course, 3D printing.

Using the STEP files from Nanuk I designed an insert that spans across the
case, with screw holes that line up with the ones in the flight case. I wasn't
sure it would be able to take the weight like that, so I also added in some
supports in the other direction (running down to the bottom of the case, if it's
stood up). I ended up printing it in three parts, and it came out perfectly first
time. I can't emphasise how nice it is having CAD models instead of having to
measure and guess at things.

With the support in place, I focused on cushioning the computer. I opted for
10mm thick EVA foam as it's easy to obtain and work with, and doesn't take up too much
room. I used contact adhesive to glue the foam directly to the case and the
3D printed divider. There's a piece that goes between the computer and the
keyboard, so I printed some little 3D supports to hold that more firmly in
place, and stuck those in with contact adhesive as well.

So how does it all hold up? It's been to one LAN event so far, and it's great.
I think the flight case might actually be easier to carry than hefting a heavy
backpack: the handles are sturdy and comfortable, and it has chunky wheels for
pulling it around. It's also actually easier to get everything set up when you
can see it all laid out in the flight case and pull out what you need, and then
the case can be tucked away out the way for the duration of the event.

[^1]: Recently-ish. It may or may not have taken me a few months to get around
to writing this post.

[^2]: Using litres to measure computer cases always feels wrong. Those are
the units we use for water. Don't put them near the computers!

[^3]: The case also has holes to put security screws in to stop all of these
nice easily removable parts from being removable. Very handy when you're
leaving it unattended at a LAN.

[^4]: Which for me means it was good fun for about an hour, and then I got fed
up and just wanted it to be done already.

[^5]: Three months in and I still can't remember these random strings of
numbers. Someone even explained the versioning scheme to me. 

[^6]: Along with a portable monitor, mouse, keyboard, headset, mouse mat,
kensington lock, power supply, cables, and controller. It was a heavy bag.