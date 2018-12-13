// Package apirequest provides a simple helper for making API requests to HTTP APIs.
package apirequest

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// Discoverer defines an interface for discovering HTTP APIs. The discoverers package provides some
// pre-defined implementations for service discovery.
type Discoverer interface {
	URL() string
}

// Requester is the base type which all helper methods hang on. It should be initialized with the
// NewRequester method.
type Requester struct {
	apiName string
	apis    map[string]Discoverer
	*http.Client
}

// NewRequester initializes a new Requester.
func NewRequester(apiName string, client *http.Client) *Requester {
	if client == nil {
		client = http.DefaultClient
	}
	return &Requester{
		apiName: apiName,
		apis:    make(map[string]Discoverer),
		Client:  client,
	}
}

// MustAddAPI adds a named API with a Discoverer that will be used when attempting to make requests
// to the API.
func (r *Requester) MustAddAPI(apiName string, discoverer Discoverer) {
	if _, ok := r.apis[apiName]; ok {
		panic(fmt.Sprintf("api [%s] already initialized", apiName))
	}

	// TODO: ping the API in some way to ensure connection at startup?

	r.apis[apiName] = discoverer
}

// Request wraps a *http.Request and allows post-creating setting of various properties of the request.
type Request struct {
	*http.Request
}

// SetQueryParams sets the URL.RawQuery (query params) by encoding the given url.Values.
func (r *Request) SetQueryParams(ps url.Values) {
	r.URL.RawQuery = ps.Encode()
}

// SetBody takes a non-nil struct, marshals it to JSON, and sets it as the requests body. It
// additionally sets the Content-Type header to application/json.
// TODO: potentially support more Content-Types
func (r *Request) SetBody(body interface{}) error {
	if body == nil {
		return errors.New("body must be non-nil")
	}

	r.Header.Set("Content-Type", "application/json")

	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(bytes.NewReader(b))

	return nil
}

// SetUserAgent sets the User-Agent header to a custom value. If this is not set the default will
// be used.
func (r *Request) SetUserAgent(ua string) {
	r.Header.Set("User-Agent", ua)
}

// NewRequest creates a new http.Request using the Discoverer for the given API name to get the
// base URL of the API.
func (r *Requester) NewRequest(apiName, method, url string) (*Request, error) {
	api, ok := r.apis[apiName]
	if !ok {
		return nil, fmt.Errorf("api [%s] not initialized", apiName)
	}

	reqURL := fmt.Sprintf("%s/%s",
		strings.TrimRight(api.URL(), "/"), // removes any trailing slash
		strings.TrimLeft(url, "/"))        // removes any leading slash

	req, err := http.NewRequest(method, reqURL, nil)
	if err != nil {
		return nil, err
	}

	// Go ahead and set a default now while we have all the data. The user can set a custom
	// value later using the request.SetUserAgent() method.
	req.Header.Set("User-Agent", fmt.Sprintf("kpurdon/apirequest (for %s)", r.apiName))

	return &Request{req}, nil
}

// Execute executes the given http.Request using the embedded http.Client and optionally decoding
// the result into a given non-nil data input.
// An error is only returned for un-handled errors. To handle an error pass a non-nil input to the
// errorData input that matches a known error response type. The first response boolean will
// indicate if the request was succesfull or not.
func (r *Requester) Execute(req *Request, successData, errorData interface{}) (bool, error) {
	resp, err := r.Do(req.Request)
	if err != nil {
		return false, err
	}

	if resp.StatusCode >= 400 {
		if errorData != nil {
			if err := json.NewDecoder(resp.Body).Decode(&errorData); err != nil {
				return false, err
			}
			return false, nil
		}
		return false, fmt.Errorf("%d:%s", resp.StatusCode, resp.Body)
	}

	if successData != nil {
		if err := json.NewDecoder(resp.Body).Decode(&successData); err != nil {
			return true, err
		}
	}

	return true, nil
}
