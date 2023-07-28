---
title: Check if a string is an integer
group: Go
---

Just try to convert it with `Atoi`:

```go
if num, err := strconv.Atoi(str); err == nil {
    // ...
}
```