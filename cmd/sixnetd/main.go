package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Mr-Chance-Productions-GmbH/sixnetd/internal/socket"
)

func main() {
	if os.Getuid() != 0 {
		log.Fatal("sixnetd must run as root")
	}

	log.Println("sixnetd starting")

	srv := socket.NewServer()
	if err := srv.Start(); err != nil {
		log.Fatalf("socket: %v", err)
	}
	defer srv.Stop()

	log.Printf("listening at %s", socket.SocketPath)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("sixnetd stopping")
}
