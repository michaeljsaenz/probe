package types

import (
	"context"
	"embed"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type PodDetail struct {
	PodName       string
	PodNamespace  string
	PodStatus     string
	PodAge        string
	PodNode       string
	PodContainers map[string][]int32
}

type K8sContainerDetail struct {
	ContainerName         string
	PodName               string
	PodNamespace          string
	ContainerExecResponse string
}

type K8sNode struct {
	Name           string
	Status         string
	Age            string
	NodeConditions []K8sNodeCondition
}

type K8sPodPortForward struct {
	PodPort string
	URL     string
}

type K8sPod struct {
	Name      string
	Namespace string
	Status    string
}

type K8sNodeCondition struct {
	Type    string
	Status  string
	Reason  string
	Message string
}

type K8sNodesDetail struct {
	TotalCount int
}

type CustomContextValues struct {
	URL      string
	Hostname string
}

type CustomContextValuesFS struct {
	HttpFS embed.FS
}

type CustomContextValuesK8s struct {
	Clientset *kubernetes.Clientset
	Config    *rest.Config
	Namespace string
	Pod       string
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
	K8sPods              []K8sPod
	K8sPodDetail         PodDetail
	K8sNode              K8sNode
	K8sNodes             []K8sNode
	K8sNodesDetail       K8sNodesDetail
	K8sPodYaml           string
	K8sPodPortForward    K8sPodPortForward
	K8sPodLog            string
	K8sContainerDetail   K8sContainerDetail
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
		K8sPodDetail:         application.K8sPodDetail,
		K8sNode:              application.K8sNode,
		K8sNodes:             application.K8sNodes,
		K8sNodesDetail:       application.K8sNodesDetail,
		K8sPodYaml:           application.K8sPodYaml,
		K8sPodPortForward:    application.K8sPodPortForward,
		K8sPodLog:            application.K8sPodLog,
		K8sContainerDetail:   application.K8sContainerDetail,
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

func UpdateSharedContextK8s(clienset *kubernetes.Clientset, config *rest.Config, namespace, pod string) {
	ContextLockK8s.Lock()
	defer ContextLockK8s.Unlock()
	SharedContextK8s = context.WithValue(SharedContextK8s, ContextKey, CustomContextValuesK8s{
		Clientset: clienset,
		Config:    config,
		Namespace: namespace,
		Pod:       pod,
	})
}
