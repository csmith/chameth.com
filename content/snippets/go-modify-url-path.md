---
title: Modify path of a URL
group: Go
---

Using `path.Join` directly doesn't work as it will break the scheme. Instead, parse the URL and only perform path operations on the path element.

```go
u, err := url.Parse(target)
if err != nil {
	// ... 
}
u.Path = path.Join(path.Dir(u.Path), relativePath)
str := u.String()
```