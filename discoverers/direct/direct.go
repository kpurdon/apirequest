// Package direct implements the kpurdon/apirequest.Discoverer interface for a direct URL with no
// pooling or discovery.
package direct

// Discoverer implements the apirequest.Discoverer interface.
type Discoverer struct {
	baseURL string
}

// NewDiscoverer returns a new Discoverer for the given API base URL.
func NewDiscoverer(baseURL string) *Discoverer {
	return &Discoverer{baseURL: baseURL}
}

// URL returns the direct API URL.
func (d Discoverer) URL() string {
	return d.baseURL
}
