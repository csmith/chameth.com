---
title: Tests with golden data
group: Go
---

```go
package foo_test

import (
	"filepath"
	"fmt"
	"os"
	"testing"

	"github.com/sebdah/goldie/v2"
)

func TestFoo_GoldenData(t *testing.T) {
	tests := []string{"file1", "file2"}
	gold := goldie.New(t)

	for i := range tests {
		t.Run(tests[i], func(t *testing.T) {
			f, _ := os.Open(filepath.Join("testdata", fmt.Sprintf("%s.ext", tests[i])))
			defer f.Close()
			actual := Foo(f)
			gold.AssertJson(t, tests[i], actual)
		})
	}
}
```