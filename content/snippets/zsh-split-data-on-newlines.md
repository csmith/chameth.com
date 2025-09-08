---
title: Split data on newlines
group: zsh
---

zsh has a nice shortcut for splitting data on newlines, instead of ugly
`read`/`IFS` hacks, by using the `f`
[parameter expansion flag](https://zsh.sourceforge.io/Doc/Release/Expansion.html#Parameter-Expansion-Flags),
e.g.:

```shell
for i in "${(f)$(ls -l)}"; do
  echo "$i";
done
```