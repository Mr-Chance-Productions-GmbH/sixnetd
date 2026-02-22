package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if os.Getuid() != 0 {
		log.Fatal("sixnetd must run as root")
	}

	log.Println("sixnetd starting")

	// TODO: start socket server
	// TODO: start ZeroTier state poller

	// Wait for signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Println("sixnetd stopping")
}
