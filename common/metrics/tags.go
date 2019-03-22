package metrics

const domainTag = "domain"

// Tag is an interface to define metrics tags
type Tag interface {
	Key() string
	Value() string
}

// DomainTag wraps a domain tag
type DomainTag struct {
	value string
}

// NewDomainTag returns a new domain tag
func NewDomainTag(value string) DomainTag {
	return DomainTag{value}
}

// Key returns the key of the domain tag
func (d DomainTag) Key() string {
	return domainTag
}

// Value returns the value of a domain tag
func (d DomainTag) Value() string {
	return d.value
}
