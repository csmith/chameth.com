---
date: 2025-05-28
title: Coming around on LLMs
permalink: /coming-around-on-llms/
tags: [opinion, ai, development]
format: long
resources:
  - src: claude-hello.png
    name: A screenshot of the Claude web UI with the prompt 'Say "Hello!"'. The response is "Hello!" 
    title: Claude says hi.
  - src: code-session.png
    name: A screenshot of claude-code. I ask it for the permalink for the latest post, and it gives a wrong answer. After prompting it again it gets it right.
    title: A simple example of a claude-code session
opengraph:
  title: "Coming around on LLMs · Chameth.com"
  type: article
  url: /coming-around-on-llms/
  image: /coming-around-on-llms/code-session.png
---

{%figure "right" "A screenshot of the Claude web UI with the prompt 'Say \"Hello!\"'. The response is \"Hello!\"" %}

For a long time I've been a sceptic of LLMs and how they're being used and
marketed. I tried ChatGPT when it first launched, and was totally underwhelmed.
Don't get me wrong: I find the technology damn impressive, but I just couldn't
see any use for it.

Recently I've seen more and more comments along the lines of "people who
criticise LLMs haven't used the latest models", and a good number of developers
that I respect have said they use coding models in some capacity. So it seemed
like it was time to give them another shake.

The first decision to make was which model to try. OpenAI are no longer the
only player in the game, every tech company of a certain size is now also
somehow an AI company[^1]. I looked a bit at some benchmarks, and then mostly
ignored them and went with the only company that I didn't outright hate:
Anthropic, and their model Claude.

### Initial impressions

The latest Claude models do feel a lot more "capable" than the earlier ChatGPT
versions I remember, but they also still have a lot of the same problems. At
their heart, they're still text-prediction models, and still seem to be trained
to predict text that will please the user rather than be factually accurate or
useful.

<!--more-->

One of the improvements that people kept mentioning was the ability for models
to access the web directly. I've seen it kick in a bit naturally, and asked
for it explicitly sometimes, and it's… nothing special? Are people just really
bad at searching the web? Maybe they should try [Kagi](https://kagi.com/)? I
feel like about 80% of the time I could have found the information just as
quickly myself, about 10% of the time it searched and then hallucinated an
answer, and the remaining 10% it found something quicker than I otherwise
would have. Those aren't great results, especially when you consider how
much cheaper and simpler just searching the web is.

There is another way to trigger web searches in Claude: using "research mode".
When you enable it, it splits off into different models: a "lead researcher" to
come up with a plan, and then some minions that execute it. It ends up doing
hundreds of web queries in the span of a few seconds. Suddenly I understand why
things like [Anubis](https://anubis.techaro.lol/) need to exist. I've not used
the "please launch a DoS attack" button since.

What did impress me, though, was its ability to churn out reasonable-ish
code. It can hack together a bash script as well as I can, and do it far
faster than I'd be able to. Sometimes they even work. That made me wonder
what it would be like doing actual coding with it. Anthropic have a CLI tool
called `claude-code`, so I paid them lots of money and gave it a spin.

### Coding with Claude

The first thing I notice about `claude-code` is that I really like the
interface. It's basically an input box in a terminal. It doesn't force me to
use a certain IDE or do things in a certain way. By default it asks before
making any changes, showing you a side-by-side diff of what it's doing and
allowing you to provide feedback. If I had to design a way to interact with
a coding agent from scratch, I can't think of many things I'd improve.

{% img "A screenshot of claude-code. I ask it for the permalink for the latest post, and it gives a wrong answer. After prompting it again it gets it right." %}

The power in `claude-code` versus just using the web UI is that it can use
tools. It can query `git`, run `grep` commands, even use `sed` if it wants to
change something in lots of files at once. It's very good at figuring out its
way around a codebase, even without any explicit instructions. You can see in
the screenshot that with a little prompting it managed to get the permalink
to this post; I didn't tell it where the posts were stored, or how to work out
the latest, I just told it when it was wrong. If I'd run `/init` before it
would have probably picked up on the fact that my posts have custom permalinks,
and noted it in `CLAUDE.md`.

But how good is it at actually writing code? It's like having a keen but not
particularly thorough Junior Engineer at your beck and call. If you give it
a clearly defined task and guidance on how to implement it (and maybe some
feedback as it suggests changes), it's more than capable of doing it. If you
don't give it enough guidance it tends to go more off the rails. I tried
having it generate a simple application from scratch with minimal technical
guidance and no review of what it was doing, and it made such a mess of it I
decided it was quicker to throw it away and start again by hand.

### The man behind the curtain

Even with sufficient guidance, at times it's _really_ obvious that it's an
LLM generating pleasing-token-strings and not something that genuinely
understands what it's doing. It will spit out code like this:

```go
if err != nil {
    if err == sql.ErrNoRows {
       return nil, err
    }
    return nil, err
}
```

Why's that check for `sql.ErrNoRows` there? It's entirely pointless. There was
no instruction to check for it, none of the existing code checked for it, but
I assume it comes up quite a bit in the training data. But it didn't
_understand_ why, so it put the check in, and returned the exact same thing as
if it hadn't.

It also occasionally tries to "cheat" or solve problems the wrong way. It
sometimes feels a bit like you're asking for wishes from a Monkey Paw. "Stop
the unit tests failing", you'll say; Claude will respond with a request to
delete the failing test. The man behind the curtain isn't particularly well
hidden, and the training data and pattern matching often shows through.

This lack of understanding also makes it very hard to get Claude to use
comments in a sensible way. It absolutely loves doing nonsense like this:

```javascript
// render form
function renderForm() {
    // ...
}
```

Even with explicit instructions to avoid useless comments, or comments that only
explain what the code does (not why it does it), or other prompts. Again, I'm
chalking it up to a mixture of training data that does that, and lack of any
actual understanding about why a human developer may want a comment to exist.
Of course, there are plenty of flesh-and-blood devs out there that also don't
comment effectively, so maybe I shouldn't give Claude too much of a hard time
on this.

### … and yet …

So with all those problems, it sounds like it's just not worth it, right?
Well, not quite. I feel like programming mostly consists of two distinct tasks:
thinking how to implement things; actually implementing them; and then debugging
why you're off-by-one somewhere. Letting an LLM do the thinking is a big no-go
for me: it's evidently not good at it, and frankly that's one of the things
I most like about programming. Having let it loose on small projects, I shudder
to think what the codebases of all the "vibe coded" projects that are popping up
are like.

For the second stage, though, it's actually quite nice. If I've thought through
how I want something to be implemented, I can feed those steps to Claude and have
it churn out the otherwise not-too-interesting code. This requires far less
time and concentration on my part than writing code, to the extent that I can
be thinking about the next feature, or doing something else at the same time. I
don't feel like I'm going to end up deskilling myself this way, as I've already
formed the idea of what I want to code, I'm just using the LLM to spit it out
faster than I can type it.

A lot of the most boring bits of coding like implementing CRUD-y operations can
be summarised as "Look at this file/function. Do the same thing but slightly
differently elsewhere". Claude is great at this, especially when explicitly
prompted like that. You're basically playing to its pattern-matching strengths,
rather than asking it to come up with anything novel. Most of my prompts tend
to be prefixed with "Look at @some_file.go and @other_file.go." to cue up the
patterns I want it to use, and I find this works well.

As for debugging, it's a mixed bag. It's sometimes amazingly insightful, and
sometimes just runs around in circles trying the wrong things over and over.
It's worth asking the question, but I definitely wouldn't rely on it over my
own abilities.

### The future

I'm probably going to carry on using `claude-code`, at least for personal
projects. I've got so much done that I just wouldn't have been
_bothered_ to do if I was doing it all by hand. I very much enjoy being in
the more "diffuse thinking" mindset, planning how things are going to work,
rather than being stuck in the mines digging out SQL queries. After all, who
wouldn't want an over-eager assistant to work on all their hobby projects?

Work is a slightly different matter: a private project or even an open source
project that disclaims any liability is different to something I'm being
paid to deliver, and bear responsibility for fixing if it's not done correctly.
I'm not saying I won't use it at all, but if I do it'll be much more constrained
than I would in personal projects.

As for non-code usages: I'm not sold. My sceptic hat is still firmly in place.
I don't think chat is a particularly good interface for many things, and
hallucinations are still a big problem despite what people say. I hate the tide
of AI slop that's taking over the Internet, and how LLM-powered chat
agents are being forced into every random product. You're definitely not going
to be seeing any AI-generated blog posts from me!

One topic I've not gone into here is the ethical concerns about using LLMs. They
obviously exist, and I do have thoughts, but that's a topic for another day.

[^1]: I'm surprised it's not gone more mainstream: why's there
no Tesco Value LLM model, yet?
