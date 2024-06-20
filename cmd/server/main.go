package main

import (
	"dnscheck"
	"log"
)

func main() {

	cfg, err := dnscheck.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	dnscheck.StartServer(cfg, cfg.JwtSecret)

}
