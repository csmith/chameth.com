---
title: Register domains as secure contexts
group: Firefox
---

If you want a domain to be treated as a secure context (and get access to all
the sensitive JavaScript APIs that are gated behind that arbitrary designation),
you can add it to the pref:

```
dom.securecontext.allowlist
```

Unfortunately due to a [bug](https://bugzilla.mozilla.org/show_bug.cgi?id=1918915),
Firefox will then try to upgrade any resource requests to HTTPS, and you'll probably
end up with no images. The workaround for now is to disable upgrades entirely:

```
security.mixed_content.upgrade_display_content false
```
