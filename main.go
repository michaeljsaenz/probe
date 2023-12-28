package main

import (
	"log"
	"net/http"

	"github.com/michaeljsaenz/probe/internal/routes"
)

func main() {
	routes.RegisterRoutes()

	server := http.Server{Addr: ":8001"}
	log.Fatal(server.ListenAndServe())

	defer server.Close()
}
