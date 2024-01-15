package utils

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
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
	url := fmt.Sprintf("http://localhost:%s", serverPort)

	err := exec.Command(cmd, url).Start()
	if err != nil {
		log.Printf("error opening browser: %v\n", err)
	}
}
