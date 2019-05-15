---
date: 2019-05-08
title: Debugging beyond the debugger
url: /2019/05/08/debugging-beyond-the-debugger/
image: /res/images/debugging/strace.png
description: What happens when your usual approach fails you?
area: troubleshooting
---

Most programming -- and sysadmin -- problems can be debugged in a
fairly straight forward manner using logs, print statements,
educated guesses, or an actual debugger. Sometimes, though, the
problem is more elusive. There's a wider box of tricks that can
be employed in these cases but I've not managed to find a nice
overview of them, so here's mine. I'm mainly focusing on Linux
and similar systems, but there tend to be alternatives available
for other Operating Systems or VMs if you seek them out.

## Networking

### tcpdump

`tcpdump` prints out descriptions of packets on a network interface. You can
apply filters to limit which packets are displayed, chose to dump the entire
content of the packet, and so forth.

<!--more-->

Typical usage might look something like:

{{< highlight console >}}
# tcpdump -nSi eth0 port 80
tcpdump: verbose output suppressed, use -v or -vv for full protocol decode
listening on eth0, link-type EN10MB (Ethernet), capture size 262144 bytes
16:03:35.577781 IP6 2001:db8::1.54742 > 2001:db8::2.80: Flags [S], seq 2815779044, win 64800, options [mss 1440,sackOK,TS val 2378811665 ecr 0,nop,wscale 7], length 0
16:03:35.586853 IP6 2001:db8::2.80 > 2001:db8::1.54742: Flags [S.], seq 1522609102, ack 2815779045, win 28560, options [mss 1440,sackOK,TS val 3063610173 ecr 2378811665,nop,wscale 7], length 0
16:03:35.586877 IP6 2001:db8::1.54742 > 2001:db8::2.80: Flags [.], ack 1522609103, win 507, options [nop,nop,TS val 2378811674 ecr 3063610173], length 0
16:03:35.620678 IP6 2001:db8::1.54742 > 2001:db8::2.80: Flags [P.], seq 2815779045:2815779399, ack 1522609103, win 507, options [nop,nop,TS val 2378811708 ecr 3063610173], length 354: HTTP: GET / HTTP/1.1
{{< / highlight >}}

Here you can see the start of a plain text HTTP request: the three-way
handshake as the TCP connection is established followed by a GET request.
Even if the data is encrypted as it will be in most cases, it's often useful
to see the "shape" of the transmissions -- did the client start sending data
when it connected, did the server ever respond, etc.

[Daniel Miessler has a good tutorial on tcpdump](https://danielmiessler.com/study/tcpdump/)
if you're not familiar with it and don't want to jump straight into the man
page.

#### ... with Docker

Docker sets up separate network namespaces for each container. To see the
traffic across the interfaces of a single container you can `nsenter` the
container's network namespace:

{{< highlight console >}}
# nsenter -t $(docker inspect --format '{{.State.Pid}}' my_container) -n tcpdump -nS port 80
{{< / highlight >}}

This retrieves the PID for the container, and tells `nsenter` to enter the
network (`-n`) namespace from the given target (`-t`) PID, and then run the
given command (in this case `tcpdump ...`).

### openssl s_client / s_server

When a connection is using TLS it's often useful to try connecting to the
server and see what certificate it presents, algorithms it negoitates, and
so forth. OpenSSL offers two useful subcommands which can help with this:
`s_client` for connecting as a client, and `s_server` for listening to
connections.

For example using `s_client` to connect to `google.com` on the standard
HTTPS port shows us details about the server cert and its verification
status:

{{< highlight console >}}
$ openssl s_client -connect google.com:443
CONNECTED(00000003)
depth=2 OU = GlobalSign Root CA - R2, O = GlobalSign, CN = GlobalSign
verify return:1
depth=1 C = US, O = Google Trust Services, CN = Google Internet Authority G3
verify return:1
depth=0 C = US, ST = California, L = Mountain View, O = Google LLC, CN = *.google.com
verify return:1
---
Certificate chain
 0 s:C = US, ST = California, L = Mountain View, O = Google LLC, CN = *.google.com
   i:C = US, O = Google Trust Services, CN = Google Internet Authority G3
 1 s:C = US, O = Google Trust Services, CN = Google Internet Authority G3
   i:OU = GlobalSign Root CA - R2, O = GlobalSign, CN = GlobalSign
---
# ...
{{< / highlight >}}

Whereas connecting to my webserver and providing an unknown host in the SNI
field results in an SSL alert 112 ("The server name sent was not recognized")
and no server certificate is sent:

{{< highlight console >}}
$ openssl s_client -connect chameth.com:443 -servername example.com
CONNECTED(00000003)
140384831313024:error:14094458:SSL routines:ssl3_read_bytes:tlsv1 unrecognized name:../ssl/record/rec_layer_s3.c:1536:SSL alert number 112
---
no peer certificate available
---
# ...
{{< / highlight >}}

Often if you hit this kind of alert in an application the exact error will be
lost somewhere in the many layers between the SSL library and the logs, so
being able to directly connect and test can help diagnose a lot of issues.

Once a connection is established you can read and write plain text and it
will be encrypted and decrypted automatically.

### Java apps

If a Java app is involved in the connection, you can enable a lot of built-in
debugging with a simple JVM property: `javax.net.debug`. You can tweak
what exactly gets logged, but the easiest thing to do is just set the property
to `all` and you'll see information about certificate chains, verification,
and packet dumps:

{{< highlight console >}}
$ java -Djavax.net.debug=all -jar ....
# ...
found key for : duke
chain [0] = [
[
  Version: V1
  Subject: CN=Duke, OU=Java Software, O="Sun Microsystems, Inc.",
  L=Cupertino, ST=CA, C=US
# ...
{{< / highlight >}}

More information about Java's debugging options is available on
[docs.oracle.com](https://docs.oracle.com/javase/7/docs/technotes/guides/security/jsse/ReadDebug.html).

## Thread and core dumps

Higher-level languages frequently provide an interactive way to dump the
current executation state of all of their threads (a "thread dump"). This
is useful to spot deadlocks, some types of race conditions, and as a
quick and dirty method of investigating hangs or excessive CPU usage.

With both Java and Go applications you can send a QUIT signal to have a
thread dump printed out; Go applications will quit after doing so, Java
ones will carry on running. At most terminals you can hit `Ctrl` and `\` to
send a QUIT signal.

For Java you can also use the `jstack` tool from the JDK to dump threads
by PID; this can be useful if the application is running in the background
or has redirected sysout:

{{< highlight console >}}
$ jstack 8321
Attaching to process ID 8321, please wait...
Debugger attached successfully.
Client compiler detected.

Thread t@5: (state = BLOCKED)
 - java.lang.Object.wait(long) @bci=-1107318896 (Interpreted frame)
 - java.lang.Object.wait(long) @bci=0 (Interpreted frame)
 - java.lang.ref.ReferenceQueue.remove(long) @bci=44, line=116 (Interpreted frame)
 - java.lang.ref.ReferenceQueue.remove() @bci=2, line=132 (Interpreted frame)
 - java.lang.ref.Finalizer$FinalizerThread.run() @bci=3, line=159 (Interpreted frame)

# ...
{{< / highlight >}}

A core dump provides more complete information about the state of a process,
but is often more complex to interpret. The `gcore` utility from GDB will
create a core dump of a process with a given PID. You can then generally
load the core file using your normal debugger, depending on the language
in question.

## System calls

`strace` is the swiss army knife for seeing what a process is doing. It
details each system call made by a program (you can filter them down, of
course). For example:

{{< highlight console >}}
$ strace -e read curl https://google.com/
read(3, "\177ELF\2\1\1\0\0\0\0\0\0\0\0\0\3\0>\0\1\0\0\0 \236\0\0\0\0\0\0"..., 832) = 832
read(3, "\177ELF\2\1\1\0\0\0\0\0\0\0\0\0\3\0>\0\1\0\0\0P!\0\0\0\0\0\0"..., 832) = 832
read(3, "\177ELF\2\1\1\3\0\0\0\0\0\0\0\0\3\0>\0\1\0\0\0\200l\2\0\0\0\0\0"..., 832) = 832
read(3, "\177ELF\2\1\1\0\0\0\0\0\0\0\0\0\3\0>\0\1\0\0\0\20Q\0\0\0\0\0\0"..., 832) = 832
# ...
read(3, "\0\0\0\0\0\0\0\4\25\345\366\302\273sE6\365wI\225\321|\3435Z\362\216\372\215\251aO"..., 253) = 253
<HTML><HEAD><meta http-equiv="content-type" content="text/html;charset=utf-8">
<TITLE>301 Moved</TITLE></HEAD><BODY>
<H1>301 Moved</H1>
The document has moved
<A HREF="https://www.google.com/">here</A>.
</BODY></HTML>
read(3, "\27\3\3\0!", 5)                = 5
# ...
{{< / highlight >}}

[Brendan Gregg](http://www.brendangregg.com/blog/2014-05-12/strace-wow-much-syscall.html)
has a nice guide on `strace` and alternatives.

### ... with docker

When the application is running in docker you can usually just `strace` it
from the host with the correct PID
(from e.g. `docker inspect --format '{{.State.Pid}}' my_container`).
Sometimes you may need to trace the startup of an application though, which is
a bit trickier. Instead you can run a new container using the same PID
namespace as your target, and the permissions needed to `strace`:

{{< highlight console >}}
$ docker run --rm -it --pid=container:my_container \
  --net=container:my_container \
  --cap-add sys_admin \
  --cap-add sys_ptrace \
  alpine
{{< / highlight >}}

From within the new container you can install strace, and trace any running
program within the target container using `strace -p` as normal. To start a
new program you need access to the target container's file system, which you
can get to via `/proc/1/root` (PID `1` being the main process that docker
started in the target container).

## Files

Sometimes the problem might relate to file access. There are a couple of
straight forward - but nonetheless useful - tools which might help here.
`inotifywait` uses the Linux `inotify` subsystem to watch files or directories
for operations. For example:

{{< highlight console >}}
$ inotifywait -mr site/content
Setting up watches.  Beware: since -r was given, this may take a while!
Watches established.
site/content/post/ MODIFY 2019-05-08-debugging-beyond-the-debugger.md
site/content/post/ OPEN 2019-05-08-debugging-beyond-the-debugger.md
site/content/post/ MODIFY 2019-05-08-debugging-beyond-the-debugger.md
site/content/post/ MODIFY 2019-05-08-debugging-beyond-the-debugger.md
site/content/post/ CLOSE_WRITE,CLOSE 2019-05-08-debugging-beyond-the-debugger.md
# ...
{{< / highlight >}}

Here the `-m` switch makes `inotifywait` monitor the files forever (instead
of exiting on the first modification, which is the normal behaviour) and `r`
makes it recurse into the directory and monitor each file and subdirectory in
there.

If you want to see what processes currently have a file open, `fuser` is the
go-to tool. For example:

{{< highlight console >}}
$ fuser -v /
                     USER PID ACCESS COMMAND
/:                   root     kernel mount /
                     chris      2961 .rc.. systemd
                     chris      2986 .r... gdm-x-session
                     chris      2994 .r... dbus-daemon
                     chris      3001 .r... gnome-session-b
# ...
{{< / highlight >}}

## Honourable mentions

These aren't really debugging tools, but I feel it's worth mentioning as
they often feature somewhere along the debugging-of-weird-problems journey.

I've seen some weird and wonderful problems happen
because a disk is full, so a quick `df` early on in the debugging process
never hurts. Some apps may hang, some may corrupt their config, some may
fall over and die; sometimes the manner in which they fail doesn't obviously
point to a disk space issue.

Another issue that comes up now and then -- especially inside VMs or
other environment that don't have a decent amount of "noise" happening --
is entropy exhaustion. A quick look at `/proc/sys/kernel/random/entropy_avail`
should be enough to confirm that everything is ticking along nicely. If it's
exceedingly low then you may find that anything involving random number
generation stalls (TLS connections for example).
