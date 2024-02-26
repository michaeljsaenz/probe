package routes

import (
	"embed"
	"net/http"

	"github.com/michaeljsaenz/probe/internal/handlers"
)

func RegisterRoutes(fs embed.FS) {
	handlers.StaticFiles(fs)
	http.HandleFunc("/", handlers.RootTemplate)
	http.HandleFunc("/network-main/", handlers.NetworkMainTemplate)
	http.HandleFunc("/istio-main/", handlers.IstioMainTemplate)
	http.HandleFunc("/kubernetes-main/", handlers.KubernetesMainTemplate)
	http.HandleFunc("/button-submit/", handlers.ButtonSubmit)
	http.HandleFunc("/button-certificates/", handlers.ButtonCertificates)
	http.HandleFunc("/button-dns/", handlers.ButtonDNS)
	http.HandleFunc("/button-ping/", handlers.ButtonPing)
	http.HandleFunc("/button-traceroute/", handlers.ButtonTraceroute)
	http.HandleFunc("/dropdown-namespace-selection/", handlers.DropdownNamespaceSelection)
	http.HandleFunc("/button-get-pods/", handlers.ButtonGetPods)
}
