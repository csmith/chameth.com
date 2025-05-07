---
title: Reverse part of a slice
group: Go
---

Performs an in-place reversal. Could be extended to take an end position as well.

```go
func reverse(input []byte, start int) []byte {
	for left, right := start, len(input)-1; left < right; left, right = left+1, right-1 {
		input[left], input[right] = input[right], input[left]
	}
	return input
}
```