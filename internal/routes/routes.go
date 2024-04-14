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
	http.HandleFunc("/network-main/", handlers.NetworkMainTemplate)
	http.HandleFunc("/istio-main/", handlers.IstioMainTemplate)
	http.HandleFunc("/kubernetes-main/", handlers.KubernetesMainTemplate)
	// network routes
	http.HandleFunc("/button-submit/", handlers.ButtonSubmit)
	http.HandleFunc("/button-certificates/", handlers.ButtonCertificates)
	http.HandleFunc("/button-dns/", handlers.ButtonDNS)
	http.HandleFunc("/button-ping/", handlers.ButtonPing)
	http.HandleFunc("/button-traceroute/", handlers.ButtonTraceroute)
	// k8s routes
	http.HandleFunc("/button-get-pods/", handlers.ButtonGetPods)
	http.HandleFunc("/button-get-pod-detail/", handlers.ButtonPodDetail)
	http.HandleFunc("/button-get-pod-yaml/", handlers.ButtonPodYaml)
	http.HandleFunc("/button-get-nodes/", handlers.ButtonGetNodes)
	http.HandleFunc("/button-get-namespaces/", handlers.ButtonGetNamespaces)
	http.HandleFunc("/dropdown-namespace-selection/", handlers.DropdownNamespaceSelection)
}
