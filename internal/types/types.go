package types

import (
	"context"
	"embed"
	"sync"

	"k8s.io/client-go/kubernetes"
)

type CustomContextValues struct {
	URL      string
	Hostname string
}

type CustomContextValuesFS struct {
	HttpFS embed.FS
}

type CustomContextValuesK8s struct {
	Clientset *kubernetes.Clientset
	Namespace string
}

type CustomContextKey string

const ContextKey CustomContextKey = "ContextKey"

type Application struct {
	HttpServerHeader     string
	HttpRequestedUrl     string
	HttpResponseStatus   string
	Certificates         string
	DNS                  []string
	PingResponses        []string
	TracerouteResult     string
	Error                error
	K8sNamespaces        []string
	K8sSelectedNamespace string
	K8sPods              []string
}

func NewApplication(application Application) Application {
	return Application{
		HttpServerHeader:     application.HttpServerHeader,
		HttpRequestedUrl:     application.HttpRequestedUrl,
		HttpResponseStatus:   application.HttpResponseStatus,
		Certificates:         application.Certificates,
		DNS:                  application.DNS,
		PingResponses:        application.PingResponses,
		Error:                application.Error,
		TracerouteResult:     application.TracerouteResult,
		K8sNamespaces:        application.K8sNamespaces,
		K8sSelectedNamespace: application.K8sSelectedNamespace,
		K8sPods:              application.K8sPods,
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

var SharedContextFS context.Context = context.Background()
var ContextLockFS sync.Mutex

func UpdateSharedContextFS(httpFS embed.FS) {
	ContextLockFS.Lock()
	defer ContextLockFS.Unlock()
	SharedContextFS = context.WithValue(SharedContextFS, ContextKey, CustomContextValuesFS{
		HttpFS: httpFS,
	})
}

var SharedContextK8s context.Context = context.Background()
var ContextLockK8s sync.Mutex

func UpdateSharedContextK8s(clienset *kubernetes.Clientset, namespace string) {
	ContextLockFS.Lock()
	defer ContextLockFS.Unlock()
	SharedContextK8s = context.WithValue(SharedContextK8s, ContextKey, CustomContextValuesK8s{
		Clientset: clienset,
		Namespace: namespace,
	})
}
