package direct

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDiscoverer(t *testing.T) {
	require.NotNil(t, NewDiscoverer(""))
}

func TestDiscovererURL(t *testing.T) {
	actual := NewDiscoverer("test.com")
	require.NotNil(t, actual)
	assert.Equal(t, "test.com", actual.URL())
}
