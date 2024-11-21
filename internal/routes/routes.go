package routes

import (
	"embed"
	"net/http"

	"github.com/michaeljsaenz/probe/internal/handlers"
)

func RegisterRoutes(fs embed.FS) {
	handlers.StaticFiles(fs)
	// template routes
	http.HandleFunc("/", handlers.RootTemplate)
	http.HandleFunc("/network/", handlers.NetworkTemplate)
	http.HandleFunc("/istio/", handlers.IstioTemplate)
	http.HandleFunc("/kubernetes/", handlers.KubernetesTemplate)
	// network routes
	http.HandleFunc("/button-submit/", handlers.ButtonSubmit)
	http.HandleFunc("/button-certificates/", handlers.ButtonCertificates)
	http.HandleFunc("/button-dns/", handlers.ButtonDNS)
	http.HandleFunc("/button-ping/", handlers.ButtonPing)
	http.HandleFunc("/button-traceroute/", handlers.ButtonTraceroute)
	// k8s routes
	http.HandleFunc("/button-get-pods/", handlers.ButtonGetPods)
	http.HandleFunc("/button-get-pod-detail/", handlers.ButtonPodDetail)
	http.HandleFunc("/click-get-pod-yaml/", handlers.ClickPodYaml)
	http.HandleFunc("/button-get-nodes/", handlers.ButtonGetNodes)
	http.HandleFunc("/button-get-node-conditions/", handlers.ButtonGetNodeConditions)
	http.HandleFunc("/button-get-namespaces/", handlers.ButtonGetNamespaces)
	http.HandleFunc("/dropdown-namespace-selection/", handlers.DropdownNamespaceSelection)
	http.HandleFunc("/clear-context-k8s-ns/", handlers.ClearContextK8sNamespace)
	http.HandleFunc("/click-container-log/", handlers.ClickContainerLog)
	http.HandleFunc("/click-container-port/", handlers.ClickContainerPort)
	http.HandleFunc("/click-container-exec/", handlers.ClickContainerExec)
	http.HandleFunc("/submit-container-exec/", handlers.SubmitContainerExec)
}
