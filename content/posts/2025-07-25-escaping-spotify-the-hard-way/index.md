---
date: 2025-07-25
title: "Escaping Spotify the hard way"
permalink: /escaping-spotify-the-hard-way/
tags: [music, personal, opinion]
format: long
resources:
  - src: wall-e.jpg
    name: "Still from WALL-E, showing two overweight people on floating beds, mindlessly consuming media"
    title: "I realised my media consumption was too close to this for comfort."
  - src: tauon.png
    name: "Screenshot of Tauon music player. There are playlist tabs on top, and album art for 1989 (Taylor's Version) on the right. Most of the screen is a track listing, and all the tracks are by Taylor Swift."
    title: "Tauon, showing off a wide range of different music in my library."
opengraph:
  image: /escaping-spotify-the-hard-way/tauon.png
---

{% figure "right" "Still from WALL-E, showing two overweight people on floating beds, mindlessly consuming media" %}

For the longest time I used Spotify for all my music needs. And I listen to
a lot of music: sometimes actively, but mostly passively as background noise.
I cancelled my premium subscription last December, and stopped using the
service entirely. Why? There's a bunch of reasons.

Let's talk about the money first. Spotify launched at £9.99/month, and stayed
that way for over a decade. Then in 2023 it went up to £10.99/month. That's
probably fair: the economy was in the toilet, and they haven't changed their
price in so long. Then in 2024 they upped it again to £11.99/month. Hmm.
The service they were providing me didn't improve in that time. I didn't want
audiobooks in my music player, I didn't want an AI DJ that spouted inane
comments at me. Paying them more money to do things I didn't want seemed silly,
and the money that actually goes to the artists is both so small and gets split
in such a convoluted way that it's not worth even thinking about.

My bigger concern was how uninvolved I became in choosing what to listen
to. I leaned hard on their algorithmic playlists like "Discover Weekly" instead
of manually curating playlists. But they ended up developing this weird 
feedback loop where it kept playing a country band[^1] that I didn't
particularly like, but also didn't dislike enough to skip. That acted as
positive reinforcement, and they kept coming up everywhere. There were
definitely ways to solve that by changing how I used Spotify, but the
realisation that I was basically just consuming whatever was fed to me made
me want to try something completely different. So I decided to go back to
buying music instead. Which is where things got… complicated.

<!--more-->

### Getting some MP3s will be easy, right?

I figured with the money I was no longer giving to Spotify, I could just buy an
album or two a month and slowly build up a nice music collection. I already
had some music kicking around from before the streaming era, so there would at
least be some variety in my listening while I got started. Where to buy music
from, though?

Amazon used to do DRM-free music, but as far as I can tell it all got folded up
into their own streaming service. Some smaller bands are on
[Bandcamp](https://bandcamp.com/), but you won't be able to grab Taylor
Swift's latest album from there. [Qobuz](https://www.qobuz.com/gb-en/shop)
have a good catalogue, but some of the stuff is _pricey_[^2]. It'd be cheaper
to buy CDs and rip them like it was 1998 again.

I finally stumbled across an unlikely answer: the iTunes store[^3]. This is
still somehow separate from the behemoth of Apple Music, and if you fire up
the iTunes software you can indeed buy reasonably priced, DRM free music.
For comparison, 1989 (Taylor's Version)[^4] is £12.89 (plus delivery) on CD
from Taylor's official website, £16.79 in CD quality from Qobuz, and
£12.99 from iTunes.

Now, I use Linux, so using iTunes isn't exactly easy. I ended up running a
virtual machine with Windows on it. Every time I wanted to buy an album, I'd
boot the VM, open iTunes, search for the album, press the buy button, enter
my password, wait for it to download, copy all the files to a shared drive,
and then back on Linux copy those files into my music directory. Smooth,
right? But it worked. Earlier in the year I switched computers, though, and
didn't want to bother with the VM rigamarole. Spoiler alert: I didn't make
things any simpler.

### Do you guys not have phones?

I tried a lot of things to avoid setting up a Windows VM again. I'm pretty
sure I went through most of the 7 stages of grief in the process. I tried to
get Apple Music to work, because that was obviously where all their development
was focused, but you can't actually buy music there (at least not on any version
I could run on Linux). I tried running iTunes under Wine, but it just wouldn't
work. I looked into third-party clients, but unsurprisingly they didn't go
anywhere near payment flows.

I eventually realised that I have an Apple device in my pocket. After a bit of
searching, it turns out there is actually still an iTunes app for iOS[^5]. I
bought an album[^6] to see if I could exfiltrate it. I connected my phone to
the computer, paired it with `idevicepair` and then mounted the storage using
`ifuse`[^7]. Looking around, I could see a bunch of m4a files, and I could play
them. Success!

Except… the file names were random hex strings, and they had no metadata.
Yuck. I grabbed a copy of [MusicBrainz Picard](https://picard.musicbrainz.org/)
and had it sort out the mess. I stuck with this process for a while, but
matching albums in Picard was sometimes a pain. MusicBrainz does a great
job of figuring out what the tracks are, but it struggles to tie those back
to a single edition of an album. I wasn't aware how many subtly different
versions of albums there were until I had to manually find the one iTunes had
given me. Not fun. Eventually I got so fed up with this that I did something
about it.

### Plists all the way down

Armed with an instance of Claude Code, I set out to see if I could write some Go
code to import the files properly. There were `plists` in the directory with
the m4a files, and I know enough to know that they're where all the juicy
metadata lives on iOS. Maybe we can parse them and do something useful?

There's a nice [plist parser](https://github.com/DHowett/go-plist) written
in Go already, so I got Claude to write a program to dump the contents of
one of the many `plist` files in the directory. Inside was all the metadata
for one of the tracks. Yay! Just one small problem: there's no way to link
the `plist` file to the audio file. There's one `plist` file that has a real
name, though: `StorePurchasesInfo.plist`. I had Claude translate it to JSON
and found that it contained a `data` field with a bunch of binary data. Huh.

I thought this was probably a dead end, but stuck the base64'd data into
[CyberChef](https://gchq.github.io/CyberChef/) and fiddled around a bit
to get a hexdump of the binary. Six characters jumped out at me: `bplist`.
I went back to Claude and suggested this might be a plist-in-a-plist scenario.
Claude then got very confused because it couldn't understand where the base64
was coming from[^8]. After some LLM-induced suffering, I eventually got the
contents of the inner-plist, and for each track it had a list of the audio
file, the metadata, and the album art. Yay!

With the data figured out, it was fairly simple to get Claude to write code to
parse everything and link them together, and after that move the files into
their correct place in my library. I also had it take the metadata from the
track's `plist` file and write it into tags on the m4a so my media player
didn't have to understand Apple nonsense. I looked around for a Go library to
do this, but couldn't find one supporting m4a files, so ended up just shelling
out to `exiftool`.

As a bonus, I made the tool automatically make a temporary directory and use
`ifuse` to mount the phone, then unmount it when it's finished. So now all I
have to do is plug my phone in, and run `import-music` on the command line[^9].
It's a very nice quality-of-life upgrade that I probably wouldn't have bothered
with if I didn't have Claude Code to speed the process up[^10].

### Sometimes I actually play music too

With all this talk about buying and sorting music, I almost forgot about the
most important bit: listening! I use [Tauon](https://tauonmusicbox.rocks/),
and it looks like this:

{% img "Screenshot of Tauon music player. There are playlist tabs on top, and album art for 1989 (Taylor's Version) on the right. Most of the screen is a track listing, and all the tracks are by Taylor Swift." %}

I tried a bunch of different music players, but Tauon hits the sweet spot of
being simple enough to use but capable enough in the right places. I
particularly like the playlist generators. My "Music" playlist just auto-imports
all music in my "Music" folder, then I have a "Classical" playlist with a
generator of `s"Music" g"classical" auto`. That makes it use the "Music"
playlist as a source, filter on tracks that have "classical" in their genre,
and automatically update when the base playlist changes[^11]. The "Disliked"
playlist does the same but filters on rating: `s"Music" a rat=0.5 auto`. That
takes all tracks (`a`) from the Music playlist, then filters it down to only
those that have a rating of 0.5 out of 5[^12]. Finally, the "Normal" playlist imports
everything from "Music" that's not in one of the others. It sounds a bit clunky,
but it works well in practice and I like the way it works.

Tauon in general just seems to work the way I'd expect it to, which is
refreshing. You can just start typing anywhere to search for a track, hit
enter, and it starts playing. It has keyboard shortcuts for other operations,
and standard features like reporting to last.fm and other sites.

Tauon also supports transcoding and copying tracks to a music player or a phone.
I use this to sync my main playlist to my phone. I'll give you a minute to just
bask in the glory of the rube goldberg machine I've created to move music from
my phone, to my computer, and ultimately back to my phone. This does actually
serve a purpose, though: it lets me control which tracks are on my phone. I
don't have to deal with excluding genres or tracks I dislike on the phone
in addition to the computer. In the future I might opt to just sync a subset
of my tracks to keep the storage size down. I'm telling you it makes sense!
Stop looking at me like that!

### Paying Attention Pays Off

I mentioned one of my aims for getting away from Spotify was to feel less
like I was just consuming whatever's in front of me. The corollary to that
is how do you discover new music without an algorithm to feed it to you?
I've ended up subscribing to a few blogs, paying more attention to music-related
news, and being more aware of songs on the radio and from other sources.

When I find a song I like these days, I tend to listen to a few other tracks
on the album (via YouTube, normally, as it's the least friction) and decide
if I want to buy it. When I do, I sit down and listen to the whole thing in
order. It's a radically different experience to just hearing a song now and
then. There are themes and connections between tracks that you just miss
if you listen to songs at random. Of course, you could do that on Spotify, but
I don't think it really lent itself to it.

My favourite example of this is my rediscovery of Linkin Park[^13]. I'm of the
age where they were around as I was growing up. I knew a couple of their big
hits but younger-me wasn't that into rock[^14], so I didn't pay them much
attention. Then I caught Up From the Bottom on the radio, and from there
ended up buying and listening to all of From Zero. It's an amazing album.
The Emptiness Machine has become my favourite song: not only is it a great tune,
performed well, it's also a fantastic introduction to a new era for the band
and to Emily Armstrong as the new vocalist. The rest of the album then builds
on it to show off what they can do.

I've since gone back and reconnected with a lot of Linkin Park's older albums,
too. And from there discovered a lot of other rock bands I really like.
Could this have happened with Spotify? Maybe, but it seems unlikely. It had me
in a country-band-I-can't-remember filter bubble, it seems pretty unlikely it'd
drop some Linkin Park on me out of nowhere. And the passivity that came along
with it would have meant that I probably wouldn't have paid enough attention
when I heard it from elsewhere.

Would I recommend that everyone runs out and cancels their Spotify
subscriptions? Eh, probably not. If it works for you, that's fine. Discovering
recursive plists is fun for me, but I appreciate it's not everyone's cup of tea.
What I would recommend is being a bit more deliberate in your choice of music.
Don't just tune in to a random playlist, or let the 'DJ' decide what to play.
Find something, and actually _listen_ to it. Then recommend it to me. I want
more music!

[^1]: I can't even remember the name of the band, that's how much I was into
them.
[^2]: And they're annoyingly deceptive about it. All the prices are "From £x",
where x is the discounted price you'd get if you paid them £15/month.
[^3]: I'd link to it, but… there isn't a website. That's not how things were
done in the 1990s, silly.
[^4]: I don't know why I'm so hung up on Taylor Swift in this post.
[^5]: Again, unrelated to the Apple Music app.
[^6]: It _probably_ wasn't a Taylor Swift album. Maybe. 
[^7]: The Arch wiki has [a useful page](https://wiki.archlinux.org/title/IOS)
on this (of course).
[^8]: We were translating the plists to JSON, and if you serialise a byte array
to JSON in Go it ends up as a base64-encoded string. Claude was then _adamant_
it needed to base64 decode the binary plist. It did not.
[^9]: I just had the sudden realisation that I can probably automate running
the `import-music` tool whenever I plug my phone in using systemd…
[^10]: I feel like I say that a lot lately.
[^11]: I like classical music, but it's a mood, and I'm often not in that
mood.
[^12]: I basically use this to filter out those introduction/bridge tracks
on albums that are just talking and not singing.
[^13]: Surprise! My library isn't _actually_ just Taylor Swift songs.
[^14]: Younger-me was a dumbass, in some ways.