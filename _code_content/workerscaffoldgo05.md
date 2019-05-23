---
name: workerscaffoldgo05
---

**main.go**
```go
package main

import (
	"fmt"

	"myorg/cadence-scaffold/common"
)

func main() {
	var h common.SampleHelper
	h.SetupServiceConfig()

	fmt.Println("Success!")
}
```
