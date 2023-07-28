---
title: Listing Files in a directory
group: Go
---

Go 1.16 and above:

```go
files, err := os.ReadDir("./")
if err != nil {
    log.Fatal(err)
}

for _, f := range files {
    fmt.Println(f.Name())
}
```

Go 1.15 and below:

```go
files, err := ioutil.ReadDir("./")
if err != nil {
    log.Fatal(err)
}

for _, f := range files {
    fmt.Println(f.Name())
}
```