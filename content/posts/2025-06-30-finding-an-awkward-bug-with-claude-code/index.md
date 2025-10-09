---
date: 2025-06-30
title: "Finding an awkward bug with Claude Code"
permalink: /finding-an-awkward-bug-with-claude-code/
tags: [development, ai, debugging]
format: long
resources:
  - src: confidence.png
    name: "Composite screenshot of 11 different Claude responses that are all very confident at having found the bug"
    title: "The many times Claude thought it was right…"
opengraph:
  image: /finding-an-awkward-bug-with-claude-code/confidence.png
---

{% figure "right" "Composite screenshot of 11 different Claude responses that are all very confident at having found the bug" %}

I recently encountered a bug in one of my projects that I couldn't immediately
figure out. It was an issue in [Centauri](https://github.com/csmith/centauri),
my reverse proxy. After its config was updated, I noticed[^1] it stopped serving
responses. Looking at the logs, I could see it was obtaining new certificates
from Let's Encrypt for a couple of domains, but I'd designed it so that wouldn't
block requests (or so I thought). After a few minutes of confusion, everything
went back to working.

It felt like an issue with locking, but a quick glance didn't show me anything
obvious. I decided to throw the problem at Claude Code and see how it did.
With some guidance[^2] it managed to identify the cause, but it was a bit
of a journey to get there. I'm going to share the prompts and responses,
as it was an interesting process.

### The initial prompt

I thought the issue might be quite obvious, and I'd just overlooked it.
Centauri isn't that complex a project, after all. So I gave Claude a fairly
minimal prompt to set it off:

```text
> Hello! Please examine this codebase. I have noticed an issue that when a
  certificate is being renewed, the application seems to not serve any other
  requests. I suspect there is a misplaced lock somewhere.
 
  Can you see what's happening? Think very hard.
```

Claude Code has [special hardcoded triggers](https://www.anthropic.com/engineering/claude-code-best-practices)
to engage "thinking" mode, which makes it take more time to think about
alternatives. The keywords are: "think", "think hard", "think harder", and
"ultrathink". I managed to not quite get the right incantation here, and only
enabled the lowest level[^3].

<!--more-->

I'll omit all the responses where Claude is "thinking" (talking to itself),
or decides to make todo lists then cross items off them then add different
items, and so on. Its responses get fairly lengthy; I won't judge you too
harshly if you skim over them. Its first response was blessedly brief:

```text
● I found the issue! Looking at the certificate manager code, I can see the
  problem:
```

… and then it went back to thinking and reading files, because it had neither
found the issue nor could it see the problem. Ho-hum. Some time later it came
back with:

```text
● I found the issue! The problem is in certificate/manager.go:53-54:
  
  m.store.LockCertificate(subject, altNames)
  defer m.store.UnlockCertificate(subject, altNames)
  
  The GetCertificate method acquires a lock at the beginning and holds it for the
  entire duration of the certificate renewal process. This includes:
  
  1. Contacting the ACME server
  2. Solving DNS-01 challenges (can take minutes for DNS propagation)
  3. Obtaining the certificate
  4. Getting OCSP staple
  5. Saving to disk
  
  During this time, any other request trying to get the same certificate will be
  blocked. While the lock is per-certificate (not global), it still means that all
  requests for that specific domain are blocked during renewal.
  
  Let me check if there are other places where this might cause blocking:
```

And it then proceeded to do even more thinking, checking files, etc, until
it concluded with:

```text
● I found the root cause! The issue is actually worse than just a
  per-certificate lock. Here's what's happening:
  
  The main blocking issue is in proxy/manager.go:133-144:
   
  func (m *Manager) CheckCertificates() {
      m.lock.RLock()  // ← Acquires read lock on the entire manager
      for i := range m.routes {
          route := m.routes[i]
          // ...
          m.updateCert(route)  // ← This can take MINUTES for ACME renewal!
      }
      m.lock.RUnlock()  // ← Only released after ALL routes are processed
  }
   
  The blocking cascade:
   1. CheckCertificates holds a read lock on the entire proxy manager while
      renewing ALL certificates
   2. Certificate renewal via ACME can take several minutes per certificate
      (DNS propagation, etc.)
   3. Every incoming request calls routeFor (proxy/manager.go:122) which needs
      the same lock:
    func (m *Manager) routeFor(domain string) *Route {
        m.lock.RLock()  // ← BLOCKED until CheckCertificates completes!
        match := m.domains[strings.ToLower(domain)]
        m.lock.RUnlock()
        // ...
    }
  
  So when certificate renewal starts, ALL incoming requests are blocked because
  they can't even look up which route to use. This explains why the application
  doesn't serve any requests during renewal - they're all waiting for the
  manager's read lock.
  
  The issue is compounded by:
  - Certificate renewal happens every 12 hours automatically
  - It also happens when routes are updated (SIGHUP)
  - Each certificate renewal involves time-consuming ACME operations
  
  This is a critical performance bug where I/O operations (certificate renewal)
  are performed while holding a critical lock.
```

This doesn't actually follow on from the earlier response, it's completely
flipped onto a different lock being the issue. Amusingly, it's nearly right
here, but it's highlighting this code for completely the wrong reason.

The lock it's talking about is a `RWMutex`, which is specifically designed to
allow many read-locks. Everything it's pointing out are read locks, so they
cannot block one another as it described. The response Claude has given sounds
very confident, seems appropriately technical, but is in fact complete rubbish.
This is one of the big issue with all LLMs at the minute, and it makes me
despair at the idea of people "vibe coding" things when they don't know enough
to push back against nonsense like this.

### Course correction

I explained briefly to Claude how this basic concurrency primitive works,
and it came right back with another solution:

```text
> But that's a RW Mutex. You can hold many read locks concurrently.

● You're absolutely right! I apologize for the confusion. With a RWMutex,
  multiple readers can hold the lock concurrently. Let me trace through this
  more carefully.
  
● Task(Analyze updateCert method)
  ⎿ Done (8 tool uses · 20.0k tokens · 1m 11.9s)
  
● Excellent analysis! You're right about RWMutex allowing concurrent readers.
  The real issue is the race condition in the certificate store's lock
  management.

  Looking at certificate/store.go:82-91:

  func (j *JsonStore) lockFor(subjectName string, altNames []string) *sync.Mutex {
      key := strings.Join(append([]string{subjectName}, altNames...), ";")

      if mu, ok := j.locks[key]; ok {
          return mu
      } else {
          mu = &sync.Mutex{}
          j.locks[key] = mu  // ← Concurrent map write!
          return mu
      }
  }

  This is a classic concurrent map access bug. When multiple requests hit the
  same domain during certificate operations:

  1. Multiple goroutines call lockFor() simultaneously
  2. They read/write to j.locks map without synchronization
  3. Go runtime can:
    - Panic with "concurrent map writes"
    - Corrupt the map structure
    - Block indefinitely if map internals get corrupted

  This explains the blocking behavior - it's not the RWMutex, but rather a race
  condition causing either panics or deadlocks when the map gets corrupted
  during concurrent certificate operations.

  The fix would be to either:
  1. Add a mutex to protect j.locks map access
  2. Use sync.Map instead of a regular map
  3. Pre-initialize all locks at startup
```

I do appreciate when me stating a basic fact is called out as "excellent
analysis". Definitely doesn't make me feel like I'm having smoke blown up my
ass.

I think the issue it's describing here is actually a problem[^4], albeit one
that's unlikely to be hit. But it's definitely not the one we're looking
for. Claude even describes the possible consequences of the bug, and they
don't match what happened at all.

### Maybe logs will help?

While Claude was chugging away, I'd been looking into the issue as well. I
had logs from when it happened, and figured Claude may be able to glean
something from them too. The first thing I'd noticed is that I wasn't logging
when a process finished, so I'd already added that to the code base, hence
the note so Claude doesn't get itself confused:

```text
> That may be an issue, but it's not the one I was looking for. I've found logs
  from the time it happened. Note that a few logs have been added to the
  codebase since (so we can't tell from the logs when ACME or OCSP processes
  *finished*, just when they started). Here are the logs:
  
  {"time":"22:40:53.187887976Z","level":"INFO","msg":"Received signal, updating routes...","signal":1}
  {"time":"22:40:53.19947532Z","level":"INFO","msg":"Renewing certificate", ...}
  {"time":"22:40:53.199507942Z","level":"INFO","msg":"Starting ACME process to obtain certificate", ...}
  {"time":"22:41:09.048465968Z","level":"INFO","msg":"Renewing certificate", ...}
  {"time":"22:41:09.048482309Z","level":"INFO","msg":"Starting ACME process to obtain certificate", ...}
  {"time":"22:41:17.419299795Z","level":"INFO","msg":"Renewing certificate", ...}
  {"time":"22:41:17.419325744Z","level":"INFO","msg":"Starting ACME process to obtain certificate", ...}
  {"time":"22:41:23.238027386Z","level":"INFO","msg":"Received signal, updating routes...","signal":1}
  {"time":"22:41:29.994330646Z","level":"INFO","msg":"Renewing certificate", ...}
  {"time":"22:41:29.994353329Z","level":"INFO","msg":"Starting ACME process to obtain certificate", ...}
  {"time":"22:41:43.658794811Z","level":"INFO","msg":"Renewing certificate", ...}
  {"time":"22:41:43.658819548Z","level":"INFO","msg":"Starting ACME process to obtain certificate", ...}
  {"time":"22:41:57.635325513Z","level":"INFO","msg":"Renewing certificate", ...}
  {"time":"22:41:57.635347515Z","level":"INFO","msg":"Starting ACME process to obtain certificate", ...}
  {"time":"22:42:10.52439779Z","level":"INFO","msg":"Renewing certificate", ...}
  {"time":"22:42:10.524417107Z","level":"INFO","msg":"Starting ACME process to obtain certificate", ...}
  {"time":"22:42:23.159430249Z","level":"INFO","msg":"Renewing certificate", ...}
  {"time":"22:42:23.159452171Z","level":"INFO","msg":"Starting ACME process to obtain certificate", ...}
  {"time":"22:42:36.347407326Z","level":"INFO","msg":"Renewing certificate", ...}
  {"time":"22:42:36.347433295Z","level":"INFO","msg":"Starting ACME process to obtain certificate", ...}
  {"time":"22:42:52.544628775Z","level":"INFO","msg":"Renewing certificate", ...}
  {"time":"22:42:52.544653993Z","level":"INFO","msg":"Starting ACME process to obtain certificate", ...}
  {"time":"22:43:25.409579107Z","level":"INFO","msg":"Renewing certificate", ...}
  {"time":"22:43:25.409602842Z","level":"INFO","msg":"Starting ACME process to obtain certificate", ...}
  {"time":"22:43:43.819479448Z","level":"INFO","msg":"Renewing certificate", ...}
  {"time":"22:43:43.819502041Z","level":"INFO","msg":"Starting ACME process to obtain certificate", ...}
  {"time":"22:43:54.358957563Z","level":"INFO","msg":"Renewing certificate", ...}
  {"time":"22:43:54.358979355Z","level":"INFO","msg":"Starting ACME process to obtain certificate", ...}
  {"time":"22:44:06.180278128Z","level":"INFO","msg":"Renewing certificate", <big list of alt names>}
  {"time":"22:44:06.180302585Z","level":"INFO","msg":"Starting ACME process to obtain certificate", <big list of alt names>}
  {"time":"22:45:03.830331922Z","level":"WARN","msg":"Failed to connect to upstream","error":"context canceled", ...}
  {"time":"22:45:03.830528134Z","level":"WARN","msg":"Failed to connect to upstream","error":"context canceled", ...}
  {"time":"22:45:03.832388931Z","level":"WARN","msg":"Failed to connect to upstream","error":"context canceled", ...}
  {"time":"22:46:10.225854865Z","level":"WARN","msg":"Failed to connect to upstream","error":"context canceled", ...}
  {"time":"22:46:10.488432434Z","level":"WARN","msg":"Failed to connect to upstream","error":"context canceled", ...}
  {"time":"22:46:10.488439397Z","level":"WARN","msg":"Failed to connect to upstream","error":"context canceled", ...}
  {"time":"22:46:30.617614536Z","level":"WARN","msg":"Failed to connect to upstream","error":"context canceled", ...}
```

I've cut down the logs a bit to remove the specific domains, simplify the
timestamps, and remove a bunch of other cruft. The key points are there were
two SIGHUPs received, and the last certificate before everything started breaking
covered 6 different hostnames.

Claude then went through a phase of periodically shouting "Eureka!" then
deciding it hadn't actually found anything yet:

```text
● Now I see it! The key insight is in the timing and the "context canceled"
  errors. Let me trace through what's happening:
  
● Ah, now I see what's happening! Looking at the logs more carefully:

  1. 22:40:53 - First SIGHUP received, routes updated
  2. 22:40:53 to 22:44:06 - Multiple certificates start renewing (14 certificates!)
  3. 22:41:23 - Another SIGHUP received while renewals are still in progress
  4. 22:45:03 onwards - "context canceled" errors start appearing

  The issue is that SetRoutes is being called AGAIN (at 22:41:23) while the
  previous CheckCertificates() goroutine is still running and renewing
  certificates. Let me check what happens in this scenario:
  
> (The context canceled errors are clients timing out because centauri wasn't
  proxying properly)
  
● Ah, now I understand! The "context canceled" errors are from clients timing
  out because Centauri wasn't proxying their requests. This confirms the
  blocking issue.

  Let me check how incoming requests actually get their certificates:
  
● Now I found the real issue! Looking at the timeline and the code:

  1. 22:40:53 - SIGHUP triggers SetRoutes() which:
    - Calls loadCertificate() for each route
    - Sets route.certificate to the existing certificate
    - Starts go m.CheckCertificates() in background
  2. 22:41:23 - Another SIGHUP while certificates are still renewing
    - SetRoutes() is called AGAIN
    - Creates NEW Route objects with route.certificate = nil initially
    - Replaces the old routes in m.routes
    - The old Route objects (with valid certificates) are replaced!

  The critical issue is in proxy/manager.go:42-65:

  func (m *Manager) SetRoutes(newRoutes []*Route, fallback *Route) error {
      // ... validation ...

      for i := range newRoutes {
          route := newRoutes[i]  // These are NEW Route objects!
          // ...
          m.loadCertificate(route)  // Try to load certificate
      }

      m.lock.Lock()
      m.routes = newRoutes  // REPLACES all routes!
      m.lock.Unlock()
  }

  When the second SIGHUP arrives:
  1. New Route objects are created
  2. For domains currently being renewed, GetExistingCertificate might fail or
     return expired certs
  3. These routes get route.certificate = nil and CertificateMissing status
  4. CertificateForClient returns nil for these domains
  5. TLS handshake fails → clients can't connect → "context canceled" errors

  The fix would be to either:
  1. Prevent concurrent CheckCertificates runs
  2. Copy certificates from existing routes before replacing them
  3. Make certificate renewal truly independent of route management
```

Midway through it had started searching for "context canceled" errors in the
codebase, which doesn't make much sense. That's what happens when a client gives
up waiting, so the corresponding backend request is terminated. You won't
actually find "context canceled" anywhere in Centauri, because it's a feature
of the Go standard library. I gave it a little nudge to try to stop it rabbit
holing.

As far as I can tell, what it's talking about in this response isn't an issue.
I guess it's possible that there's a race condition whereby a route would be
replaced and end up with an old version of a certificate that was just being
renewed, but it wouldn't cause an issue in practice; it'd just get updated
next time the certificates were refreshed.

There's no way for routes to get a `nil` certificate in this scenario, as far
as I know. But I don't have the unfounded confidence of an LLM, so I simply
said:

```text
> I don't follow why the certificate would be nil?
```

### Attempt the third

I expected Claude to just explain its reasoning, but it apparently didn't have
any reasoning so set off on another voyage of discovery:

```text
● You're right to question that. Let me trace through this more carefully.

● Good point! Since GetExistingCertificate doesn't acquire locks, it should
  return the existing certificate even during renewal. Let me dig deeper:
  
● Now I see a potential race condition! Here's what could happen:

● I found it! There's a critical race condition that explains the blocking:

  The Race Condition

  1. Background renewal starts: CheckCertificates() → updateCert() →
     GetCertificate() acquires lock and starts ACME renewal
  2. SIGHUP arrives: SetRoutes() is called, which for each route calls
     loadCertificate() → GetExistingCertificate() → store.GetCertificate()
  3. The race:
    - store.GetCertificate() returns a pointer to the certificate in the slice
    - Meanwhile, the renewal completes and calls SaveCertificate()
    - SaveCertificate() calls removeCertificate() which modifies the slice while
      it's being read
    - This slice modification (append(j.certificates[:i], j.certificates[i+1:]...))
      can cause:
        - The pointer returned by GetCertificate to become invalid
        - Slice bounds issues
        - Memory corruption
  4. The result:
    - GetExistingCertificate might return corrupted data or nil
    - Routes end up with certificate = nil
    - TLS handshakes fail
    - Clients timeout → "context canceled"

  The Timeline Confirms This

  - 22:40:53 - First SIGHUP, starts renewal of 14 certificates
  - 22:41:23 - Second SIGHUP arrives (30 seconds later) while renewals are still running
  - 22:45:03 - First timeout errors (4 minutes after first SIGHUP)

  The 4-minute delay matches ACME renewal time. The second SIGHUP's SetRoutes
  likely read certificates while they were being modified by the first SIGHUP's
  renewal goroutines.

  This is a classic concurrent slice modification bug - no synchronization
  between readers (GetCertificate) and writers (SaveCertificate/removeCertificate).
```

Once again we have a very confident and technical answer that does not explain
the symptoms we're looking for. It's also mostly rubbish again.
`GetCertificate` and `SaveCertificate` are only called by code that is holding
a lock for that certificate; they won't interfere with one another. There may,
again, be a small concurrency issue if different certificates are handled in
parallel, but it's definitely not the cause of our current problems.[^5]

Working through these explanations and debunking them was annoying me at this
point. I decided to give it once last chance, and used the "ultrathink"
shibboleth:

```text
> So there may be a concurrent modification issue, but that doesn't explain the
  symptoms. If a TLS handshake failed it wouldn't result in the clients timing
  out. Likewise if there was any kind of corruption there'd be a panic, not a
  period where it didn't work and then started working again. Engage ultrathink.
```

### Fourth time lucky

I'm going to include a piece of the thinking here because it amused me so much:

```text
✻ Thinking…

  Oh! I need to think about this differently. Let me trace through what happens
  when a request comes in:
```

Apparently it takes "ultrathink" to actually work through the problem logically?
Amusingly it did actually get there:

```text
● You're absolutely right. Let me think more carefully about this...

● EXACTLY! That's the issue. Looking at the timeline:

  1. 22:40:53 - First SIGHUP: CheckCertificates() acquires READ lock, starts
     renewing 14 certificates
  2. 22:41:23 - Second SIGHUP: SetRoutes() tries to acquire write lock, blocks
     waiting
  3. Incoming requests call routeFor() for route lookup, but Go's RWMutex blocks
     new readers when a writer is waiting (to prevent writer starvation)
  4. 22:45:03 - Clients timeout because requests can't even get past route lookup

  The key insight is Go's RWMutex writer starvation prevention: when a writer
  is waiting, new readers are blocked even though the current lock holder is
  also a reader.

  So the cascade is:
  - CheckCertificates (read lock) → certificate renewal takes minutes
  - SetRoutes (write lock) → waits for read lock to release
  - routeFor (read lock for every request) → blocked by waiting writer

  Result: All incoming HTTP requests hang at route lookup, never reach TLS
  handshake or proxying logic. Clients timeout waiting for any response
  → "context canceled".

  This perfectly explains:
  - No panics (no corruption)
  - No immediate TLS errors (never gets that far)
  - Requests hang then timeout (blocked on lock)
  - Eventually recovers (when certificate renewals finish)

  The fix would be restructuring to avoid holding the manager lock during slow
  certificate operations.
```

We're back to the `RWMutex` again, but this time it's dropping knowledge on me
not the other way around: if there's a writer waiting, then new readers are
blocked. This does indeed perfectly explain the symptoms, and was something
I'd overlooked. The write lock is only held briefly to update some struct
members to prevent it from blocking the readers, but this behaviour coupled
with the long-running read lock ends up doing that anyway.

### The fix

After all this work it's very tempting just to ask Claude to fix it, but I
had almost no confidence it'd be able to maintain any of the concurrency
requirements given how much it flailed around with bad explanations. But in the
interests of science, I asked it to try. Here's what it came up with:

```diff
   // CheckCertificates checks and updates the certificates required for registered routes.
   // It should be called periodically to renew certificates and obtain new OCSP staples.
   func (m *Manager) CheckCertificates() {
+    // Create a snapshot of routes to avoid holding the lock during slow certificate operations
     m.lock.RLock()
+    routes := make([]*Route, len(m.routes))
+    copy(routes, m.routes)
+    provider := m.provider
+    m.lock.RUnlock()

+    // Process certificate renewals without holding the manager lock
-    for i := range m.routes {
+    for i := range routes {
-      route := m.routes[i]
+      route := routes[i]
 
       if provider == nil {
         route.certificateStatus = CertificateNotRequired
       } else {
         m.updateCert(route)
       }
     }
-    m.lock.RUnlock()
   }

   // updateCert updates the certificate for the given route.
```

That's about what I expect. It definitely fixes the exact issue described, but
it's a bit… sloppy. Routes are now being updated when the lock isn't being held,
which is a change in semantics. It also doesn't address the underlying issue
that requests can be blocked if there's a pending writer, and pending writes
can be blocked by readers. A better approach would be to allow these operations
to happen without blocking, maybe by using an atomic pointer to an immutable
slice. I guess I've got some coding to do!

As for Claude: while the journey was frustrating, I think trying to figure this
out by myself would have been even more so. It's one of those things where
everything _looks_ right to me, and I'd have to go down a fairly long chain
of checking my assumptions to find where I'd gone wrong.

There's definitely a lot of room for improvement in coding LLMs: being able
to estimate their own confidence level instead of stating everything as
absolute truth would go a long way. I imagine there's also a lot of benefit
of having one LLM vet the responses of another; that's sort-of how thinking
mode works, but it doesn't seem to go far enough to stop it outputting rubbish.
It'll be interesting to see how they develop in the future, anyway.

[^1]: read as "I got spammed by a bunch of monitoring notifications"
[^2]: read as "Me repeatedly telling it that it was talking rubbish"
[^3]: If only we had some kind of language model technology that could understand this kind of thing.
[^4]: I sense some fun times with Go's race detector in my near future
[^5]: The most likely outcome from these concurrency issues is Centauri panicking,
      which isn't great but would only interrupt ongoing connections before it
      restarts and starts serving traffic again. Much less problematic than the
      "block for five minutes" behaviour I'm trying to fix.