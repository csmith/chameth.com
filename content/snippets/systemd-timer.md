---
title: Running a service on a cron-like timer
group: Systemd
---

Create a service, e.g. `~/.config/systemd/user/cron-thing.service`:

```systemd
[Unit]
Description=Do thing

[Service]
ExecStart=/path/to/script.sh
```

Create a timer, e.g. `~/.config/systemd/user/cron-thing.timer`:

```systemd
[Unit]
Description=Run thing on timer

[Timer]
OnCalendar=*:0/15

[Install]
WantedBy=timers.target
```

(The calendar format is documented in [systemd.time(7)](https://www.freedesktop.org/software/systemd/man/latest/systemd.time.html#Calendar%20Events))

Enable and start the timer.

