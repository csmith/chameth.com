---
title: List creation dates of tags
group: Git
---

```shell
git for-each-ref --sort=creatordate --format '%(refname) %(creatordate)' refs/tags
```