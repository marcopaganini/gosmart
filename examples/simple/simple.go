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
	flagAll       = flag.Bool("all", false, "Show Information about all devices found")
)

func main() {
	flag.Parse()

	// No date on log messages
	log.SetFlags(0)

	if *flagDevID != "" && *flagAll {
		log.Fatalln("Invalid flag combination: --devid and --all are mutually exclusive.")
	}

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
	endpoint, err := gosmart.GetEndPointsURI(client, gosmart.EndPointsURI)
	if err != nil {
		log.Fatalln(err)
	}

	devices := []string{}

	if *flagDevID != "" {
		devices = append(devices, *flagDevID)
	}
	// List all info about devices if --all specified
	if *flagAll {
		devs, err := gosmart.GetDevices(client, endpoint)
		if err != nil {
			log.Fatalln(err)
		}
		for _, d := range devs {
			devices = append(devices, d.ID)
		}
	}

	if len(devices) == 0 {
		devs, err := gosmart.GetDevices(client, endpoint)
		if err != nil {
			log.Fatalln(err)
		}
		for _, d := range devs {
			fmt.Printf("ID: %s, Name: %q, Display Name: %q\n", d.ID, d.Name, d.DisplayName)
		}
	} else {
		for _, id := range devices {
			dev, err := gosmart.GetDeviceInfo(client, endpoint, id)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Printf("\nDevice ID:      %s\n", dev.ID)
			fmt.Printf("  Name:         %s\n", dev.Name)
			fmt.Printf("  Display Name: %s\n", dev.DisplayName)
			fmt.Printf("  Attributes:\n")
			for k, v := range dev.Attributes {
				fmt.Printf("    %v: %v\n", k, v)
			}

			fmt.Printf("  Commands & Parameters:\n")
			cmds, err := gosmart.GetDeviceCommands(client, endpoint, id)
			for _, cmd := range cmds {
				fmt.Printf("    %s", cmd.Command)
				if len(cmd.Params) != 0 {
					fmt.Printf(" Parameters:")
					for k, v := range cmd.Params {
						fmt.Printf(" %s=%s", k, v)
					}
				}
				fmt.Println()
			}
		}
	}
}
