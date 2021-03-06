REPLACED BY https://github.com/kpurdon/apir

[![CircleCI](https://circleci.com/gh/kpurdon/apirequest.svg?style=svg)](https://circleci.com/gh/kpurdon/apirequest)
[![codecov status](https://codecov.io/gh/kpurdon/apirequest/branch/master/graph/badge.svg)](https://codecov.io/gh/kpurdon/apirequest)
[![godoc](https://godoc.org/github.com/kpurdon/apirequest?status.svg)](http://godoc.org/github.com/kpurdon/apirequest)
[![Go Report Card](https://goreportcard.com/badge/github.com/kpurdon/apirequest)](https://goreportcard.com/report/github.com/kpurdon/apirequest)

apirequest
-----

A simple helper for making requests to HTTP APIs that return JSON.

## Examples

### Initialize a New Client

The first step is to initialize a [client](https://godoc.org/github.com/kpurdon/apirequest#Client). Next we add any number of APIs to the requester using a [discoverer](https://godoc.org/github.com/kpurdon/apirequest#Discoverer). There are some pre-defined discoverers in the [/discoverers](https://godoc.org/github.com/kpurdon/apirequest/discoverers) directory.

``` go
client := apirequest.Client("thisapi", nil)
client.MustAddAPI("anotherapi", direct.NewDiscoverer("http://127.0.0.1:1234"))
```

Ideally the `client` should be injected (use `apirequest.Requester` as the type) into whatever methods need to make requests to the services we have registered instead of being a globally defined resource.

### Make a Request

Note: This example code makes use of [kpurdon/apiresponse](https://github.com/kpurdon/apiresponse).

TODO: add more notes here in the code ... or examples.

``` go
req, err := client.NewRequest("anotherapi", http.MethodGet, "/data")
if err != nil {
    log.Printf("%+v", err)
    responder.InternalServerError()
    return
}

var (
    data    Data
    errData apiresponse.GenericError
)
ok, err := client.Execute(req, &data, &errData)
if err != nil {
    log.Printf("%+v", err)
    responder.InternalServerError()
    return
}
if !ok {
    responder.WithData(errData)
    responder.InternalServerError()
    return
}

responder.WithData(data)
responder.OK()
```
