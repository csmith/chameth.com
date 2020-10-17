---
date: 2020-10-17
title: Apple, Google and aligned incentives
description: Why I switched from Android to iOS after over a decade
area: mobile
slug: apple-google-aligned-incentives

resources:
  - src: htc-dream.jpg
    name: White HTC Dream mobile phone
    title: The HTC Dream, the first phone released running Android.
    params:
      default: true
  - src: pixels.jpg
    name: Pixel and Pixel XL phones
    title: The Pixel and Pixel XL
---

{{< figure "right" "White HTC Dream mobile phone" >}}

For the past decade I've exclusively used Android phones. I got the HTC Dream (aka the T-Mobile G1)
shortly after it came out, and dutifully upgraded every 1-2 years. In that timespan I used Android
as the basis for my Master's Thesis, took a job on the Android team at Google, and eventually
became a contractor specialising in Android app development. So when I switched to using an iPhone
earlier this year a few people were surprised[^1].

### The good old days

When Android was announced in 2007 -- alongside the formation of the Open Handset Alliance -- it
was positioned as a bastion of openness: it would be built on open standards and the operating
system would be open source. At the time iPhones were strongly coupled to iTunes and Apple was
exercising strict control over what app developers could do.

<!--more-->

When the HTC Dream was released it lived up to expectations. You could write apps for it without
shelling out for a Mac! You could get root access, and it was running Linux under the hood! A
whole ecosystem of custom firmwares and bootloaders started to appear, thanks to the open source
nature of the OS. It shipped with some Google apps, but they were just normal apps that served
as examples of what could be done.

After the Dream came a line of Nexus devices. These were Android's flagship devices, designed to
show off what a good Android phone should look like. Both the hardware and software releases
tended to feature interesting, useful upgrades. Some of the original open source apps were
replaced with closed source, Google proprietary ones, but that was OK - the open source 
versions lived on in the open source project as examples of what you _could_ do. The devices
allowed flashing custom firmware, and the OS source was always released... eventually.

### The downfall

{{< figure "left" "Pixel and Pixel XL phones" >}}

Over time, Android has got less and less free. The Nexus line of phones gave way to the Pixel
range, which got rid of the clearly demarcated border where Android ended and Google began.
[The Android 11 highlights](https://www.android.com/android-11/) list is dotted with sections
that say "On Pixel devices...", and every single one is a software feature that could be
implemented on any device, but Google have decided to keep it proprietary instead of releasing
it as part of the open source platform.

At the same time more and more functionality has been added to Google Play Services. This was
originally a shared location for Google specific services - in 2012 it merely handled some
Google+ functionality and dealing with OAuth for Google accounts. These days it contains a
huge swathe of Google services, as well as platform functionality such as push notifications,
barcode scanning, geolocation, and so on. Google Play Services isn't part of the Android
Open Source Project, and is only available under license from Google. You can't really have
a device without these functions, so as a manufacturer you have the choice between agreeing
to whatever terms Google requires[^2] or spending an awful lot of development time creating an
alternative.

While these are fairly abstract arguments about what a free platform should look like,
during the same period there has been a marked restriction in how you can actually use
Android devices -- both as an end-user and as a developer. iOS has always had very stark
restrictions on what apps can do in the background; Android started its life allowing
pretty much anything, like any good general purpose computing device. This inevitably
lead to lots of apps doing lots of things that ranged to stupid to mildly suboptimal,
creating a [tragedy of the commons](https://en.wikipedia.org/wiki/Tragedy_of_the_commons)
amongst apps. The victim was the phone's battery life, and Google's solution was a
progressive series of restrictions on what apps can do and when. This includes limits
on when push notifications are delivered to a device, how frequently apps can wake up,
and so forth. Some of these can be bypassed by the end-user, but not all, and the
process is fairly cumbersome.

### Aligned incentives

From my point of view, Android and iOS are now pretty much in a similar place.
They're not general purpose computers, but 
[app consoles](https://daringfireball.net/linked/2020/08/14/orland-epic-game-consoles):
much like games consoles they consist of hardware and software that the user lacks
control over but accepts in order to access the library of apps/games.

Given there's no clear winner between them in terms of hardware and software, my
decision came down to a more holistic question: how aligned are their
incentives to my own? Apple is a product company: they make their money by
producing shiny things that people want to purchase; Google is an advertising
company: they make their money by using my personal information to show me
targeted adverts.

For Google, Android was originally a strategic move to ensure that
Apple couldn't dominate the mobile web, and by extension the revenue from ads.
As a company, there's nothing to really push them forward in any particular
direction other than one that facilitates advertising. Of course, Google employs
tonnes of good people who want to do good things which counterbalances this, but 
I'd still prefer to deal with a company that has an intrinsic motivation to
do things I want them to do, rather than one forced to by regulation, custom,
or the good intent of their employees.

So the decision became obvious: if I have no strong opinions about the software
and hardware, Apple is the clear winner because their incentives are a lot
better aligned to mine.

---

### Image credits

Photo | Creator | Licence | Source
--- | --- | --- | ----
HTC Dream | Akela NDE | CC BY-SA 3.0 | [Wikimedia](https://commons.wikimedia.org/w/index.php?curid=6680413)
Pixel and Pixel XL | Maurizio Pesce from Milan, Italia | CC BY 2.0 | [Wikimedia](https://commons.wikimedia.org/w/index.php?curid=52110138)

[^1]: Or, at least, politely feigned surprise.

[^2]: And Google's terms were particularly onerous: as well as requiring Chrome and Google
      Search to be preinstalled, they prevented manufacturers from selling any devices
      powered by alternative versions of Android - e.g., Samsung wouldn't be allowed to
      sell a refrigerator that ran Amazon's FireOS. The European Union handed Google a
      $5,000,000,000 fine for this anti-competitive behaviour, which Google are still in
      the process of contesting.

