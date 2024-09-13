package main

import (
	"embed"
	"log"

	"github.com/michaeljsaenz/probe/internal/k8s"
	"github.com/michaeljsaenz/probe/internal/routes"
	"github.com/michaeljsaenz/probe/internal/utils"
)

//go:embed static/* templates/* templates/*/*
var fs embed.FS

func main() {
	// Initialize the client for the first time
	k8s.GetClientSet()

	go k8s.RefreshClientSet()

	routes.RegisterRoutes(fs)

	server, listener := utils.ServerPortAndListener()

	go func() {
		log.Fatal(server.Serve(listener))
	}()
	log.Print("Server is up.")
	defer server.Close()

	utils.OpenBrowser(server.Addr)

	utils.Shutdown(&server)
}
