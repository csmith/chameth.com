---
date: 2016-05-21
strapline: It's containers all the way down...
thumbnail: /res/images/docker/logo.thumb.png
title: Automatic reverse proxying with Docker and nginx
url: /2016/05/21/docker-automatic-nginx-proxy/
aliases: ["/2016/05/21/docker-automatic-nginx-proxy.html"]
image: /res/images/docker/reverse-proxy.png
description: Automatically retrieve certificates from Let's Encrypt and configure an SSL-terminating reverse proxy based on running containers.
area: Docker
---

<figure class="right">
  <img src="/res/images/docker/logo.png" alt="Docker logo">
  <figcaption>The Docker project logo</figcaption>
</figure>

Over the past few weeks I've gradually been migrating services from running in LXC containers to
Docker containers. It takes a while to get into the right mindset for Docker - thinking of
containers as basically immutable - especially when you're coming from a background of running
things without containers, or in "full" VM-like containers. Once you've got your head around that,
though, it opens up a lot of opportunities: Docker doesn't just provide a container platform, it
turns software into discrete units with a defined interface.

With all of your software suddenly having a common interface, it becomes trivial to automate a lot
of things that would be tedious or complicated otherwise. You don't need to manage port forwards
because the containers just declare their ports, for example. You can also apply labels to the
application containers, and then query the labels through Docker's API.

<!--more-->

### Reverse proxying and SSL termination with Nginx and Let's Encrypt

A fairly significant chunk of the software I run has a web interface. I don't really want to
expose and remember dozens of non-standard ports, so I configure an nginx instance as a reverse
proxy. I'm of the opinion that [all web traffic should be encrypted](https://www.eff.org/encrypt-the-web),
so I also have to provide nginx with trusted certificates to use for each site it reverse proxies.
[Let's Encrypt](https://letsencrypt.org/) makes the process of obtaining free, trusted certificates
approximately a thousand times easier than it was previously, but my workflow still ends up looking
like this:

 1. Create a new config file from a template and save it in `/etc/nginx/sites-available`
 2. Temporarily disable SSL for the site as there's no valid certificate yet
 3. Enable the site by symlinking to it from `/etc/nginx/sites-enabled`
 4. Reload nginx
 5. Run the Let's Encrypt client to obtain certificates
 6. Enable SSL and for the site
 7. Reload nginx

... And that's not including the extra steps when I miss a semi-colon, accidentally skip a step and
have to spend time figuring out why it's not working, or any of the other human-induced problems
that creep in.

I'd been toying with making a script to run through these steps manually for me, if I gave it a
domain name and a reverse proxy target, but I never got around to it. Now I'm moving things to
Docker, though, there's an opportunity to automate the entire thing with no human interaction at
all.

### Existing solutions

It seemed like this probably wasn't a unique idea, so I had a look around for existing solutions.
The most popular by far seems to be [nginx-proxy](https://github.com/jwilder/nginx-proxy) by
Jason Wilder. This is based on his [docker-gen](https://github.com/jwilder/docker-gen) project
that takes a template and populates values from docker containers.

It's a good solution, but there were a few bits I didn't like. Firstly, templates don't really lend
themselves well to every step of the process: to request Let's Encrypt certificates, the
container uses a template to create a shell script which it then sources. Each container that
generates a template also needs access to the Docker socket. Both of those cause an itch in the
back of my head and make me want to say phrases like "attack surface". I don't think there's
actually a problem, but it doesn't really sit well with me.

Secondly, the whole system seems slightly too tightly coupled for my liking. The Let's Encrypt
component needs to modify the nginx config in order to obtain the certificate, while the main
nginx component is also making different changes to add and remove sites. It feels like if it
doesn't just work, it's going to be difficult to debug and pry apart the different components.

Another potential solution is [Rancher](http://rancher.com/). This is a complete platform for
managing containers, and I'm fairly sure if configured right it can grab certificates from
Let's Encrypt and do SSL termination using haproxy. I tried it for a bit but the whole platform
seemed a bit overkill for my purposes, and I didn't want to invest the time I'd need to fully
understand it all.

### Rolling my own

In the end I decided to roll my own solution. Here's a high-level overview of how it all works:

<img src="/res/images/docker/reverse-proxy.png" alt="Diagram">

As you probably noticed, there are quite a few containers involved. Each one performs a small,
well-defined task, and its output can easily be inspected in either a volume or a database. I
think there's some similarity to piping commands together on a command line &mdash; it's a lot
easier to reason about simpler commands like `head`, `cut` and `tr` than it would be one giant
command that combined them. And, if it does go wrong, you can inspect the pipe at each stage to
see where the problem is happening.

#### service-reporter and etcd

The first part of the chain is my [service-reporter](https://github.com/csmith/docker-service-reporter)
container.  This uses the Docker API to get a list of containers, and store information about them
in etcd.  Etcd is a distributed key-value store (similar in some ways to redis or memcached).
The container also watches for containers that are added and removed, and keeps etcd updated
appropriately.

As the service metadata is stored in a database, no other part of the system needs to interact
with Docker. If the Docker API changes, or the host configuration changes, then only this container
has to be updated.

#### service-letsencrypt and letsencrypt-lexicon

The left fork of the diagram deals with obtaining SSL certificates. To keep it separate from the
nginx configuration, it uses DNS-based challenge to prove that we control the domains. It does this
by plumbing together two great open source projects:
[letsencrypt.sh](https://github.com/lukas2511/letsencrypt.sh), a Let's Encrypt client implemented
in bash with support for the dns-01 challenge type, and
[Lexicon](https://github.com/AnalogJ/lexicon), a python library for updating DNS records using a
variety of providers.

My [service-letsencrypt](https://github.com/csmith/docker-service-letsencrypt) container connects
to etcd and pulls a list of containers that have a label with the key `com.chameth.vhost`. It uses
this to build a plain text list of certificates we require (in a format understood by
letsencrypt.sh), and then monitors etcd for changes and repeats as necessary.

The [letsencrypt-lexicon](https://github.com/csmith/docker-letsencrypt-lexicon) container runs
letsencrypt.sh, using Lexicon to perform the required DNS updates, and produces certificates.
The nice thing about this is that it can be used in a completely standalone fashion (you can just
write a domains.txt yourself). It uses `iowait` to watch the domains text file for updates, and
automatically reruns when there are changes. It also runs once a day to renew any certs that are
coming up for expiry.

#### service-nginx and nginx.

The right fork of the diagram is concerned with nginx. My
[service-nginx](https://github.com/csmith/docker-service-nginx) container again connects to etcd
and pulls a list of containers. It uses a couple of labels to determine the vhost, proxy port,
and proxy protocol. It then feeds these values into a template to create a `server` block for
each site, configured with SSL certificates and a reverse proxy setup. The template covers only
the very minimal settings, with the expectation that everything else will be done in the global
config (things such as SSL ciphers, redirection from HTTP, etc).

This container works completely independently of the Let's Encrypt side. You *can* use the
Let's Encrypt containers and mount the certificate volume, or you could just provide your own
certificates. It doesn't really make any difference.

### Putting it all together

The only downside to having many small containers is that it's a bit of a nuisance to get them
all set up. Fortunately, Docker has a solution for this in the form of
[Docker compose](https://docs.docker.com/compose/). This allows you to write a YAML file defining
all of the services you want to run, and bring them up or down in one go. It can handle volumes,
dependencies, networking, etc.  I'll be publishing a docker-compose.yml file to get this entire
stack up and running soon.
