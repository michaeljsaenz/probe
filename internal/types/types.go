package types

import (
	"context"
	"sync"
)

type CustomContextValues struct {
	URL      string
	Hostname string
}

type CustomContextKey string

const ContextKey CustomContextKey = "ContextKey"

type Application struct {
	HttpServerHeader string
	HttpRequestedUrl string
	Certificates     string
	DNS              []string
	Error            error
}

func NewApplication(httpServerHeader, httpRequestedUrl, certificates string, dns []string, err error) Application {
	return Application{
		HttpServerHeader: httpServerHeader,
		HttpRequestedUrl: httpRequestedUrl,
		Certificates:     certificates,
		DNS:              dns,
		Error:            err,
	}
}

var SharedContext context.Context = context.Background()
var ContextLock sync.Mutex

func UpdateSharedContext(requestedUrlHeader, requestHostnameHeader string) {
	ContextLock.Lock()
	defer ContextLock.Unlock()
	SharedContext = context.WithValue(SharedContext, ContextKey, CustomContextValues{
		URL:      requestedUrlHeader,
		Hostname: requestHostnameHeader,
	})
}
