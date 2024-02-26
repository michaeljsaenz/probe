package main

import (
	"embed"
	"log"
	"net/http"

	"github.com/michaeljsaenz/probe/internal/routes"
	"github.com/michaeljsaenz/probe/internal/utils"
)

//go:embed static/* templates/*
var fs embed.FS

func main() {
	routes.RegisterRoutes(fs)

	serverPort := "8099"
	server := http.Server{Addr: ":" + serverPort}

	go func() {
		log.Fatal(server.ListenAndServe())
	}()
	log.Print("server is up.")
	defer server.Close()

	utils.OpenBrowser(serverPort)
	utils.Shutdown(&server)
}
