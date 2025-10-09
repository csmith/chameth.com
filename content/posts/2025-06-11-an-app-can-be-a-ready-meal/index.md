---
date: 2025-06-11
title: "An app can be a ready meal"
permalink: /an-app-can-be-a-ready-meal/
tags: [personal, development, ai]
format: long
resources:
  - src: readymeal.jpg
    name: "A spaghetti carbonara ready meal, fresh out the microwave"
    title: "It's not a home-cooked meal, but it does the job sometimes."
  - src: is_it_worth_the_time.png
    name: "'Is It Worth the Time?' comic from XKCD. A table of 'How often you do the task' vs 'How much time you shave off', with values showing how long you can work on making a task more efficient before spending more time than you save."
    title: "XKCD 1205"
opengraph:
  image: /an-app-can-be-a-ready-meal/readymeal.jpg
---

{% figure "right" "A spaghetti carbonara ready meal, fresh out the microwave" %}

Three years ago I read "[an app can be a home-cooked meal](https://www.robinsloan.com/notes/home-cooked-app/)"
by Robin Sloan. It's a great article about how Robin cooked up an app for his
family to replace a commercial one that died. It's been stuck in my head ever since.
It's only recently that I've actually done anything like Robin described,
though. Part of the reason was my brain got too hung up on the family aspect:
in my head, a home-cooked meal is one where your family or friends all gather
around to eat it with you (in much the same way as Robin's app is used in
the article). It took me an embarrassingly long time to realise that you can
apply all the same arguments to an app built just for you. And it doesn't even
have to be difficult. In fact, it can be more like a ready meal[^1] than a
family dinner.

### Why not open source?

I love open source software. Almost everything I use day-to-day is open source,
and most things I write for myself I release as open source. I believe that
should be the default stance for most software. So why would you want to make
something and keep it just for yourself?

<!--more-->

Even if an open source project garners no users whatsoever, there's still some
pressure for it to meet certain standards[^2]: if it needs configuring, then there
needs to be a mechanism for people to do that; the code needs to be of a
reasonable quality --- my open source projects are linked from my CV, potential
clients might be looking at them --- and there needs to be at least some attempt
at documentation or showing other people how to use it. There's "building in
public" and then there's "inviting the whole world to poke around your
drawer of shame".

The way I usually work on things that are mostly for me is I hack them together,
and then I gradually force myself to fix them up into a state where I consider
them acceptable for release. That stage isn't fun, and sometimes it adds more
complexity than the whole original project. I'm not saying open source isn't
worth it --- far from it! --- but there is definitely a balance to be struck
between the utility of making something open source and the amount of effort it
takes.

For example, [I mentioned previously](/home-automation-without-megacorps/)
that I wrote a home automation system. Hidden within that is a ~40 line
function that determines whether a fan should be turned on to keep a room cool.
The logic is entirely unique to the devices on my network and my requirements.
To try to put it into a form that would be useful to anyone but me would take
exponentially more effort. I'd basically be re-inventing Home Assistant, and
I explicitly started that project because I didn't need that kind of complexity.
For a while I had this nagging feeling that I should find a way to open source
it. Then I had my[^3] "lightbulb" moment: I could just _not_.

Writing something entirely for yourself, without planning on open sourcing it
is surprisingly liberating: you don't need to worry about documenting things;
you can change anything and everything at will without having to worry about
migration paths or whether it'll break anyone's workflow; you don't have to
configure things, you can just code them how you want. Hell, you can hard-code
API keys right in the source code. Who cares!

### The microwave revolution

{% figure "right" "'Is It Worth the Time?' comic from XKCD. A table of 'How often you do the task' vs 'How much time you shave off', with values showing how long you can work on making a task more efficient before spending more time than you save." %}

One of the other problems with making software just for yourself is that often
the time investment just isn't worth it. If I spend a day automating something
that normally takes me five minutes, it's going to take an awful long time to
"break even" on that time spent. If you're open sourcing something then there
are ancillary benefits that may tip those scales: you might help other people,
have something to show off, etc. But if it's just for you, then XKCD 1205 is a
harsh mistress.

Recently, though, a new tool has emerged that tips the balance: LLM assistants.
[I've talked about my experience before](/coming-around-on-llms/). I'm definitely
not comfortable letting an LLM run roughshod over my published code, even if it
can help write some of it, but for private code? Why not! For a trivial example:
I have a USB key with an Arch Linux ISO on it, in case I need to troubleshoot
or reinstall my PC. Every so often I'll update the ISO to the latest version.
It probably takes 5 minutes to do, so it's probably only worth 25-50 minutes
to optimise it. Can I write, test, and debug a script to do it myself in that
time? Probably not. Can I get an LLM to generate some code to do it for me, and
then spend 10 minutes reviewing it[^4]? Easily.

To continue with the strained metaphor: LLMs are like the microwave that nukes
your ready meal. Pop in a prompt, let it use a load of power for a while, and
out pops your app. You can even do it on your phone while sat on a sofa[^5]. 
That changes the time equation even more: you can just squeeze in a prompt
whenever something comes to mind.

This also goes back to the open source issue: I don't feel comfortable
publishing something created by an LLM with minimal intervention on my part.
If I can prompt an LLM to do something, so can anyone else. We don't need more
slop out there, and I don't want there to be any confusion about what I've
written and what I've just prompted into existence. But for personal projects
that will never see the light of day it's perfect.

### What I've made

Besides the home automation controller, I've built a constellation of smaller
tools: a script to arrange windows on my monitors how I want them[^6], one to
help me verify that my backups are working and can be restored, one to create
the files/folders for a new blog post, and some other odds and ends. They're
all pretty small, and pretty specific to my setup.

My biggest just-for-me project is a web app that I started to help me aggregate film recommendations.
It's since morphed into a general personal data aggregation service: it deals
with data from GitHub, Todoist, Letterboxd, TMDB, Healthkit, and others. It also lets
me make re-orderable lists, store recipes, and more. Parts of this could
definitely be open sourced, and I might carve them out at some point, but it's
mostly a glorious hodge-podge of things specific to me. Having all these
services in one place lets me make quick and dirty automations, for example:
when I create a Todoist note on my phone or watch, I often forget to set the
due date, so it doesn't show up in the "Today" view. It was literally a few
lines of code to plumb things together so any inbox task without a due date
gets set to today automatically.

If I was trying to write these things in a way that would be useful to other
people, or --- for some of the features --- without the aid of an LLM, I just
wouldn't be bothered. It'd take too much time for questionable benefit. But
there's a kind of joy in just being able to hack things together that work just
well enough for you. A ready meal will never compete with a home-cooked meal,
but sometimes it perfectly hits the spot.

---

| Image credits       | Creator             | Licence      | Source                                                                                  |
|---------------------|---------------------|--------------|-----------------------------------------------------------------------------------------|
| Spaghetti carbonara | Wikimedia user Geni | CC BY-SA 4.0 | [Wikimedia](https://commons.wikimedia.org/wiki/File:Spaghetti_carbonara_ready_meal.JPG) |
| XKCD 1205           | Randall Munroe      | CC BY-NC 2.5 | [XKCD](https://xkcd.com/1205/)                                                          |

[^1]: or a "TV dinner", if you're North-American-ly inclined.
[^2]: Self-imposed pressure, for sure, but brains are going to do brain things.
[^3]: very obvious in retrospect
[^4]: Reviewing LLM-generated code that shells out to `dd` to write to the
root of a storage drive is perhaps the most intensely I've ever reviewed any
code to date. I really didn't want to write a blog post about how an LLM
blatted my hard drive!
[^5]: Phones are terrible input devices for code, but they're just about
passable for typing English.
[^6]: I basically want a tiling window manager, but without all the effort and
weirdness of a tiling window manager.