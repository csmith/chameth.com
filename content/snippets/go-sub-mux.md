---
title: Use nested ServeMux with path stripping
group: Go
---

If you want to nest multiple ServeMux within each other, stripping the path at
each level, you tend to end up with problems with trailing slashes and
redirects.

This function:

- Registers both "/path" and "/path/" to ensure that the mux doesn't redirect
- Strips the prefix
- Re-adds a "/" if the prefix became empty

```go
func addSubmux(mux *http.ServeMux, path string, submux *http.ServeMux) {
    wrapper := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path = strings.TrimPrefix(r.URL.Path, path); r.URL.Path == "" {
            r.URL.Path = "/"
        }

        submux.ServeHTTP(w, r)
    })

    mux.Handle(path+"/", wrapper)
    mux.Handle(path, wrapper)
}
```