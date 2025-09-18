---
date: 2025-07-16
title: "How tech companies failed to build the Star Trek computer"
permalink: /how-tech-companies-failed-to-build-the-star-trek-computer/
tags: [opinion, ai]
format: long
resources:
  - src: enterprise-computer-room.jpg
    name: "Still from an episode of Star Trek: The Next Generation, with various characters stood around in a computer core room"
    title: "A computer core room on the Enterprise-D"
opengraph:
  title: "How tech companies failed to build the Star Trek computer · Chameth.com"
  type: article
  url: /how-tech-companies-failed-to-build-the-star-trek-computer/
  image: /how-tech-companies-failed-to-build-the-star-trek-computer/enterprise-computer-room.jpg
---

{% figure "right" "Still from an episode of Star Trek: The Next Generation, with various characters stood around in a computer core room" %}

In most Star Trek series, the ship or station computer is ever-present in the
background, waiting to be called on by the main characters[^1]. It nearly
always does exactly the right thing, and there's little limit to the functions
it can perform. Take this mundane example from DS9:

> KIRA: Computer, establish link with the Bajoran Medical Index for the Northwestern District. \
> COMPUTER: Link established. \
> KIRA: Access all information on Doctor Surmak Ren. \
> COMPUTER: There are no records matching that name. \
> KIRA: Try the Northeastern District, same search. \
> COMPUTER: Doctor Surmak Ren, currently serving as Chief Administrator of the Ilvian Medical Complex. \
> KIRA: Computer, open a channel to the Ilvian Medical Complex. Administrator's office.

The computer is doing some kind of networking to a database only identified by
name. It does a search and summarises the lack of results. It then repeats the
process with another database, and succinctly announces the results. Finally,
it opens a communication channel to a specific room in a facility, based only
on its name.

This whole interaction is remarkably boring[^2]. Kira doesn't have to know
any URLs or API endpoints, or what protocol she wants to use. She doesn't have
to open a specific app and then login and then try the query again. She just
says what she wants and the computer does it.

It seems like this should be one of the most easily obtainable bits of sci-fi
wizardry with our current technology. We have multiple massive companies
throwing lots of money at digital assistants, LLMs that are improving at an
insane rate, but we're somehow not even close to the usability or usefulness of
the Trek computers. What gives?

### Boring is, well, boring.

Larry Page once said something that might help explain it:

> The Star Trek computer doesn't seem that interesting. They ask it random
> questions, it thinks for a while. I think we can do better than that.

This is the same Larry Page that founded Google, whose mission statement is
"to organize the world's information and make it universally accessible and
useful". Of all people, surely he should find an omnipresent computer that can
answer 'random questions' interesting?! It seems like it should be the epitome
of Google's mission!

<!--more-->

Google's "better than that" seems to have been to stuff LLMs into every product
they can, even when you don't want them there. Even when they're worse than the
normal content they displace. These things look _exciting_ when they're part of
a scripted demo at Google I/O, but they fall flat and just get in the way when
they're exposed to the reality of day-to-day use.

The Star Trek computer is the opposite: it isn't snazzy, but it is genuinely
useful. That means it's not an attractive target for the company execs who want
marketing opportunities, and it's not appealing for engineers who need to
demonstrate "impact". But even if Google did try to make the Trek computer,
there are other problems…

### Assistants need to be free

A significant amount of tech companies' business models currently revolves
around trapping users in walled gardens. They want you using _their_ ecosystem;
that way they get more data from you, and you're more likely to spend more money
on their other offerings that work together. There's barely any incentive to
allow any kind of interoperability with other platforms outside carefully
contracted integrations.

I remember trying to help a family member move their photos from iCloud to
Google Photos. At one point they turned around and said, exasperated, "why is
this so hard? Aren't they both in the cloud?!". It's easy to dismiss that as
someone who hasn't quite grasped the fundamental idea that "the cloud" is just
someone else's computers, but that's not the whole story. There's no reason why
there shouldn't be a quick and easy transfer: both services already allow
uploading and downloading, there's just no incentive for the companies involved
to make it so[^3].

These kinds of misaligned incentives and walled garden business models cause
even more problems when it comes to digital assistants. Siri is basically never
going to be able to interact with, say, your Google Drive; ~~Bard~~ Gemini
is never going to be able to send a message via iMessage. Even when there are
appropriately blessed interactions, they're so clunky. Can you imagine Captain
Picard saying "Computer, ask the turbolift skill to take me to deck 5"?

### Someone else's computer

Software issues aside, there's still a key difference between the Star Trek
computers and our current batch of digital assistants: where they run. The
Trek computers are all housed within the ship or station they serve; they can
connect elsewhere to gather information, but they run entirely independently.
If they go wrong, a local engineer can go in and fix things. While some of our
assistants may have physical hardware in your home, they don't work without
a vast cloud apparatus behind them. If your Internet connection fails, they
become paperweights. If the company running them decide to remove some
functionality you depend on, you have no recourse.

That kind of helplessness isn't limited to assistants, either. There's a rapidly
growing trend of being unable to modify or repair hardware you fully own and
control. Part of this is just that they're becoming more complex: it's a lot
harder to replace a microchip than a gear, but companies are also going out
of their way to make it more difficult for users through draconian DRM
regimes[^4] and aggressive intellectual property enforcement. If the US Navy
can't repair their own equipment because a corporation says so, what hope do
consumers have?

We're approaching a point where you don't actually own anything. Software
is cloud and subscription based, hardware is unrepairable. Even cars can
be remotely updated and have features added or removed. The Federation wouldn't
allow a third party control over their ships[^5], so why are we so happy to
put up with it in everything we consume?

### A small ray of hope?

The most promising way of tackling all of these problems is through legislation.
The EU's [Digital Market Act](https://digital-markets-act.ec.europa.eu/index_en)
is an attempt to force 'gatekeepers' like Google, Apple and Meta, to allow
third-party access to their services. It seems like a pretty reasonable
approach, but the tech companies are unsurprisingly resisting it. Apple in
particular have gone out of their way to refuse to comply, and when forced to
do so have limited the functionality to people in Europe.
Still, the DMA is a promising start, and if similar legislation is introduced
(and robustly enforced) elsewhere it might start forcing companies to behave a
bit better.

There are also smaller companies that actually do the right thing.
[Framework](https://frame.work/gb/en) make laptops that are user-serviceable;
[Fairphone](https://www.fairphone.com/) do the same for mobile phones. Smaller
software companies provide useful, open APIs. The average person on the street
will probably have never heard of these, unfortunately, but they do still
exist. Maybe as the bigger tech companies tighten the screws more, people will
turn to alternatives like this? Or maybe we'll just keep accepting that our
computers work for everyone but us?

[^1]: Unless, of course, the computer is playing the role of the episode's
MacGuffin and has contracted space-computer-COVID or something, then it's a lot
less in-the-background.
[^2]: It's almost like it only exists to move the plot along.
[^3]: You can generally export your data, thanks to a combination of legislation
and efforts like Google's "Data Liberation Front", but I've never seen an export
format that could then just be imported into an equivalent commercial product.
[^4]: Oh, you've changed the screen on your iPhone? Better hope it can do the
secret handshake with the Apple hardware.
[^5]: I think there might actually have been an episode where that did in fact
happen. We'll just ignore that as a plot contrivance.