---
date: 2024-12-06
title: Simple backups with Restic and Hetzner Cloud
permalink: /simple-backups-restic-hetzner/
resources:
  - src: restic.png
    name: The Restic logo — a gopher with two umbrellas.
    title: Restic's mascot, who's dual-wielding umbrellas to save you from a rainy day.
opengraph:
  title: "Simple backups with Restic and Hetzner Cloud · Chameth.com"
  type: article
  url: /simple-backups-restic-hetzner/
  image: /simple-backups-restic-hetzner/restic.png
---

{%figure "right" "The Restic logo — a gopher with two umbrellas." %}

I have a confession: for the past few years I've not been backing up any of my
computers. Everyone knows that you _should_ do backups, but actually getting
around to doing it is another story.

Don't get me wrong: most of my important things are "backed up" by virtue of
being committed to remote git repositories, or attached to e-mails, or
re-obtainable from the original source, and so on. I don't think any machine
failing completely would be a disaster for me, but it would certainly be a pain.

This week I finally got around to doing something, and it ended up being a lot
more straight forward than my previous forays into backup-land.

## Restic

After soliciting a few opinions, the choice of backup software came down to
either [Borg](https://www.borgbackup.org/) or [Restic](https://restic.net/).
I'm pretty sure either would have done what I want, but I leaned towards Restic
for a few reasons: it has a more informative website, it's written in Go
rather than Python[^1], and Borg seems to be transitioning between major
releases at the moment[^2].

<!--more-->

The way Restic works is pretty simple: you initialise a 'repository', and can
then call `restic backup /some/path` and it'll get backed up to the repository.
Restic handles keeping different backups separate, and only sending data that's
changed, and deduplicating, and so on. You basically point it at a thing you
don't want to lose, and it sorts it out for you. Perfect.

There are equally straight-forward commands for removing old snapshots
(`restic forget`) and verifying backups (`restic check`). One of the nice things
about modern backup solutions is they support a whole range of backends. I was
originally going to spin up a small VPS to host my backups, but noticed that
Restic supported S3-compatible stores…

### Hetzner Cloud Object Storage

I host my servers with Hetzner, and was going to use them to spin up a VPS as
well. Despite the "Cloud" branding on a bunch of products, they offer reasonable
prices and good service. A couple of months ago, they started offering
[S3-compatible object storage](https://docs.hetzner.com/storage/object-storage/overview).

The pricing isn't totally straight forward, but for continuous use the "free
quota" amounts to 1TB of storage and 1TB of egress a month. Ingress is free,
as is traffic within the `eu-central` region (where all my servers are). That
quota is only awarded when you pay the "base price", though, which is €4.99 a
month. So basically it's €5 a month for 1TB of storage and enough egress to
fully restore every single byte. That's better value than any VPS I can find,
much cheaper tham Amazon S3, and about the same as
[Backblaze B2](https://www.backblaze.com/cloud-storage).

### Getting them to work together

Now you'd think making the tool that supports S3-compatible object storage
work with your S3-compatible object storage would be easy, right? Well not
quite. The library Restic uses to deal with S3 backends has some
strange logic for figuring out the bucket name given a URL. It doesn't quite
seem to work right, though…

Hetzner buckets have URLs like `s3://bucketname.hel1.your-objectstorage.com`[^3],
but just passing that to Restic gives an error that the bucket is not specified.
The docs mention there's an advanced option to make it use the virtual host for
the bucket name: `-o s3.bucket-lookup=dns`. But that… also doesn't work.
What I ended up doing was specifying the URL as
`s3://hel1.your-objectstorage.com/bucketname`, and also passing in the `dns`
option. The library then seems to muddle its way back to a real, working URL.
I'm not sure why it works like this: maybe it's just a weird aspect of S3 that
I'm oblivious to?

The next fun part is that you can configure Restic entirely by using environment
variables, except for that `-o s3.bucket-lookup=dns` argument. That has to go
on the command line. I ended up making a little wrapper script to invoke Restic
correctly:

```shell
#!/bin/sh

export RESTIC_PASSWORD=repo-password
export RESTIC_REPOSITORY=s3:hel1.your-objectstorage.com/bucket-name
export AWS_ACCESS_KEY_ID=hetzner-access-key-id
export AWS_SECRET_ACCESS_KEY=hetzner-access-key

exec restic -o s3.bucket-lookup=dns "$@"
```

Then in the backup script I just alias `restic` to use the script:

```shell
#!/bin/sh

set -eu

alias restic='~/.bin/restic'

dirs=(
	"/some/path/to/backup/"
	"/some/other/path/"
)
for i in "${dirs[@]}" 
do
	echo $i
	(cd $i && restic --verbose backup .)
done

restic forget --keep-daily 7 --keep-weekly 10 --keep-monthly 24 --keep-yearly 10
```

The only other notable thing here is that the script changes into the directory
to be backed up. If you give Restic an absolute path, it will create a new
snapshot if the metadata of any folder in the path changes, which is not what
I want.

If it wasn't for the S3 URL issues, the whole thing would've probably taken
me about half an hour. That's including setting up the object storage,
installing Restic, and so on. It's painfully easy. Why didn't I do this three
years ago?!

[^1]: I'm not trying to be a language snob, but given the choice between two
otherwise equal projects one in Go and one in Python, I'll take the Go one.
I know I'm not going to have weird library issues down the line, and I'm much
more comfortable rummaging around the source.

[^2]: With a "Don't use this in production!" notice on the shiny new version.

[^3]: Aside: I really hate URLs like that. What is with the trend for completely
generic domains divorced from the service they're a part of?