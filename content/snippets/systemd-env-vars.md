---
title: Passing env vars to a service
group: Systemd
---

If you want to run a service that's usually configured using env vars, you can define them in a nice location like `/etc/default/$service.conf` and tell systemd to load them:

```systemd
EnvironmentFile=-/etc/default/service.conf
PassEnvironment="ENVVAR1 ENVAR2"
```