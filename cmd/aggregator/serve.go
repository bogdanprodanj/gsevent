package main

import (
	"fmt"
	"time"

	"github.com/gsevent/runtime"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/valyala/fasthttp"
)

// The serveCmd lets us serve our application.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "starts web server",
	Long:  "starts web server which provides endpoints for collecting, storing and viewing events",
	Run:   serve,
}

func init() {
	flags := serveCmd.Flags()
	flags.Int("http_port", 8080, "http port to listen to traffic on")
	flags.String("log_level", "debug", "log level")
	flags.String("service_name", "gsevent", "name of the application")
	flags.String("redis_address", "localhost:6379", "redis address")
	flags.String("redis_password", "", "redis password")
	flags.Duration("new_file_interval", time.Hour, "interval of creating new csv files for storing events")
	flags.Int("max_workers", 1000, "the maximum number of workers responsible for writing events into a file")
	rootCmd.AddCommand(serveCmd)
}

// serve initializes the server based on the configuration, and then starts
// listening to traffic.
func serve(cmd *cobra.Command, _ []string) {
	cfg := new(runtime.Config)
	err := runtime.NewConfig(cmd.Flags(), "config", cfg)
	if err != nil {
		log.Fatalf("serve: failed to load config %#v", err)
	}
	app, err := runtime.NewApp(cfg)
	if err != nil {
		log.Fatalf("serve: failed to create new app: %+v", err)
	}
	run(
		func() {
			log.Printf("serve: start listening on port %d", cfg.HTTPPort)
			app.Start()
			log.Fatal(fasthttp.ListenAndServe(fmt.Sprintf(":%d", cfg.HTTPPort), app.Router.Handler))
		},
		func() {
			app.Stop()
			log.Println("serve: stopping application")
		},
	)
}
