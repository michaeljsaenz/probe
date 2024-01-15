package main

import (
	"log"
	"net/http"

	"github.com/michaeljsaenz/probe/internal/routes"
	"github.com/michaeljsaenz/probe/internal/utils"
)

func main() {
	routes.RegisterRoutes()

	serverPort := "8082"
	server := http.Server{Addr: ":" + serverPort}

	go func() {
		log.Fatal(server.ListenAndServe())
	}()
	log.Print("server is up.")
	defer server.Close()

	utils.OpenBrowser(serverPort)
	utils.Shutdown(&server)
}
