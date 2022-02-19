---
date: 2021-06-12
title: Reverse engineering an Arctis Pro Wireless Headset
description: In which we go from "is it on or not?" to "can it play gifs?"
area: reverse-engineering
slug: reverse-engineering-arctis-pro-wireless-headset

resources:
- src: headset.png
  name: Boxed SteelSeries Arctis Pro Wireless Headset
  title: The Arctis Pro Wireless Headset.
  params:
    default: true
- src: wireshark1.png
  name: Wireshark, showing a packet capture of a SET_REPORT request from the host to the headset
- src: wireshark2.png
  name: Wireshark, showing a packet capture of the headset's response to the earlier request
- src: settings2.png
  name: SteelSeries Good Game software, showing the advanced headset settings
- src: display.png
  name: Diagram of how pixels are addressed on the receiver's OLED display
- src: helloworld.png
  name: Headset receiver displaying a custom Hello World message
- src: nyan.webm
  name: Video of receiver playing Nyan cat gif
---

{{< figure "left" "Boxed SteelSeries Arctis Pro Wireless Headset" >}}

For the last year and a bit, I've been using a [SteelSeries Arctis Pro Wireless Headset](https://steelseries.com/gaming-headsets/arctis-pro-wireless)
for gaming and talking to friends. It's a fine headset, but because there's an always-on receiver there's no
way to detect if the headset is turned on or not from the desktop.

Whenever I start using the headset, I set my desktop's sound to go to the headset, and then when I stop using
the headset I set it to go back to speakers. It doesn't take more than a second, but some days I might put the
headset on a dozen times as I'm on calls, or if it's noisy outside, etc. That means it's probably
[worth at least a few hours of my time](https://xkcd.com/1205/) trying to automate it.

At first, I hoped I'd be able to tell from the state of the USB device whether there was a headset
connected but nothing at all changed when flipping it on and off. Then I went hunting for existing
open source tools that might work with it and found that while people have reverse engineered many
of the older Arctis headsets, no one has done the same for the Pro Wireless. I finished off with
a search to see if anyone had documented the wire protocol even if there was no nice open source
software to go with it; I came up short there, too. Looks like I'd have to do it myself.

<!--more-->

### Capturing data with WireShark

The headset exposes a Human Interface Device (HID), and the wire protocols for earlier versions
of the Arctis series looked to be very simple messages passed over the HID connection. It should
therefore be fairly easy to use [WireShark](https://www.wireshark.org/) to capture the data
being sent and received by the official SteelSeries application[^1]. 

After setting up WireShark and making it record only the headset's HID connection, it quickly
became apparent that the software was sending three different requests each second, over and
over again:

{{< img "Wireshark, showing a packet capture of a SET_REPORT request from the host to the headset" >}}

The packets we're interested in are the `SET_REPORT Request` frames sent from the host to the
device. Wireshark understands the HID protocol, so it nicely shows us the raw HID data; in the
screenshot this is `0x40AA` (ignoring the trailing zero bytes). The other requests sent immediately
after have `0x41AA` and `0x42AA` payloads — clearly the first byte is indicating which piece
of data is being requested.

The responses to these requests come back in an `URB_INTERRUPT in` frame:

{{< img "Wireshark, showing a packet capture of the headset's response to the earlier request" >}}

So the answer to our first request appears to be `0x04`. The second response is `0x0402` and the third is `0x04` again.
The first thing I tried doing was popping the spare battery out of the charger. The response to the `0x42AA` request
changed from `0x04` to `0x00` - that's certainly clear enough! After a bit of waiting around, the battery in my headset
dropped from full on the display to three bars, and at that point the `0x40AA` response dropped from `0x04` to `0x03`.
The display on the receiver shows battery state as four bars, and it appears the wire protocol directly corresponds to
that particular representation.

That left the `0x41AA` request as an unknown. I tried everything I could think of, but it stubbornly kept returning
`0x0402`. I enlisted a friend who has the same headset to run some hacky Go code and report his values, and he also
got a `0x0402` response. As I explained that I couldn't figure out what these values are, he reported back that turning
his headset off made the response change to `0x0202`. In all my testing, I'd forgotten to try turning the headset off!
That's the one thing I was actually trying to detect, as well. Thanks, Simon, for helping me get past that bit of
stupidity!

I still don't know what the second byte of the response is, or whether there are other values than `0x04` for on
and `0x02` for off, but I'm happy enough to label it as "device status" and move on.

### Exploring other features

The software allows you to tweak a bunch of different settings:

{{< img "SteelSeries Good Game software, showing the advanced headset settings" >}}

I went through each one and fiddled with all the values, recording the requests in WireShark as I did so. It turns
out the protocol is very simplistic - the wire protocol directly corresponds to the UI elements in the software (or
on the receiver, if you navigate through its menus). For example, the software allows you to set the "Headset auto
shutoff" value in increments of 10 minutes. You might usually expect this to be converted to seconds or something
similar before being passed to the device, but on the wire it's actually sent as `0x00` for off, `0x01` for 10 minutes,
up through to `0x0C` for the maximum of 120 minutes.

All the dropdown options seem to function this way — if you pick the 6th option then the request payload will be
an `0x06` byte. The sliders have fixed positions they snap to and function similarly: for the two brightness sliders
pictured they snap to 11 positions and on the wire these range from `0x00` for off to `0x0A` for the maximum.

At this point I was well past my goal of being able to detect whether the headset was on or off, and I was now just
trying to see if I could figure out enough of the protocol that I could reimplement the control software on Linux if
I ever wanted to. Just as I was about to close the app, something caught my eye: there are integrations with games
and other software that can display information on the receiver's OLED display!

There's an [official API](https://github.com/SteelSeries/gamesense-sdk) for this, but it involves sending JSON to
a webserver which runs as part of their app[^2]. That doesn't feel very nice, and definitely won't work for me on Linux,
so I started a new WireShark session and took some captures while I spoke to myself on Discord.

### Decoding pixels

The frames sent to the device whenever Discord showed a notification had a 1060 byte payload. The display on the
receiver is 140 pixels wide[^3] and each pixel can only seem to be on or off, so I figured each bit in the payload
corresponded to one bit of the output. I exported the data to binary in the hope that I could visually see what was
going on - in theory if I line wrap the data at 140 characters it should look vaguely like the final output.
Unfortunately, it did not. There was roughly the right number of high bits, but no amount of fiddling with them in
a text editor could give me a coherent picture.

Instead, I wrote some code to write different values to the device. Starting with a payload of all zeroes and gradually
increasing a bit at a time every second. After the first few seconds, I saw a line of pixels being drawn downwards on
the left of the screen; had I just got the axes the wrong way around? After the first eight pixels lit up, though, they
jumped over to the next column. The actual addressing scheme looks something like this:

{{< img "Diagram of how pixels are addressed on the receiver's OLED display" >}}

So the first 140 bytes given the pixels for the first 8 rows, the next 140 bytes fill in the 8 rows below that, and
so on and so forth. Armed with this information I wrote a simple program to read an image and output it to the
display:

{{< img "Headset receiver displaying a custom Hello World message" >}}

At this point I was thinking about trying to get Doom rendering on the screen, but I couldn't find anything nicely
hackable that would let me grab the output and pass it on to the receiver. Instead, I decided to try a GIF decoder
and before very long had a nice little animated display:

{{< video "Video of receiver playing Nyan cat gif" >}}

### Protocol reference

All the HID messages have a single byte that determines the message type, then an `0xAA` byte, then any payload
required by the command. These are the ones I've figured out:

Byte | Command | Payload
--- | --- | ---
`0x09` | Save changes | None.
`0x10` | Request firmware version(?) | None.
`0x27` | Volume limiter | `0x00` for off, `0x01` for on.
`0x2E` | Equalizer preset | ID of the equalizer preset to use.
`0x39` | Sidetone level | `0x00` for lowest to `0x09` for highest.
`0x3C` | Set headset timeout | Timeout, as a number of 10 minutes. `0x00` for off to `0x0C` for 120 mins.
`0x3E` | Mic mute LED brightness | `0x00` for lowest to `0x0A` for highest.
`0x40` | Request headset battery | None.
`0x41` | Request device status | None.
`0x42` | Request receiver battery | None.
`0x51` | Surround sound mode | `0x00` for off, `0x01` for on.
`0x62` | Auto-start Bluetooth | `0x00` for off, `0x01` for on.
`0x63` | Auto-mute game audio during calls | `0x00` for off, `0x01` for on.
`0x85` | OLED brightness | `0x00` for lowest to `0x0A` for highest.
`0x83` | Equalizer | (Not yet decoded)
`0x89` | Screensaver mode | `0x00` to dim, `0x01` for off, `0x02` for screensaver.
`0xD2` | Render image | Pixel array as described above.

Some of these I've not dug too much into because they didn't seem very interesting. If you have an Arctis Pro Wireless
and figure anything more out, let me know, and I'll update the list.

If you just want to check device state like I originally did, I've contributed support for this headset to the
excellent [HeadsetControl](https://github.com/Sapd/HeadsetControl) project. It should be in the next release.

[^1]: Unfortunately (as you'd expect) their application only runs on Windows, so this process
      involved an annoying amount of rebooting to Windows, fleeing back to Linux, and then
      realising I hadn't actually recorded enough to figure it out and repeating.
[^2]: You can also send lisp to the webserver, and it will execute it. No comment.
[^3]: I know this not because it's mentioned in the technical specs (it's not), but because I took a photo and counted
      them out one by one.
