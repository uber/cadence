---
name: workerscaffoldgo02
---

**Gopkg.toml**
```toml
[[constraint]]
  name = "github.com/uber-go/tally"
  version = "3.0.1"

[[constraint]]
  name = "go.uber.org/cadence"
  version = "0.8.0"

[[constraint]]
  name = "go.uber.org/yarpc"
  version = "1.25.1"

[[constraint]]
  name = "go.uber.org/zap"
  version = "1.4.0"

[prune]
  go-tests = true
```
