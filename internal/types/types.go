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
	PingResponses    []string
	TracerouteResult string
	Error            error
}

func NewApplication(application Application) Application {
	return Application{
		HttpServerHeader: application.HttpServerHeader,
		HttpRequestedUrl: application.HttpRequestedUrl,
		Certificates:     application.Certificates,
		DNS:              application.DNS,
		PingResponses:    application.PingResponses,
		Error:            application.Error,
		TracerouteResult: application.TracerouteResult,
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
