---
title: Tidy up merged branches
group: Git
---

Remove all branches that have been merged into `master`:

```shell
git branch -d $(git branch --merged=master | grep -v master)
```