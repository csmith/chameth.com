---
title: Enable SQLite logging
group: Android
---

To log information about any SQLite query:

```shell
$ adb shell setprop log.tag.SQLiteStatements VERBOSE
```