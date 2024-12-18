---
title: Ignore revisions in git-blame
group: Git
---

Create a file containing the full commit object names to ignore. Usually it's
named `.git-blame-ignore-revs`. Comments and whitespace are ignored.

```gitexclude
# Reformatting
7cc8f43b2b8bbad1af306f50e1da2f28bf5d2046
```

Then either specify it with `--ignore-revs-file` each time:

```shell
git blame --ignore-revs-file .git-blame-ignore-revs -- ...
```

Or configure git to always use it:

```shell
git config blame.ignoreRevsFile .git-blame-ignore-revs
```