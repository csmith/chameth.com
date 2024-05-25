---
title: Listing available targets
group: Systemd
---

To see all targets:

```shell
systemctl list-units --type=target --all
```

Or to see the hierarchy that leads to the default target:

```shell
systemd-analyze critical-chain
```
