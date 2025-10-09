---
date: 2024-08-04
title: Further adventures in 3D printing
permalink: /further-adventures-in-3d-printing/
tags: [printing]
format: long
resources:
 - src: p1s.png
   name: Marketing image of the Bambu Labs P1S
   title: The Bambu Labs P1S.
 - src: comparison.png
   name: A side-by-side comparison of printing time on the P1S and SV06. The P1S prints 8 objects at higher quality than the SV06 prints 3.
   title: P1S vs SV06.
opengraph:
  image: /further-adventures-in-3d-printing/p1s.png
---

{%figure "right" "Marketing image of the Bambu Labs P1S" %}

Not quite a year ago, I bought a Sovol SV06 3D printer and [wrote about](/adventures-in-3d-printing/)
my initial experiences. At the end I joked:

> I’m sure there will be more problems, but hopefully I get at least a week or
> so before anything else goes wrong. Maybe I’ll print an “X days since the
> last SV06 problem” counter…

That turned out to be more prophetic than I expected. It became a running joke
that I'd flip the imaginary "days since" sign back to 0 almost every time I
printed. I still really enjoyed having a printer, though: it's so useful to
be able to be able to go from an idea to a CAD drawing to a physical object
in the space of minutes or hours. So when Bambu Labs reduced the price of the
[P1S](https://uk.store.bambulab.com/products/p1s) --- a printer generally regarded
as about as trouble free as you can get --- to £522 including shipping, I splurged on one.

### Concerns

I make it sound like that was an easy decision, but I actually ummed and ahed
a lot before pulling the trigger. I had a lot of objections to a Bambu Labs
printer, and wasn't quite sure if the price made up for them.

<!--more-->

The Bambu Labs printers are designed to be more consumer friendly than most
previous 3D printers. You can start a model printing directly from their
website or mobile app, which employs a cloud-based slicer that then sends the
model down to your printer. This is one of those things that sounds amazing for
consumers, but should terrify anyone who works with computers. There are so many
ways for a "cloud-connected" printer to go wrong, and with such potentially
disastrous results it gives me anxiety just thinking about it.

This is not just theoretical, either. In a blog post with the rather under-stated
title of ["Initial Investigation in the Bambu Cloud Temporary Outage"](https://blog.bambulab.com/cloud-temporary-outage-investigation/),
Bambu describe how a server-side error caused printers to start printing unprompted.
Users on Reddit reported printers starting in the middle of the night, trying
to print on top of other objects that were left on the build plate, and so on.
3D printers can easily start fires; having them cloud controlled is frankly
insane.

As a result of that "temporary outage", Bambu further improved the "LAN mode"
for their printers, which allows you to sever its connection from their servers
and control it over the network. That's basically how my old printer worked
anyway, so I was happy enough that I could enable LAN mode and ignore all the
cloud nonsense.

My other objection to Bambu was more nebulous. Almost all 3D printers are built
upon open source software, and some companies like Prusa go out of their way to
make their products more open to give back to the community. Bambu don't really
do any of that: I'd be very surprised if they _hadn't_ incorporated a bunch of
open source software into their printers and just ignored the licences[^1]. I'd
much rather give my money to a company doing good, but the only equivalent
printer was the [Prusa MK4](https://www.prusa3d.com/product/original-prusa-mk4-2/)
which was _double_ the price and didn't come with an enclosure. To add an
enclosure kit would have made it £800 more expensive than the P1S, and I just
can't justify that.

### Experience

The initial setup of the P1S was a breeze. While the SV06 shipped in parts, the
P1S came fully assembled and just needed some packaging and retaining screws
removed. These were all clearly marked with large coloured arrows. It's very
much designed so anyone can do it.

I ran through the initial calibration, firmware update, and then put it
in LAN mode. Then I just started printing and it went fine. No messing around
levelling a bed, or using a sheet of paper to try and set the Z-offset properly;
it just worked. It was _really_ loud though. After talking a bit with a friend
who'd also bought a P1S, I realised that most of his noise was from the fans,
while most of mine was from the servos. A bit of research later and I found out
they added servo noise compensation in a firmware update, but I hadn't run the
calibration for it because the initial calibration happened before the update.
After doing that, the printer went from sounding like it was impersonating a
continental police siren to just being a bit fan-y.

The biggest difference over the SV06 is the speed. I'd swapped the nozzle on the
SV06 to a 0.6mm one, and generally printed at a layer height of 0.4mm. The P1S
still has its 0.4mm nozzle, and I mostly print at a layer height of 0.2mm. So
it's laying down twice the number of layers, with a nozzle that can extrude
half the amount of plastic… And it's at least twice as fast! Prints that I
previously wouldn't have even attempted on the SV06 are now 8-10 hours on the
P1S, without even adjusting the speed or lowering the quality. Here's a
side-by-side comparison:

{%img "A side-by-side comparison of printing time on the P1S and SV06. The P1S prints 8 objects at higher quality than the SV06 prints 3."%}

The P1S can print eight objects because of its slightly larger bed, while the
SV06 can only manage three. The P1S is also printing at a much higher quality.
Yet somehow it still ends up being faster. It's mind blowing.

The other big difference I've not mentioned yet is that the P1S is a "core XY"
printer rather than a "bed slinger". That means in order to move along the
Y-axis (from the front of the bed to the back), the P1S moves the extruder
while the SV06 moves the bed. This makes a big difference when printing tall
objects, as they're far more likely to wobble and come loose when the bed is
moving back and forth. It also presumably makes the kinematics a lot easier
as the mass of the moving parts is constant, instead of gradually becoming
heavier as more and more plastic gets added.

### The only failures

I've only had a handful of failures since getting the P1S. I had one weird
issue where the slicer decided to print some parts extremely slowly, and the
print just didn't work out in those parts; that's more of a slicer issue,
though: after tweaking the model to stop the weird behaviour it printed fine.

The only failure I can directly attribute to the printer was when I tried
ironing. This is a process where after printing the top layer, the printer
slowly goes back over it at the same height while extruding a small amount of
plastic. This is meant to fill in the small gaps between extrusions and give
a smooth finish. In my case, though, it just clogged the nozzle and extruder.

At first I thought it was because I was printing too hot: the profile I was
using was designed for fast printing, so had to melt the filament rapidly, but
there was no way to turn that temperature down just for the ironing pass. I
tried another run at a lower temperature and everything clogged again. I then
swore off ironing, did a normal print, and everything clogged once more.

At this point I was getting exceedingly good at disassembling the hot end, but
also fairly annoyed. It felt like I was using an SV06 again[^2]! I'd previously
read some recommendations that you should leave the door or top of the enclosure
open when printing PLA, but had never really had any issues so hadn't bothered.
I tried a print with the top propped open and everything worked smoothly again.
In hindsight all these clogging issues happened when it was particularly hot
and humid outside, which was probably a factor.

### Summary

Upgrading to the P1S has made the 3D printer feel more like a tool and less like
an additional ongoing project. The speed increase continues to amaze me, as
does the increase in quality.

I still think it was valuable to have the SV06 to learn how printers work --- you
definitely don't get as much of that from using something so consumer-focused
as the P1S --- and figure out if I was actually going to use it long-term or
whether it was just a fad for me. So it served its purpose, but I'm rather
glad to see the back of it in favour of something better in pretty much every
way.

[^1]: And if they did there's nothing anyone can do about it because copyright
doesn't exist in the same way in China.

[^2]: Although taking apart the P1S (which is a proprietary machine you might
not expect to be very user-servicable) was orders of magnitude easier than taking
apart the SV06 (which was an open source design). Everything is just much nicer:
the connectors are easier to attach and detach, the screws are all a consistent
size instead of needing a whole suite of allen keys, and so on.