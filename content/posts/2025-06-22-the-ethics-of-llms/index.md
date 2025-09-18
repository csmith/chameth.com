---
date: 2025-06-22
title: "The Ethics of LLMs"
permalink: /the-ethics-of-llms/
tags: [opinion, personal, ai]
format: long
opengraph:
  title: "The Ethics of LLMs · Chameth.com"
  type: article
  url: /the-ethics-of-llms/
---

I've written about LLMs a few times recently, carefully dodging the issue of
ethics each time. I didn't want to bog down the other posts
with it, and I wanted some time to think over the issues. Now I've had time
to think, it's time to remove my head from the sand. There are a lot of
different angles to consider, and a lot of it is more nuanced than is often
presented. It's not all doom and gloom, and it's also not the most amazing
thing since sliced bread. Who would have thought?

It's worth noting that I'm just setting out my position here. I'm not trying
to convince you to change your mind. To set the scene a bit, I mainly use
Claude Code as a programming tool. I use the Claude chat interface sometimes to
proofread things, do random one-off data analysis, or help organise things.
More rarely I'll try to use it to brainstorm things, or recommend things,
or do more "creative" things, but I don't trust it enough in those domains
to do it often. I don't use it for research or as a Google replacement[^1],
which I recognise probably makes me a weird half-in-the-water, half-out-of-it 
class of user.

### Copyright & Corporate Control

One of the key issues, and something that is being prosecuted in several court
cases right now, is how LLMs interact with the copyright system. And by
"interact with" I mean "run roughshod all over". It seems pretty obvious from
my lay perspective that if having 10 seconds of pop music in the background of
a YouTube video is copyright infringement, then
[Meta pirating books via BitTorrent](https://www.theguardian.com/technology/2025/jan/10/mark-zuckerberg-meta-books-ai-models-sarah-silverman)
must also be.

<!--more-->

That said, the entire copyright system as it exists now is fundamentally
broken. I don't mind the idea of individual creators having protection, but I absolutely hate
how copyright is wielded as a blunt instrument by massive corporations to
intimidate individuals and to try to earn more money. If a law enables a
huge conglomerate to sue the author of an open source project for a
[theoretical 75 trillion dollars](https://en.wikipedia.org/wiki/Arista_Records_LLC_v._Lime_Group_LLC)
then that is a bad law. Copyright today has been twisted well beyond what it
was meant to be originally, thanks to lobbying by people and businesses with
lots of money that want to keep it.

So if copyright law is not fit for purpose, are LLMs fine? Well, not quite.
The argument that LLM training is a bit like a child learning to read holds
some merit, but it falls flat when LLMs can regurgitate the input material
verbatim[^2]. Personally I don't care if they steal all the content of the
New York Times and all the Disney characters that are still in copyright. It's
the individual harm that bothers me: the struggling author whose words are
slurped up by the LLMs and regurgitated, potentially costing them readers; the
open source dev who chooses a copyleft licence on their work, only for an LLM
to dump complete copies of functions in someone else's codebase with no
attribution or licence, stopping any future improvements from being released
for the public good.

I can mostly mitigate those concerns in my personal use. When coding I generally
direct the agent closely enough that the opportunity to drop in massive chunks
of someone else's code is negligible. I don't use LLMs to generate big chunks
of written content. Overall I'm reasonably happy that I'm not making anyone
worse off.

One related area that doesn't affect me much because of my limited use, but does
concern me a lot, is corporate control. Most of the frontier LLMs are built by
big tech companies. As more and more people are using LLMs, they become an
avenue for control over information. It's clear the people in charge should not
be trusted with that responsibility, as
[Grok's whole Boer War thing demonstrated](https://www.theguardian.com/technology/2025/may/14/elon-musk-grok-white-genocide).
I'm not sure what the answer is, though. Would a government built LLM be any
better? Probably not. The requirements to actually do the training are so high
that the only people who can do it are the ones you wouldn't want to.

So I'm left in the uncomfortable position of using tools by companies I don't
trust, trained on data they shouldn't have been trained on. Still, there can't
be many more ethical issues, right?

### Forced Features & Flattery

Before we get onto the more obvious ethical topics, let's go on a brief detour
so I can complain about how LLMs are being crowbarred into seemingly **EVERY**
**SINGLE** **SERVICE**. I don't want an LLM in my text editor. Or when shopping.
Or taking screenshots of everything my computer does every few seconds. I don't
want LLMs anywhere outside a clearly-defined LLM box. With these forced
integrations, you're even more in the dark about what their prompts are,
what information they have access to, what underlying model they're using,
and so on.

There's no good reason for 90% of these implementations; there's just some
weird Silicon Valley hype train that everyone is scared to miss out on. And
as a result we get more and more screen real estate taken up with things with
animated gradients that can't be turned off and are prompted to talk in an
insufferable manner.

The problem isn't just that they're everywhere, though, it's how they behave
when they're there. The overly sycophantic behaviour that most LLMs adopt can
hook users in the same way a social media app can fiddle with its algorithms to
encourage people to keep scrolling. It's a lot harder to dismiss an LLM when
it's blowing smoke up your ass. Now I don't believe this behaviour is actually
intentionally nefarious, but more of an artifact of the models overfitting on
user approval instead of something useful like accuracy.

The combination of going too far to please the user and hallucinating things
is incredibly problematic. How are ordinary, non-tech savvy, people meant
to deal with these models being forced in front of them? If you ask a question
and get back a flattering answer that's roughly the shape of what you were
expecting, would you question it? Would a student, who was doing their homework
in Google Docs when an LLM inserted itself into their lives question it?

This isn't just a problem for naive users, either: I occasionally ask Claude
Code "Can we do X?" and discover much later that the answer should either be an
emphatic "NO", or at least have a page of caveats attached. But the LLM wouldn't
want to displease me, so it tries to do what's asked of it, with all kinds of
weird hacks, and ultimately gets itself tied up in knots. This is not how a tool
should behave. If I wanted someone to say "Yes, and" repeatedly I'd join an improv troupe.

### Environmental Expenses

The biggest ethical issue with LLMs is their environmental impact. Are we
adding more fuel to the dumpster fire that is our planet, just so my phone
can inaccurately summarise news headlines for me? The answer is a resounding
"maybe". The companies making the LLM models are cagey about how much energy
goes into making them and then answering queries. There are estimates for
the latter, but there are several orders of magnitude between the most
optimistic and most pessimistic.
LLMs are definitely not _good_ for the environment. Basically no computer activity
is[^3]. But you don't see people campaigning to shutter YouTube (for example)
because of its environmental impact. I struggle to see the difference.

I'm not saying considering the environmental impact isn't important.
We are killing the planet. We are not doing enough about it. More attention
to the climate crisis in general is a good thing. But even with the most
pessimistic estimates on power usage, my personal LLM usage is a drop in
the ocean. I can have far more impact by reducing the amount I travel,
avoiding red meat, and so on.

One issue I do have, though, is the forced integrations I mentioned. I'm happy
with _my_ personal usage. I'm not happy that a whole slew of LLMs are being
operated on my behalf. How much power is wasted by Google generating LLM
results in every search results page, versus how much benefit it provides?
How do people know if a search box will just use a few Watts doing a database
search, or burn kilowatts[^4] spinning up an LLM and filling its context window?

Then there's the issue with training. Who knows what resources go into training
these models? Whether the amortised cost of that across all its users is worth
it? I think the answer to the first question is "only the tech companies making
the models", and their utter silence on the second question says volumes.

But this kind of industrial power isn't unique to LLMs. There are many
industries that do far worse. And that's not a justification, but I think it
makes sense to look at them more holistically. In an ideal world, where
we have functioning governments that listen to scientists[^5], we could address
all of this with legislation on emissions and clean power. Then all of these
industries will either manage to function in a manner compatible with keeping
the planet alive, or will have to stop. I appreciate that actually legislating
that is about as likely as waking up one day to find Claude has become an AGI,
though.

### Slop & Survival

We're nearly done, I promise. You can see why I didn't want to try to tackle
these issues in an earlier post! The last big issue I want to talk about is
slop. It feels like it's everywhere, and still getting worse.

The Internet has slowly morphed from a wonderful place full of individual,
quirky sites, into a desert of bland corporate silos, and now into a wasteland
of AI slop. It's becoming more and more difficult to find original information.
There's this conspiracy theory called "Dead Internet Theory" that basically
says nearly all traffic on the Internet is generated by bots as part of a
co-ordinated effort to manipulate us. Obviously it's unhinged, but at the
same time some elements of it are not that far off the mark. How long will it
be before the Internet is so drowned in slop that it is, effectively, dead?

The other issue is what happens to the human jobs that have been replaced
with slop? It's way too easy for a money-pinching publication to get rid of
their human staff in favour of pushing a few buttons and publishing the output.
You'd hope that market forces would balance things out, but there doesn't seem
to be much sign of that happening. There are small movements of people putting
badges and other marks on their work to show it was made by a human, perhaps 
that will take off?

On a more personal level: what happens to software developers? As it stands,
code agents are nowhere near good enough to replace a human. At least not if
you want a non-trivial output that works, can be changed in the future, and
doesn't have massive issues. Even if they get substantially better, I feel
like you'd still need a knowledgeable human in the loop to keep a rein on
everything. Does that mean the world needs far fewer software engineers?
Probably. But it doesn't mean we're all going to be out of jobs overnight
as some people seem to be predicting. The only way I can see that ever
happening is if there's a radical shift in how software is made: a new language
of some kind that's uniquely suited to LLMs, or new ways of composing software
together from smaller parts[^6].

So… Yeah. Where does that leave us? I've written several thousand words about
how LLMs are an ethical mess, but I'm going to carry on using them. That's uncomfortable.
But it's not unlike using Amazon: they're not a great company,
I'd rather give the money to a local shop or supplier, but I'm also not going
to massively inconvenience myself by being quixotic about it. Life is about
compromises like this, and it'd be silly to think otherwise. I'll stand on
morals on some things[^7], but not to the extent of becoming a reclusive old
crank shouting at the clouds.

[^1]: That's what [Kagi](https://kagi.com/) is for. It's great. You should try it.
[^2]: Or, maybe even worse, with added hallucinations.
[^3]: Hell, basically no _human_ activity is good for the environment.
[^4]: ±several kilowatts, because who knows?
[^5]: and didn't elect people who withdraw from climate agreements, re-open coal plants, and so on…
[^6]: I don't _think_ I'm just saying this to reassure myself…
[^7]: I'll grudgingly give Jeff Bezos money, but I sure as hell am not going to use any service from Elon Musk.