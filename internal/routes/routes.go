package routes

import (
	"net/http"

	"github.com/michaeljsaenz/probe/internal/handlers"
)

func RegisterRoutes() {
	handlers.StaticFiles()
	http.HandleFunc("/", handlers.RootTemplate)
	http.HandleFunc("/istio/", handlers.IstioTemplate)
	http.HandleFunc("/kubernetes/", handlers.KubernetesTemplate)
	http.HandleFunc("/button-submit/", handlers.ButtonSubmit)
	http.HandleFunc("/button-certificates/", handlers.ButtonCertificates)
	http.HandleFunc("/button-dns/", handlers.ButtonDNS)
	http.HandleFunc("/button-ping/", handlers.ButtonPing)
	http.HandleFunc("/button-traceroute/", handlers.ButtonTraceroute)
}
