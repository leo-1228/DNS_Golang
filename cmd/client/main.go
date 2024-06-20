package main

import (
	"dnscheck"
	"flag"
	"log"
	"net/url"
)

func main() {

	serverUrl := flag.String("server", "", "Url to access the api of the server, e.g. https://mydomain.au/api, or the serverUrl")
	secret := flag.String("secret", "", "Shared secret with the server to generate JWT")
	clientId := flag.String("clientId", "", "Static clientId that should match the client settings in config.yml")

	flag.Parse()
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
