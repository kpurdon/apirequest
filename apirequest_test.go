package apirequest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kpurdon/apirequest/discoverers/direct"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: need a more complete and well formed test ... this is just to ensure things are working
func TestExecute(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		requester := NewRequester(t.Name(), nil)
		requester.MustAddAPI("test", direct.NewDiscoverer(ts.URL))
		req, err := requester.NewRequest("test", http.MethodGet, "/")
		require.NoError(t, err)
		ok, err := requester.Execute(req, nil, nil)
		assert.True(t, ok)
		assert.NoError(t, err)
	})

	t.Run("with success data", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"test":"test"}`))
		}))
		defer ts.Close()

		type data struct {
			Test string `json:"test"`
		}
		var d data

		requester := NewRequester(t.Name(), nil)
		requester.MustAddAPI("test", direct.NewDiscoverer(ts.URL))
		req, err := requester.NewRequest("test", http.MethodGet, "/")
		require.NoError(t, err)
		ok, err := requester.Execute(req, &d, nil)
		assert.True(t, ok)
		assert.NoError(t, err)
	})

	t.Run("with handled error data", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"test":"test"}`))
		}))
		defer ts.Close()

		type data struct {
			Test string `json:"test"`
		}
		var d data

		requester := NewRequester(t.Name(), nil)
		requester.MustAddAPI("test", direct.NewDiscoverer(ts.URL))
		req, err := requester.NewRequest("test", http.MethodGet, "/")
		require.NoError(t, err)
		ok, err := requester.Execute(req, nil, &d)
		assert.False(t, ok)
		assert.NoError(t, err)
		assert.Equal(t, "test", d.Test)
	})

	t.Run("with un-handled error data", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"test":"test"}`))
		}))
		defer ts.Close()

		requester := NewRequester(t.Name(), nil)
		requester.MustAddAPI("test", direct.NewDiscoverer(ts.URL))
		req, err := requester.NewRequest("test", http.MethodGet, "/")
		require.NoError(t, err)
		ok, err := requester.Execute(req, nil, nil)
		assert.False(t, ok)
		assert.Error(t, err)
	})
}
