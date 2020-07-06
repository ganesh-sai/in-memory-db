package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/ganesh-sai/in-memory-db/db"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	srv := db.NewServer()
	select {
	case <-stop:
		srv.Stop()
	}
}
