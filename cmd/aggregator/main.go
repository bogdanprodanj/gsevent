package main

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// rootCmd is the root command for the app.
var rootCmd = &cobra.Command{
	Use:   "gsevent",
	Short: "Provides an API for collecting, storing and viewing a variety of events",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err.Error())
		os.Exit(2)
	}
}

func run(start, stop func()) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	complete := make(chan bool, 1)
	go func() {
		start()
		complete <- true
	}()
	log.Println("press Ctrl-C to shutdown")
	select {
	case <-signals:
		stop()
	case <-complete:
	}
}
