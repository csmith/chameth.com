---
title: Stop LFS requiring multiple security key touches
group: Git
---

In a git-lfs enabled repo, pushing can require you to interact with a security
key many times (up to 5 or so) even if you're not touching LFS content.

The workaround is to disable multiplexing:

```shell
git config --global lfs.ssh.autoMultiplex false
```

and then configure SSH to persist a control connection:

```text
Host github.com
    ControlPath ~/.ssh/controlmaster-%C.sock
    ControlMaster auto
    ControlPersist 5m
```
