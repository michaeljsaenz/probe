package routes

import (
	"net/http"

	"github.com/michaeljsaenz/probe/internal/handlers"
)

func RegisterRoutes() {
	handlers.StaticFiles()
	http.HandleFunc("/", handlers.BaseUrl)
	http.HandleFunc("/submit-url/", handlers.SubmitUrl)
	http.HandleFunc("/button-certificates/", handlers.ButtonCertificates)
	http.HandleFunc("/button-dns/", handlers.ButtonDNS)
	http.HandleFunc("/button-ping/", handlers.ButtonPing)

}
