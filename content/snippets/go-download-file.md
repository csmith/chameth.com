---
title: Download a file
group: Go
---

Does a `http.Get` and copies the body to a local file:

```go
func download(url, target string) error {
    // Optionally:
	if _, err := os.Stat(target); err == nil {
		// File already exists, don't bother redownloading
		return nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(target)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
```