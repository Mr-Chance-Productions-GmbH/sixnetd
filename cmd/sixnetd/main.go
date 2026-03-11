package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Mr-Chance-Productions-GmbH/sixnetd/internal/socket"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version":
			fmt.Println(version)
			return
		default:
			log.Fatalf("unknown flag: %s", os.Args[1])
		}
	}

	requireRoot()

	log.Println("sixnetd starting")

	srv := socket.NewServer(version)
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

func requireRoot() {
	if os.Getuid() != 0 {
		log.Fatal("sixnetd must run as root")
	}
}
