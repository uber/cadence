package membership

import "github.com/hashicorp/serf/serf"

// NewSerfMonitor returns a new serf based monitor
func NewSerfMonitor() {
	serf.Create(nil)
}
