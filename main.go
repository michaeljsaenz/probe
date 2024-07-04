package main

import (
	"embed"
	"log"
	"net/http"
	"os"

	"github.com/michaeljsaenz/probe/internal/k8s"
	"github.com/michaeljsaenz/probe/internal/routes"
	"github.com/michaeljsaenz/probe/internal/utils"
)

//go:embed static/* templates/*
var fs embed.FS

func main() {
	// Initialize the client for the first time
	k8s.GetClientSet()

	go k8s.RefreshClientSet()

	routes.RegisterRoutes(fs)

	serverPort := "8099"
	if port, exists := os.LookupEnv("PROBE_SERVER_PORT"); exists {
		serverPort = port
	}
	server := http.Server{Addr: ":" + serverPort}

	go func() {
		log.Fatal(server.ListenAndServe())
	}()
	log.Print("server is up.")
	defer server.Close()

	utils.OpenBrowser(serverPort)
	utils.Shutdown(&server)
}
