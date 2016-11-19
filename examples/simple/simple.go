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
	"log"
)

const (
	tokenFilePrefix = ".example_st_token"
)

var (
	flagClient    = flag.String("client", "", "OAuth Client ID")
	flagSecret    = flag.String("secret", "", "OAuth Secret")
	flagTokenFile = flag.String("tokenfile", "", "Token filename")
	flagDevID     = flag.String("devid", "", "Show information about this particular device ID")
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
	}

	// List devices or get specific info about one device
	if *flagDevID == "" {
		devs, err := gosmart.GetDevices(client, endpoint)
		if err != nil {
			log.Fatalln(err)
		}
		for _, d := range devs {
			fmt.Printf("ID: %s, Name: %q, Display Name: %q\n", d.ID, d.Name, d.DisplayName)
		}
	} else {
		dev, err := gosmart.GetDeviceInfo(client, endpoint, *flagDevID)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("Device ID:      %s\n", dev.ID)
		fmt.Printf("  Name:         %s\n", dev.Name)
		fmt.Printf("  Display Name: %s\n", dev.DisplayName)
		fmt.Printf("Attributes:\n")
		for k, v := range dev.Attributes {
			fmt.Printf("  %v: %v\n", k, v)
		}
	}
}
