---
date: 2016-08-11
strapline: With bonus completely over-the-top security
thumbnail: /res/images/yubikey/keys.thumb.png
title: Creating an offline GnuPG master key with Yubikey-stored subkeys
url: /2016/08/11/offline-gnupg-master-yubikey-subkeys/
image: /res/images/yubikey/keys.png
description: How to use an aircapped computer, a large dose of paranoia, an ironkey, and some yubikeys to create a new GPG key and subkeys.
---

<div class="image left">
 <img src="/res/images/yubikey/keys.png" alt="Two yubikeys">
</div>

I recently noticed that I'd accidentally lost my previous GPG private key &mdash; whoops. It was on
a drive that I'd since formatted and used for a fair amount of time, so there's no hope of
getting it back (but, on the plus side, there's also no risk of anyone else getting their hands on
it). I could have created a new one in a few seconds and been done with it, but I decided to treat
it as an exercise in doing things properly.

### Background: GPG? Yubikey?

GPG or GnuPG is short for [Gnu Privacy Guard](https://www.gnupg.org/), which is a suite of
applications that provide cryptographic privacy and authentication functionality. At a basic level,
it works in a similar way to HTTPS certificates: each user has a public key which is shared widely,
and a private key that is unique to them. You can use someone else's public key to encrypt messages
so only they can see them, and use your own private key to sign content so that others can verify
it came from you.

A [Yubikey](https://www.yubico.com/faq/yubikey/) is a small hardware device that offers two-factor
authentication. Most Yubikey models also act as smartcards and allow you to store OpenPGP
credentials on them.

<!--more-->

### Introducing subkeys

GnuPG supports subkeys, which provide fairly significant security advantages. Instead of just having
a single public and private key, you have a master pair and then any number of subkey pairs. The
subkeys are automatically associated with the master key, but they can be revoked independently.

Having a master key fall into the wrong hands is a problem &mdash; you have to revoke the whole
thing (assuming you have access to a revocation certificate) and start again, convincing everyone
else that your new key is the "real" you. With subkeys, you can issue a revocation signed with your
master key and then sign some new subkeys. There's no loss of trust, and as long as people refresh
your key from a keyserver, everything carries on as normal.

The other advantage to using subkeys is that you can keep the master key elsewhere. It doesn't
need to be routinely accessible, and using it doesn't require access to the Internet. The master
key is kept offline, significantly reducing the risk of anything bad happening to it.

### Setting up a secure environment

My main desktop runs Windows, and most of my other devices are work ones which come with automatic
backups and network mounts that I don't fully grok. Neither of those is a particularly good option
if I want to do something security sensitive.

[Tails](https://tails.boum.org/) is the defacto choice for a secure, live linux system, so I went
through their installation process and eventually ended up with a USB drive that can boot into
Tails. The installation process from Windows is slightly convoluted, as it involves creating a
bootable Tails image, then booting to that and using the Tails installer to create the real image
on a different drive. If you're starting on a Linux box, you can just use the Tails installer
directly instead of doing the two-drive shuffle.

Just to be completely paranoid, I disconnected my PC from the network before booting Tails. This
is known as [air-gapping](https://en.wikipedia.org/wiki/Air_gap_%28networking%29), and is done to
eliminate the possibility of a remote attacker doing something to your system. Ideally the machine
would never have been connected to the network, but I didn't happen to have an unused machine
laying around.

The final thing I needed was a secure place to store my master key. I opted for an
[IronKey](http://www.ironkey.com/en-US/) &mdash; a hardware-encrypted USB drive that self-destructs
if there are too many unsuccessful attempts to access it. It works out-of-the-box on both Windows
and Linux, presenting a small unencrypted drive with software to run to interact with the secure
partition.

### Creating the keys

Now I had a nice over-the-top setup it was time to actually the keys. There is [an excellent
guide by Simon Josefsson](https://blog.josefsson.org/2014/06/23/offline-gnupg-master-key-and-subkeys-on-yubikey-neo-smartcard/)
that walks through the entire process of creating the master key, creating three subkeys, and
then transferring them to a Yubikey.

The only point where I had to deviate from Simon's guide was setting the machine up to work with
the Yubikeys. I was setting up two keys (a nano and a neo), and one just worked out of the box
with the version of `libykpers-1-1` that was in Tails' apt repository. The other needed a slightly
newer version but that was also available in apt and can be selected by specifying the version
manually as pointed out in [this GitHub issue](https://github.com/freedomofpress/securedrop/issues/1035#issuecomment-140172267).
The version numbers have since changed, but `apt-policy` makes it easy to figure out what's needed.
As I was using an air-gapped system this process was a bit more complicated than it sounds,
involving several USB drive transfers.

After finishing the guide, I had my master key and a pre-generated revocation certificate stored
securely on my IronKey, and the three subkeys stored on each Yubikey. Time to go back to Windows.

### GPG, Windows and SSH

Now with the IronKey disconnected and the master key out of harms way, it's time to go back to
Windows. I downloaded the [GnuPG Modern](https://www.gnupg.org/download/) distribution and
followed the instructions at the end of Simon's guide to import my public key and make GPG aware
of the subkeys on the Yubikey. After that [Enigmail](https://www.enigmail.net/index.php/en/)
was able to sign and encrypt e-mail in Thunderbird.

<figure class="image left">
  <img src="/res/images/yubikey/wisdom_of_the_ancients.png" alt="XKCD: Wisdom of the ancients">
  <figcaption><a href="https://xkcd.com/979/">XKCD #979: Wisdom of the ancients</a></figcaption>
</figure>

Next up, I enabled PuTTy support and started the GPG agent, as documented over on
[Yubico's site](https://developers.yubico.com/PGP/SSH_authentication/Windows.html). This allows
you to use the authentication GPG key to authenticate SSH sessions from PuTTy. To find the
SSH key you need to add to `.authorized_keys`, simply run `gpg --export-ssh-key`. At first I
could SSH into a host but not use agent forwarding. After lots of unsuccessful Googling, I realised
that GPG couldn't access the key anymore locally. Another quick search and I found a
[forum thread](http://forum.yubico.com/viewtopic.php?f=35&t=2231) where someone had the same
issue and found it was a problem with exclusive access to the card. They even passed on their
wisdom and updated the thread with a solution, which got everything working for me.

### VMWare

I have a Ubuntu image running inside VMWare on my desktop that I use for most development
activities. I'm unlikely to want to sign e-mail, but I probably want to sign commits (especially
now that [GitHub exposes verified signatures](https://github.com/blog/2144-gpg-signature-verification)).

To do that I need to pass the Yubikey through to the virtual machine. In its default configuration,
VMWare recognises the Yubikey device but doesn't pass it through correctly. You need to configure it
to [allow HIDs](http://www.timothysalmon.com/2014/12/vmware-workstation-connect-yubikey-to.html),
after which `gpg --card-status` starts working from the VM.

Unfortunately passing through the device makes it exclusively available to the VM so the host OS
can no longer use it. As I have two Yubikeys, I just configured VMWare to ignore the nano key that's
always plugged in, and pass through the neo when I plug that in. While swapping the configuration
around, I found out that GPG remembers the ID of the smartcard that store credentials - if you
swap the two keys with identical subkeys it demands the other one is reinserted. You can get rid of
the references to the previous card using `gpg --delete-secret-keys`.

### In conclusion...

... I have a <a href="/16402FE2.txt">new PGP key</a> you can use to verify things I sign or encrypt
messages to me.
