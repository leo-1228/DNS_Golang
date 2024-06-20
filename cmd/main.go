package main

import (
	"dnscheck"
	"flag"
	"log"
	"net/url"
)

func main() {
	mode := flag.String("mode", "client", "Mode to start application [client|server]")
	serverUrl := flag.String("server", "", "Url to access the api of the server, e.g. https://mydomain.au/api, or the serverUrl")
	secret := flag.String("secret", "", "Shared JWT secret")
	clientId := flag.String("clientId", "", "Static client id to match client specific settings")

	flag.Parse()

	var err error
	if *mode == "server" {
		dnscheck.Cfg, err = dnscheck.LoadConfig()
		if err != nil {
			log.Fatal(err)
		}
		if secret == nil || *secret == "" {
			log.Fatal("Missing secret flag to run in server mode")
		}

		dnscheck.StartServer(dnscheck.Cfg, *secret)
	} else {
		if serverUrl == nil || *serverUrl == "" {
			log.Fatal("Missing url flag to run in client mode")
		}
		if _, err := url.ParseRequestURI(*serverUrl); err != nil {
			log.Fatal("Invalid url flag to run in client mode")
		}

		if secret == nil || *secret == "" {
			log.Fatal("Missing secret flag to run in client mode")
		}

		dnscheck.StartClient(*serverUrl, *secret, *clientId)
	}

}
