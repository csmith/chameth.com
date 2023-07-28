---
title: Listing files recursively
group: Go
---

`WalkDir` is much easier than chaining together `ReadDir` calls:

```go
var files []string
err := filepath.WalkDir(dir, func(path string, info fs.DirEntry, err error) error {
  if !info.IsDir() {
    files = append(files, info.Name())
  }
  return nil
})
```