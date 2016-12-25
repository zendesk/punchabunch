package main

import (
	"flag"
	"log"

	punchabunch "github.com/zendesk/punchabunch/lib"
)

func main() {
	var config punchabunch.Config

	configFile := flag.String("c", "config.toml", "Path to configuration file")
	verbose := flag.Bool("v", false, "Log verbosely")

	flag.Parse()
	config, err := punchabunch.ParseConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	ready := make(chan bool)

	config.Verbose = *verbose
	config.Ready = ready

	punchabunch.Run(config)
	<-ready

	log.Println("Connections ready")
	select {}
}
