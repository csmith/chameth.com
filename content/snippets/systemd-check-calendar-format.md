---
title: Check calendar event formats
group: Systemd
---

```shell
$ systemd-analyze calendar "*:0/15"
  Original form: *:0/15
Normalized form: *-*-* *:00/15:00
    Next elapse: Wed 2024-01-24 22:15:00 GMT
       (in UTC): Wed 2024-01-24 22:15:00 UTC
       From now: 9min left
```