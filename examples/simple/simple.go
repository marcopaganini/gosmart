// Simple example for the gosmart libraries.
//
// This is a simple demonstration of how to obtain a token from the smartthings
// API using Oauth2 authorization, and how to request the status of some of your
// sensors (in this case, temperature).
//
// This file is part of gosmart, a set of libraries to communicate with
// the Samsumg SmartThings API using Go (golang).
//
// http://github.com/marcopaganini/gosmart
// (C) 2016 by Marco Paganini <paganini@paganini.net>

package main

import (
	"flag"
	"fmt"
	"github.com/marcopaganini/gosmart"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
)

const (
	tokenFilePrefix = ".example_st_token"
)

var (
	flagClient    = flag.String("client", "", "OAuth Client ID")
	flagSecret    = flag.String("secret", "", "OAuth Secret")
	flagTokenFile = flag.String("tokenfile", "", "Token filename")
)

func main() {
	flag.Parse()

	// No date on log messages
	log.SetFlags(0)

	// If we have a token file from the command line, use that directly.
	// Otherwise, form the name from tokenFilePrefix and the Client ID.
	tfile := *flagTokenFile
	if tfile == "" {
		if *flagClient == "" {
			log.Fatalf("Must specify Client ID (--client) or Token File (--tokenfile)")
		}
		tfile = tokenFilePrefix + "_" + *flagClient + ".json"
	}

	// Create the oauth2.config object and get a token
	config := gosmart.NewOAuthConfig(*flagClient, *flagSecret)
	token, err := gosmart.GetToken(tfile, config)
	if err != nil {
		log.Fatalln(err)
	}

	// Create a client with the token. This client will be used for all ST
	// API operations from here on.
	ctx := context.Background()
	client := config.Client(ctx, token)

	// Retrieve Endpoints URI. All future accesses to the smartthings API
	// for this session should use this URL, followed by the desired URL path.
	endpoint, err := gosmart.GetEndPointsURI(client)
	if err != nil {
		log.Fatalln(err)
		return
	}

	// Fetch temperature
	resp, err := client.Get(endpoint + "/temperature")
	if err != nil {
		log.Fatalln()
		return
	}
	contents, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("Temperature content: %s\n", contents)

	// Fetch batttery
	resp, err = client.Get(endpoint + "/battery")
	if err != nil {
		log.Fatalln()
		return
	}
	contents, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("Battery content: %s\n", contents)
}
