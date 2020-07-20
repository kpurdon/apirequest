package apirequest

// MockClient implements the Requester interface offering complete control over the outputs and tracking calls.
type MockClient struct {
	MustAddAPIFn     func(apiName string, discoverer Discoverer)
	MustAddAPICalled bool

	NewRequestFn     func(apiName, method, url string) (*Request, error)
	NewRequestCalled bool

	ExecuteFn       func(req *Request, successData, errorData interface{}) (bool, error)
	ExecuteFnCalled bool
}

// MustAddAPI implements the Requester.MustAddAPI method.
func (c *MockClient) MustAddAPI(apiName string, discoverer Discoverer) {
	c.MustAddAPICalled = true
	c.MustAddAPIFn(apiName, discoverer)
}

// NewRequest implements the Requester.NewRequest method.
func (c *MockClient) NewRequest(apiName, method, url string) (*Request, error) {
	c.NewRequestCalled = true
	return c.NewRequestFn(apiName, method, url)
}

// Execute implements the Requester.Execute method.
func (c *MockClient) Execute(req *Request, successData, errorData interface{}) (bool, error) {
	c.ExecuteFnCalled = true
	return c.ExecuteFn(req, successData, errorData)
}
