package main

import (
	"github.com/asnelzin/stream-api/pkg/rest"
	"github.com/asnelzin/stream-api/pkg/store/inmem"
	"github.com/jessevdk/go-flags"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var revision = "unknown"

type Opts struct {
	Port        int `long:"port" env:"API_PORT" default:"8080" description:"port"`
	FinishAfter int `long:"finish-after" env:"FINISH_AFTER" default:"10" description:"finish interrupted stream after (in seconds)"`
}

func main() {
	var opts Opts
	p := flags.NewParser(&opts, flags.Default)
	if _, e := p.ParseArgs(os.Args[1:]); e != nil {
		os.Exit(1)
	}

	log.Print("[INFO] started stream-api")

	dataStore := inmem.NewStore(opts.FinishAfter)

	server := rest.Server{
		Version:   revision,
		DataStore: dataStore,
	}

	go func() { // catch signal and invoke graceful termination
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Print("[WARN] interrupt signal")
		server.Shutdown()
	}()

	server.Run(opts.Port)
}
