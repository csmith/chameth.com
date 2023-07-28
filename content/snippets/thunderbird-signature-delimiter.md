---
title: Remove signature delimiter
group: Thunderbird
---

To get rid of the `--` that Thunderbird inserts above a signature, go to the config editor (Preferences -> Advanced -> Config Editor) and set the following pref to `true`:

```
mail.identity.default.suppress_signature_separator
```