---
date: 2019-04-01
title: Understanding Docker volume mounts
description: It's basically magic.
tags: [docker, research]
format: long
permalink: /understanding-docker-volume-mounts/

resources:
  - src: logo.png
    name: The Docker project logo
---

{% figure "left" "The Docker project logo" %}

One thing that always confuses me with Docker is how exactly mounting
volumes behaves. At a basic level it's fairly straight forward: you
declare a volume in a Dockerfile, and then either explicitly mount
something there or docker automatically creates an anonymous volume
for you. Done. But it turns out there's quite a few edge cases...

### Changing ownership of the folder

Perhaps the most common operation done on a Docker volume other than
simply mounting it is trying to change the ownership of the directory.
If your Docker process runs as a certain user you probably want the
directory to be writable by that user.

<!--more-->

At first we might try something like:

```docker
FROM alpine
RUN adduser -D -u 1113 test123
USER test123
VOLUME /testing
```

But changing the user doesn't seem to have any effect on the volume.
Why? Checking the docs for the 
[`USER` instruction](https://docs.docker.com/engine/reference/builder/#user)
shows that only affects certain future operations -- namely
`RUN`, `CMD`, and `ENTRYPOINT`. It doesn't affect the `VOLUME` instruction;
if it did, you'd probably just get a permission denied error unless the user
you switch to had privileges to create mount points.

OK, so instead we might try using the good old `chown` command:

```docker
FROM alpine
RUN adduser -D -u 1113 test123
VOLUME /testing
RUN chown test123 /testing
```

But again, the directory is just owned by root at runtime.
Back to the docs, this time for the
[`VOLUME` instruction](https://docs.docker.com/engine/reference/builder/#volume)
and towards the bottom is this little tidbit:

> Changing the volume from within the Dockerfile: If any build steps change
> the data within the volume after it has been declared, those changes will
> be discarded.

As soon as Docker hits the `VOLUME` instruction the directory becomes a mount
point, and anything we do to the temporary volume mounted there is discarded
during the build process. So we have to change the ownership *before* the
instruction, which may seem a little counter-intuitive:

```docker
FROM alpine
RUN adduser -D -u 1113 test123
RUN mkdir /testing && chown test123 /testing
VOLUME /testing
```

Now when the container runs, the /testing directory is owned by the test123
user. It's not quite over, yet, though. This works if we let Docker create
a volume automatically for us, or if we create a named volume and mount that;
if we try and mount a host directory, though, it falls flat:

```text
$ docker run --rm -it -v "$PWD/testing:/testing" testing ls -al /testing
total 8
drwxr-xr-x    2 1000     1000          4096 Apr  1 19:39 .
drwxr-xr-x    1 root     root          4096 Apr  1 20:44 ..
```

Docker handles mounting host directories differently to mounting volumes,
even though the syntax is basically the same. Host directories are bind
mounted directly into the container, so the permissions and ownership
are the same as the directory on your host. The only way to fix them are
to either change the permissions on the host, or have the container
change them at runtime (assuming it has sufficient privileges).

One final wrinkle in all this happens when you use the same volume
in multiple containers. Here we have two images built from the
Dockerfile above, one with userid 1113 and one with userid 1114:

```text
$ docker volume create testing
testing

$ docker run --rm -it -v testing:/testing testing1113 ls -nal /testing
total 8
drwxr-xr-x    2 1113     0             4096 Apr  1 19:49 .
drwxr-xr-x    1 0        0             4096 Apr  1 20:51 ..

$ docker run --rm -it -v testing:/testing testing1114 ls -nal /testing
total 8
drwxr-xr-x    2 1114     0             4096 Apr  1 20:47 .
drwxr-xr-x    1 0        0             4096 Apr  1 20:52 ..

$ docker run --rm -it -v testing:/testing testing1114 touch /testing/Hello

$ docker run --rm -it -v testing:/testing testing1113 ls -nal /testing
total 8
drwxr-xr-x    2 1114     0             4096 Apr  1 20:52 .
drwxr-xr-x    1 0        0             4096 Apr  1 20:53 ..
-rw-r--r--    1 0        0                0 Apr  1 20:52 Hello

$ docker run --rm -it -v testing:/testing testing1114 ls -nal /testing
total 8
drwxr-xr-x    2 1114     0             4096 Apr  1 20:52 .
drwxr-xr-x    1 0        0             4096 Apr  1 20:52 ..
-rw-r--r--    1 0        0                0 Apr  1 20:52 Hello
```

Can you see what's going on? When the volume is empty, the ownership
changes based on the mount point in the container. Once it has something
in it, the ownership is fixed.

So Docker behaves differently with regard to permissions:

 * when the folder is mounted from the host vs a volume
 * when the volume is empty vs having content

### Pre-populating mounts with files from the image

One of the more esoteric features of the way Docker handles volume
mounts is that in some cases files from the image are copied over
into the container. For example:

```text
$ docker volume create testing
testing

$ docker run --rm -it -v testing:/etc testing sleep 1

$ docker run --rm -it -v testing:/tmp testing ls -al /tmp
total 184
drwxr-xr-x   15 root     root          4096 Apr  1 20:58 .
drwxr-xr-x    1 root     root          4096 Apr  1 20:59 ..
-rw-r--r--    1 root     root             4 Jun  7  2018 TZ
-rw-r--r--    1 root     root             6 Dec 20 21:31 alpine-release
...
```

The first container we run mounts the newly created `testing` volume
at `/etc`. Docker copies all the existing files and folders into the
volume; when we then run the second container with the volume mounted
at `/tmp`, we can see all of the files that were in the first container's
`/etc`.

As with permissions, this behaviour is anything but consistent. Say we
switch from a volume to a host directory:

```text
$ mkdir testing
$ docker run --rm -it -v "$PWD/testing:/usr/bin" testing sleep 1
$ ls -al testing
total 8
drwxr-xr-x 2 root  root  4096 Apr  1 22:05 .
drwxr-xr-x 3 chris chris 4096 Apr  1 22:05 ..
```

Nothing is copied in, and inside the container the folder will be empty.
Based on our discoveries with permissions, it's reasonable to assume the
same will happen with a non-empty volume too:

```text
$ docker volume create testing
testing

$ docker run --rm -it -v testing:/testing testing touch /testing/Hello

$ docker run --rm -it -v testing:/usr/bin testing sleep 1

$ docker run --rm -it -v testing:/tmp testing ls -al /tmp
total 8
drwxr-xr-x    2 root     root          4096 Apr  1 21:09 .
drwxr-xr-x    1 root     root          4096 Apr  1 21:09 ..
-rw-r--r--    1 root     root             0 Apr  1 21:08 Hello
```

So at least that's consistent. If you're very observant, though, you
might notice I switched from `/etc/` to `/usr/bin` in the examples.
That's because within the container `/etc/` has some files bind-mounted
into it, such as `/etc/resolv.conf`, and these *do* always result in files
being created in the mounted volumes or folders:

```text
$ mkdir testing
$ docker run --rm -it -v "$PWD/testing:/etc" testing sleep 1
$ ls -al testing
total 8
drwxr-xr-x 2 chris chris 4096 Apr  1 22:12 .
drwxr-xr-x 3 chris chris 4096 Apr  1 22:12 ..
-rwxr-xr-x 1 root  root     0 Apr  1 22:12 hostname
-rwxr-xr-x 1 root  root     0 Apr  1 22:12 hosts
-rwxr-xr-x 1 root  root     0 Apr  1 22:12 resolv.conf
```

### Summary

 * Docker treats mounting host folders and mounting volumes differently.
   Don't just assume that you can swap one for another and get the exact
   same behaviour.
 * Empty volumes will inherit permissions and files from the image
   they are mounted in; non-empty volumes and host folders will not.
 * Relying on Docker copying files into volumes is a very bad idea,
   as if you change those files in a future version of your image
   they will not be copied unless the volume is deleted and
   recreated.

I can't find anywhere that these points are documented properly;
if you know of anywhere, please drop me a message!
