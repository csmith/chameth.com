---
title: Get symbolic names for keycodes
group: X
---

`xmodmap` can print out its current keymap including symbolic names:

```shell
xmodmap -pke
```

Which gives output like:

```text
keycode 171 = XF86AudioNext NoSymbol XF86AudioNext
keycode 172 = XF86AudioPlay XF86AudioPause XF86AudioPlay XF86AudioPause
keycode 173 = XF86AudioPrev NoSymbol XF86AudioPrev
keycode 174 = XF86AudioStop XF86Eject XF86AudioStop XF86Eject
```