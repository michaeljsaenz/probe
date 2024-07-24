package utils

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

func Shutdown(server *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)

	// block until we receive our signal.
	<-interruptChan

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	log.Print("server is shutting down...")
	defer cancel()
	server.Shutdown(ctx)
}

func OpenBrowser(serverPort string) {
	var cmd string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "start"
	default:
		cmd = "xdg-open"
	}
	url := fmt.Sprintf("http://localhost%s", serverPort)

	err := exec.Command(cmd, url).Start()
	if err != nil {
		log.Printf("error opening browser: %v\n", err)
	}
}

func ServerPortAndListener() (server http.Server, listener net.Listener) {
	startingServerPort := 8099
	var err error
	if port, exists := os.LookupEnv("PROBE_SERVER_PORT"); exists {
		port, err := strconv.Atoi(port)
		if err == nil {
			startingServerPort = port
		}
	}

	for {
		serverPort := strconv.Itoa(startingServerPort)
		server = http.Server{Addr: ":" + serverPort}

		listener, err = net.Listen("tcp", server.Addr)
		if err != nil {
			log.Printf("Port %s in use, trying next port...\n", serverPort)
			startingServerPort++
			continue
		}

		log.Printf("Server starting on port %s\n", serverPort)
		return
	}
}

func FindLocalPort(startingLocalPort int) (localPort string, err error) {
	startingPoint := startingLocalPort
	var upperLimit int = 10000
	var listener net.Listener
	for {
		if startingLocalPort == upperLimit {
			return "", fmt.Errorf("no available local port in the range %d-%d for use", startingPoint, upperLimit-1)
		}
		localPort = strconv.Itoa(startingLocalPort)

		listener, err = net.Listen("tcp", fmt.Sprintf("127.0.0.1:%s", localPort))
		if err != nil {
			log.Printf("Local port %s in use, trying next port...\n", localPort)
			startingLocalPort++
			continue
		}
		_ = listener.Close()
		log.Printf("Port avaialble on %s\n", localPort)
		return
	}

}
