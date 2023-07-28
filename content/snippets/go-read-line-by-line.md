---
title: Read line-by-line
group: Go
---

A `bufio.Scanner` is the most straight forward way. It can also be configured to split on other tokens instead of line-ends.

```go
scanner := bufio.NewScanner(reader)

for scanner.Scan() {
	line := scanner.Text()
	// ...
}

err := scanner.Err()
```