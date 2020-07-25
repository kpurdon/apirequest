package apirequest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/kpurdon/apirequest/discoverers/direct"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	testCases := []struct {
		label           string
		client          *http.Client
		expectedTimeout time.Duration
	}{
		{
			label:           "defaults",
			client:          nil,
			expectedTimeout: time.Duration(0),
		}, {
			label:           "custom httpclient",
			client:          &http.Client{Timeout: 10 * time.Second},
			expectedTimeout: 10 * time.Second,
		},
	}

	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			t.Log(tc.label)

			c := NewClient("test", tc.client)
			require.NotNil(t, c)

			assert.Equal(t, "test", c.apiName)
			assert.NotNil(t, c.apis, "apis is an initialized map")
			assert.Len(t, c.apis, 0, "apis is an empty map")
			assert.Equal(t, tc.expectedTimeout, c.Client.Timeout)
		})
	}
}

func TestClientMustAddAPI(t *testing.T) {
	c := NewClient("test", nil)
	c.MustAddAPI("test1", direct.NewDiscoverer("test1"))
	c.MustAddAPI("test2", direct.NewDiscoverer("test2"))

	for k, actual := range c.apis {
		require.NotNil(t, actual)
		assert.Equal(t, k, actual.URL())
	}

	assert.Panics(t, func() {
		c.MustAddAPI("test2", direct.NewDiscoverer("test2"))
	})
}

func TestClientNewRequest(t *testing.T) {
	baseURL := "http://127.0.0.1"

	testCases := []struct {
		label         string
		apiName       string
		method        string
		url           string
		expectedError error
		expectedURL   string
	}{
		{
			label:         "api not initialized",
			apiName:       "notanapi",
			method:        http.MethodGet,
			url:           "",
			expectedError: errors.New("api [notanapi] not initialized"),
			expectedURL:   fmt.Sprintf("%s/", baseURL),
		}, {
			label:         "strips leading slash",
			apiName:       "test",
			method:        http.MethodGet,
			url:           "/foo/bar",
			expectedError: nil,
			expectedURL:   fmt.Sprintf("%s/foo/bar", baseURL),
		}, {
			label:         "method set",
			apiName:       "test",
			method:        http.MethodPost,
			url:           "foo/bar",
			expectedError: nil,
			expectedURL:   fmt.Sprintf("%s/foo/bar", baseURL),
		}, {
			label:         "basic",
			apiName:       "test",
			method:        http.MethodGet,
			url:           "foo/bar",
			expectedError: nil,
			expectedURL:   fmt.Sprintf("%s/foo/bar", baseURL),
		},
	}

	c := NewClient("test", nil)
	require.NotNil(t, c)
	c.MustAddAPI("test", direct.NewDiscoverer(baseURL))

	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			t.Log(tc.label)

			actual, err := c.NewRequest(tc.apiName, tc.method, tc.url)
			if err != nil {
				assert.EqualError(t, tc.expectedError, err.Error())
				assert.Nil(t, actual)
				return
			}
			require.NotNil(t, actual)

			assert.Equal(t, tc.expectedURL, actual.Request.URL.String())
			assert.Nil(t, actual.Request.Body)
			assert.Equal(t, tc.method, actual.Request.Method)
		})
	}
}

func TestRequestSetBody(t *testing.T) {
	testCases := []struct {
		label       string
		body        interface{}
		expectError bool
	}{
		{
			label:       "valid body",
			body:        &struct{ Test string }{Test: "test"},
			expectError: false,
		}, {
			label:       "invalid non-nil body",
			body:        make(chan int),
			expectError: true,
		}, {
			label:       "invalid nil body",
			body:        nil,
			expectError: true,
		},
	}

	c := NewClient("test", nil)
	require.NotNil(t, c)
	c.MustAddAPI("test", direct.NewDiscoverer("http://127.0.0.1"))

	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			t.Log(tc.label)

			req, err := c.NewRequest("test", http.MethodGet, "foo/bar")
			require.NoError(t, err)
			require.NotNil(t, req)

			err = req.SetBody(tc.body)
			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, "application/json", req.Request.Header.Get("Content-Type"))

			actualBody, err := ioutil.ReadAll(req.Request.Body)
			require.NoError(t, err)

			expectedBody, err := json.Marshal(tc.body)
			require.NoError(t, err)

			assert.Equal(t, string(expectedBody), string(actualBody))
		})
	}
}

func TestRequestSetQueryParams(t *testing.T) {
	testCases := []struct {
		label  string
		params url.Values
	}{
		{
			label:  "single param",
			params: url.Values{"one": []string{"onev"}},
		}, {
			label: "two params",
			params: url.Values{
				"one": []string{"onev"},
				"two": []string{"twov"},
			},
		},
	}

	c := NewClient("test", nil)
	require.NotNil(t, c)
	c.MustAddAPI("test", direct.NewDiscoverer("http://127.0.0.1"))

	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			t.Log(tc.label)

			req, err := c.NewRequest("test", http.MethodGet, "foo/bar")
			require.NoError(t, err)
			require.NotNil(t, req)

			req.SetQueryParams(tc.params)
			require.Equal(t, tc.params, req.Request.URL.Query())
		})
	}
}

func TestRequestSetUserAgent(t *testing.T) {
	testCases := []struct {
		label string
		ua    string
	}{
		{
			label: "empty",
			ua:    "", // TODO: should this use the default?
		}, {
			label: "basic",
			ua:    "test",
		},
	}

	c := NewClient("test", nil)
	require.NotNil(t, c)
	c.MustAddAPI("test", direct.NewDiscoverer("http://127.0.0.1"))

	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			t.Parallel()
			t.Log(tc.label)

			req, err := c.NewRequest("test", http.MethodGet, "foo/bar")
			require.NoError(t, err)
			require.NotNil(t, req)

			req.SetUserAgent(tc.ua)
			assert.Equal(t, tc.ua, req.Request.UserAgent())
		})
	}
}

func TestClientExecute(t *testing.T) {
	// TODO: refactor this test

	t.Run("simple", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		c := NewClient(t.Name(), nil)
		c.MustAddAPI("test", direct.NewDiscoverer(ts.URL))
		req, err := c.NewRequest("test", http.MethodGet, "/")
		require.NoError(t, err)
		ok, err := c.Execute(req, nil, nil)
		assert.True(t, ok)
		assert.NoError(t, err)
	})

	t.Run("with success data", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`{"test":"test"}`))
			require.NoError(t, err)
		}))
		defer ts.Close()

		type data struct {
			Test string `json:"test"`
		}
		var d data

		c := NewClient(t.Name(), nil)
		c.MustAddAPI("test", direct.NewDiscoverer(ts.URL))
		req, err := c.NewRequest("test", http.MethodGet, "/")
		require.NoError(t, err)
		ok, err := c.Execute(req, &d, nil)
		assert.True(t, ok)
		assert.NoError(t, err)
	})

	t.Run("with handled error data", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`{"test":"test"}`))
			require.NoError(t, err)
		}))
		defer ts.Close()

		type data struct {
			Test string `json:"test"`
		}
		var d data

		c := NewClient(t.Name(), nil)
		c.MustAddAPI("test", direct.NewDiscoverer(ts.URL))
		req, err := c.NewRequest("test", http.MethodGet, "/")
		require.NoError(t, err)
		ok, err := c.Execute(req, nil, &d)
		assert.False(t, ok)
		assert.NoError(t, err)
		assert.Equal(t, "test", d.Test)
	})

	t.Run("with un-handled error data", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")
			_, err := w.Write([]byte(`{"test":"test"}`))
			require.NoError(t, err)
		}))
		defer ts.Close()

		c := NewClient(t.Name(), nil)
		c.MustAddAPI("test", direct.NewDiscoverer(ts.URL))
		req, err := c.NewRequest("test", http.MethodGet, "/")
		require.NoError(t, err)
		ok, err := c.Execute(req, nil, nil)
		assert.False(t, ok)
		assert.Error(t, err)
	})
}
