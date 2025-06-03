---
date: 2022-02-18
title: Reproducible Builds and Docker Images
description: How hard can it be...?
tags: [docker, security]
permalink: /reproducible-builds-docker-images/

resources:
- src: dependency.png
  name: Comic showing all modern digital infrastructure is built upon one project by a random person in Nebraska
  title: "XKCD 2347: Dependency"
  params:
    default: true
---

{% figure "left" "Comic showing all modern digital infrastructure is built upon one project by a random person in Nebraska" %}

[Reproducible builds](https://reproducible-builds.org/) are builds which you are able to reproduce byte-for-byte,
given the same source input. Your initial reaction to that statement might be "Aren't nearly all builds
'reproducible builds', then? If I give my compiler a source file it will always give me the same binary, won't it?"
It _sounds_ simple, like it's something that should just be fundamentally true unless we go out of our way to break it,
but in reality it's actually quite a challenge. A group of Debian developers have been working on reproducible packages
for the best part of a decade and while they've made fantastic progress,
[Debian still isn't reproducible](https://isdebianreproducibleyet.com/). Before we talk about why it's a hard problem,
let's take a minute to ponder why it's worth that much effort.

### On supply chain attacks

Suppose you want to run some open-source software. One of the many benefits of open-source software is that anyone
can look at the source and, in theory, spot bugs or malicious code. Some projects even have sponsored audits or
penetration tests to affirm that the software is safe. But how do you actually deploy that software? You're probably
not building from source - more likely you're using a package manager to install a pre-built version, or downloading
a binary archive, or running a docker image. How do you know whoever prepared those binary artifacts did so from
an un-doctored copy of the source? How do you know a
[middle-man hasn't decided to add malware to the binaries to make money](https://en.wikipedia.org/wiki/SourceForge#Controversies)?

<!--more-->

Even worse: if the software you're trying to use includes any dependencies, you have the same issue of trust
with them. Maybe _your_ supplier isn't compromising the software, but that doesn't mean _their_ supplier isn't. The
beauty-cum-horror of a supply chain attack is that it can target the weakest link anywhere along the supply chain.
Even if there aren't any binary files involved, dependencies can still be attacked: what if `npmjs.com` or
`proxy.golang.org` or `github.com` return a different version of a dependency-of-a-dependency when the request
comes from your IP address? It doesn't even need to be a modified dependency, it could be a perfectly un-tampered,
properly signed copy of the source, just from an older version with a known vulnerability.

Enter stage left: reproducible builds, here to save the day! If the build process is reproducible then you - or anyone
else on the internet - can perform the same build on the same source and validate the output has the same checksum or
hash. If Debian publish a binary package and an independent re-builder comes up with the exact same build artifact,
there's a reasonably good chance that the build is good. An attacker would have to compromise both the build machine
and the re-build machine to do anything nefarious. The more re-builders there are, the less feasible a supply chain
attack is.

### So why isn't software just reproducible?

#### Compilers

As a bit of an experiment, I asked some friends to run the following for me and report the answer:

```shell
echo -e "#include <stdio.h>\nint main() { printf(\"Hello\"); return 0; }" | \
  gcc -x c -o hello.out - && \
  sha256sum hello.out
```
This compiles a super-simple hello world program and then prints the SHA-256 hash of the resulting binary. Here are
the results:

| Hash                                                             | System | GCC    |
|------------------------------------------------------------------|--------|--------|
| 1f62feab5a06861dc575201d807781926d1ae49fb113da018fde8b670a1346f7 | Arch   | 11.2.0 |
| b8e6f2c7082be69f65ffa5e7a3d749eb47866a1b2e1ec19efb63cc59a8b160cd | Debian | 8.3.0  |
| cbad2e47a22c234b5e7fa55e029a8db4d64ac7a962e2176bd2e1373d78954088 | Debian | 8.3.0  |
| e0f6bbc13b29fea8cfa2a975ba4661e781323298aec166c8311d342e6f93c4a6 | Alpine | 10.3.1 |
| e379156895e06c7a0bf18ac4d648860edcb2655576b0ab9fab172bd6c8b92075 | Debian | 10.2.1 |
| 7ffdaee4eb64e016b89dc5e54d2c8eebab3cebafe2c7aa97de627b5972ecea46 | Debian | 11.2.0 |
| 8ae52cc166743b6ae1eb3e14179ef33de5061a04237f8f97088c896c41a2f698 | Arch   | 11.1.0 |
| 8ae52cc166743b6ae1eb3e14179ef33de5061a04237f8f97088c896c41a2f698 | Arch   | 11.1.0 |

As you can see, there are barely any duplicates. Even the same version of GCC on the same OS sometimes produces
different results. And this is the most basic program I could write! Differences arise from the compiler version,
the build flags, the libraries installed, and a whole host of other factors. If you compile a Go application instead of
a C one, then by default the compiler will include debug information in the binary. This includes the full path to the
source file on disk, so building a project in `/home/chris/` will produce a different binary to building the same
source in `/tmp`. Future versions of Go are also going to stamp in other meta-data such as VCS info, so building inside
and outside a Git repository will produce different binaries.

#### Archives

Compilers are only half the problem. Build processes are usually multistep, involving compiling, moving, compressing,
and so on. Consider creating an archive of a file:

```shell
repeat 4 touch hello && tar zcf hello.tgz hello && sha256sum hello.tgz && sleep 0.5
f3d5c56f6b8089de95d62d060e6ffcbbad26875807ae7bc253f07cd097ea61be  hello.tgz
ab67f2e865b5afa87d9b2434d92b0c271b3cf730fa85988f84852551749ba6ed  hello.tgz
ab67f2e865b5afa87d9b2434d92b0c271b3cf730fa85988f84852551749ba6ed  hello.tgz
738678c9650b10fd83636997dd1aba4016bbf0ec5ebf3dfd4ef75d770b56e23b  hello.tgz
```

Any file added to a tar takes with it a timestamp, so the build is only reproducible if it happens at the exact same
time! We can make this reproducible by forcing `tar` (and the same goes for `zip` and most other archive formats) to
set a certain timestamp on the files:

```shell
repeat 4 touch hello && tar --mtime 2022-02-18T01:00 -zcf hello.tgz hello && sha256sum hello.tgz && sleep 0.5 
081060a900beff2a6aad9957a8cbb8792f8db7904f86b318dbf26b682a2d3f0a  hello.tgz
081060a900beff2a6aad9957a8cbb8792f8db7904f86b318dbf26b682a2d3f0a  hello.tgz
081060a900beff2a6aad9957a8cbb8792f8db7904f86b318dbf26b682a2d3f0a  hello.tgz
081060a900beff2a6aad9957a8cbb8792f8db7904f86b318dbf26b682a2d3f0a  hello.tgz
```

In a real build there are basically two approaches here: you can set it to a pre-defined value (like the unix epoch),
or you can set it to match the modification time of the source files. The former is easiest, but the latter is more
cosmetically and semantically appealing.

#### Iteration order

So we've pinned our build environment, we're manipulating timestamps when adding files to archives, now what? Imagine
part of the build process involves looping through all the files in a directory and doing _something_. What order do
these files get iterated in? Well, that very much depends on the filesystem and perhaps when the files themselves
were created. To ensure this is reproducible we need to explicitly sort any such operation so that it's always
consistent. This iteration could be happening in a tool that's called by another tool that's called by a build script,
so the fix isn't necessarily straight-forward.

{% sidenote "a bug war story" %}
I've personally been victim to this kind of non-determinism. I was working on an Android app, and committed a new
test that worked fine on my machine, and worked fine on the CI server. But it failed consistently for a colleague.

We both did fresh checkouts of the source, and ran the tests. Mine passed, his failed. He sent me an archive of
his checkout in case there was something weird going on there, and the tests passed on my machine. We compared
hashes of our checkouts, and they were the same. It was obviously environmental somehow, but everything else worked
fine, and the build system went to great pains to ensure things were the same.

After a _lot_ of debugging, I worked out that his test was running with a different version of a library to me,
despite the libraries being defined in the build files and the build files being identical. After _even more_
debugging it turned out there were two versions of the library on the classpath, and the ordering of them was
different between my machine and his.

The actual issue turned out to be that the build tool generated the classpath by iterating over the library
files, and that iteration was done in order of file creation time. The two libraries were added at different points
in the project history, so the creation time in your local cache depended on which versions of the app you'd built
in the past. With no cache everything worked as expected but there was a slim range of commits where only one
library was in use, and if you had run the tests during that period your cache was effectively poisoned.

We fixed the issue by excluding the older version of the library (which was being pulled in as a transient dependency),
and filed a bug against the build tool to make the classpath properly deterministic. I think that stands as the most
difficult to diagnose bug I've ever dealt with.
{% endsidenote %}

Interestingly, if you iterate over a map in Go, the iteration is _deliberately_ non-deterministic. That's an attempt
to defeat [Hyrum's Law](https://www.hyrumslaw.com/) and prevent developers from relying on whatever the current
behaviour happens to be. This actually makes it easier to make things reproducible as the problem is loud and
in-your-face, rather than subtle and hard to spot.

#### Other sources

There's an awful lot of other places that non-determinism can come from. If the app pulls in dependencies, their
versions have to be pinned, otherwise your build changes depending on the latest release of that dependency. If
the build process pulls any information from a website, it's liable to change. Hopefully the website is under your
control so that you can version the resource and pin that version.  Obviously, anything to do with dates or the
current user will probably cause problems. Timezones and locales can cause subtle differences.

### What about Docker?

Docker comes with some good and some bad points for reproducibility. The biggest advantage is that it inherently
completely describes the build environment; it should work exactly the same from one system to another, even across
different OS families. The biggest draw back is it sprays timestamps around like no-one's business. Each layer in
a container image is a `.tar.gz` file, meaning each file within it is timestamped as discussed above. Making an image
involves a lot of copying of files around, so these timestamps invariably end up causing reproducibility issues.

Even worse than timestamps in the filesystem, the image format also contains some meta-data that includes the
timestamp at which each layer was built. That means even if you go out of your way to set the timestamp of every
single file in your image, the image itself will be different every time you rebuild it. There is no way to deal
with this in Docker, which is a very sad state of affairs. Fortunately, [Buildah](https://buildah.io/) provides
a `--timestamp` flag for _its_ build commands; this not only sets the layer timestamp but also the creation
timestamp of any file within the layer.

The other major issue that affects Docker images is the pinning of packages pulled in by package managers. An awful
lot of images are based on Alpine or Debian derivatives, and use `apk` or `apt` to install dependencies. These need
to have a version specified as otherwise the package manager will just pull in the latest at the time of the build.
But this isn't quite enough: you also need to pin the version of any packages that they depend on, recursively.
This means flattening the entire package hierarchy and installing all the packages explicitly and with pinned
versions.

One more wrinkle in the package management space is that Alpine don't keep old packages in their main repositories.
If you have a Docker image with pinned alpine packages in, it will stop building if the package is updated. This
isn't necessarily fatal to making a reproducible build -- as long as it's reproducible for its useful lifetime,
I don't really see an issue.

Honestly, though, the biggest issue with making Docker images reproducible is getting people to care. Dockerfiles
are a relatively new way of packaging software, and there's no centralised organisation like you find with Linux
distributions. There are enough challenges that most casual packagers aren't going to bother, and no real
incentive for them to. That won't stop me trying, though!
