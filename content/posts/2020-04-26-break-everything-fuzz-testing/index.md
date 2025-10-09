---
date: 2020-04-26
title: How to break everything by fuzz testing
description: Or "I'm sorry, I didn't mean to take down your service"
tags: [development, testing]
format: long
permalink: /break-everything-fuzz-testing/

resources:
  - src: chimp.jpg
    name: Chimp sat at a typewriter
    title: Fuzz testing is a bit like the infinite monkey theorem, but instead of Shakespeare you get crashes.

opengraph:
  image: /break-everything-fuzz-testing/chimp.jpg
---

{% figure "left" "Chimp sat at a typewriter" %}

Fuzz testing, if you’re not aware, is a form of testing that uses procedurally generated random
inputs to see how a program behaves. For instance, if you were fuzz testing a web page renderer
you might generate a bunch of HTML - some valid, and some not - and make sure the rendering
process didn’t unexpectedly crash.

Fuzz testing doesn’t readily lend itself to all types of software, but it particularly shines
in cases where some kind of complex user input is accepted and processed in some way - like
the aforementioned web page renderer. I was recently adding a library to parse EXIF data to
images to an Internet-facing service and realised it was a perfect opportunity to do some fuzz
testing. Even if I didn’t find any issues, I’d improve my confidence that the library was safe
enough to expose to the Internet.

### Breaking my EXIF library

I wrote a quick harness to run [go-fuzz](https://github.com/dvyukov/go-fuzz) on the library,
and gave it some pre-existing demo files as sample input. The way go-fuzz works is that it
instruments your code and then mutates the inputs to try to improve the coverage. For example,
if I had some sample data that had an EXIF tag with a value of 1 then go-fuzz might change it
to a 2 and see if the code follows a different path. In most cases it won’t but when it does,
they tend to be very interesting cases.

<!--more-->

One of the first issues that go-fuzz found was that some values in a maker note field would cause
the library to panic (i.e., crash). This happened because there was a check to see if the first
six characters were “Nikon” and a null byte, without first checking to see if there were actually
six characters available. This is a kind of bug that doesn’t happen much with “real” data - as
the field is either not present or completed correctly - but could easily be exploited once this
code is exposed to the Internet.

Another interesting bug that go-fuzz found was that if a tag had a particularly large count, the
library would try to allocate an obscene amount of memory and die. There was already a check in
the code that was meant to avoid this exact scenario, but go-fuzz managed to find a way around
it. Each tag has a size (for example an integer tag takes a fixed number of bytes) and a count;
the existing check multiplied the two together and made sure that the result wasn’t too large.
For most cases this was fine but go-fuzz found a case where the count was so large that when
multiplied by the size of the tag it overflowed the integer and became negative, thus passing
the sanity check but then subsequently failing when it came around to actually allocating the
memory.

The final bug of note that go-fuzz found was the most interesting. EXIF data is stored in IFDs
(“Image File Directories”), and each IFD provides what is effectively a pointer (a byte offset)
to the next one. The EXIF library already had a check to make sure that these didn’t loop, but
it only checked the immediately preceding IFD - so if IFD 1 linked to IFD 2, it would catch IFD 2
linking back to IFD 1 and break the loop with an error. Go-fuzz found that having three interlinked
IFDs had the same issue, though, and the guard code wasn’t triggered. This created an infinite
loop, maxing out a CPU core until the process was eventually killed - one of the worst kind of
bugs you could have in an Internet-facing service which doesn’t deal with private data! The fix
for this was fairly straightforward - I just made the library keep a record of the previously
visited IFDs and bail out if it found a loop.

### Breaking my IDE

When go-fuzz detects an issue it outputs not only the details of the problem (the stack trace,
error message, and so forth) but also the input that generated the problem. This is useful for
reproducing and making sure the issue is fixed, but it also makes it really easy to write
a test to ensure that the behaviour never regresses in the future.

As I was working through fixing the bugs that go-fuzz found, I dutifully added new tests where
needed. After adding the sample input with looping IFDs to the project, I switched to IDEA to
write a test to use it. I clicked on the input file to copy the file name, and then the entire
IDE hung and had to be restarted. Uh oh! When I restarted IDEA, it immediately began indexing
the project and again hung. It turns out IDEA parses EXIF data (presumably, even if it does
nothing else with the data, to get the rotation property for images), and the library they use -
an independent one written on Java - had the same bug as the Go library I was using.

In order to stop IDEA from indexing the file and becoming unusable I renamed it from a ‘.tif’
extension to ‘.dat’, and everything went back to normal. I thought I’d best report the bug to
JetBrains, though, so they could put a proper fix in.

### Breaking YouTrack

JetBrains use their own issue tracker called YouTrack for reporting bugs in IDEA. I dutifully
went over and described the problem, attaching the log files from the IDE, a description of how
the file was malformed, and carefully selected the .dat version of the file to upload so that it
wouldn’t cause anyone else the same immediate problem.

After trying to upload the file I got a strange error back. Uh oh! I submitted the IDEA issue as
it stood, unable to see if the attachments had even uploaded, and went and wrote up an issue for
YouTrack itself about the error message. While I was doing that, YouTrack seemed to slow down and
become really annoying to use. I had a sinking feeling the exact same thing was probably
happening as with IDEA and my library - but this time YouTrack had content-sniffed the file
instead of relying on the file extension. In hindsight, I should've put the file in a passworded
archive to ensure no automated tools got hold of it. I marked the issue as a security problem as in
a service like YouTrack it presents a denial-of-service opportunity[^1] (remember when I said it was
one of the worst kinds of bugs you could have in an Internet-facing service?...)

Shortly after I raised my YouTrack ticket, a notice appeared at the top of the page saying they
were investigating the current performance issues. Uh oh! I was holding out hope that this was
unrelated to me uploading the buggy dat file, but the timing all seemed a bit suspect. I shot
support an e-mail saying I think I might be the root cause for their performance issues and
linked to the ticket. In the time it took me to e-mail them, the entire site had been
put into maintenance mode. I got an e-mail back a few hours later confirming the outage
was in fact all my fault, as I’d feared. Within the space of days the JetBrains security team
had fixed the issue in YouTrack, which was a pretty nice turnaround.

So if you were trying to access YouTrack at the start of March and couldn’t - I’m sorry, I didn’t
mean to! Also, if you’re building an Internet-facing service that takes user input you should
really consider running a fuzz tester against it!

[^1]: "We have a problem". "Remember, there are no such things as problems,
      only opportunities". "Well then we have a DDoS opportunity."
      -- [@J4vv4d](https://twitter.com/J4vv4D/status/671090709588496384) 
