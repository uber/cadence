package types

// RedirectRequest interface for a redirect-able request
type RedirectRequest interface {
	ToRedirectKey() RedirectKey
}

// RedirectKey entity of keys used for redirect
type RedirectKey struct {
	ShardID     *int
	WorkflowID  string
	DomainID    string
	HostAddress string
}
