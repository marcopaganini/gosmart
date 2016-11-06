// Example of how to upload sensor data from SmartThings sensors to ThingSpeak
// using the GoSmart libraries.
//
// This file is part of gosmart, a set of libraries to communicate with
// the Samsumg SmartThings API using Go (golang).
//
// http://github.com/marcopaganini/gosmart
// (C) 2016 by Marco Paganini <paganini@paganini.net>

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/marcopaganini/gosmart"
	"golang.org/x/net/context"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	tokenFile         = ".smartthings-thingspeak.json"
	thingSpeakBaseURL = "https://api.thingspeak.com/update?api_key="
)

// TempCapability represents the information returned by the
// Temperature Capability in SmartThings.
type TempCapability struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// tsChannelMap maps a SmartThings sensor name to a ThingSpeak field ID.
type tsChannelMap map[string]int

var (
	flagClient = flag.String("client", "", "OAuth Client ID")
	flagSecret = flag.String("secret", "", "OAuth Secret")
	flagAPIKey = flag.String("apikey", "", "ThingSpeak write API key")

	// tscmap maps the sensor names to the ThingSpeak channel field numbers.
	// All SmartThings temperature capable sensors must be added here. The
	// values here are just examples.
	tscmap = tsChannelMap{
		"Front Door Sensor":           1,
		"Garage Door Sensor":          2,
		"Laundry Door Sensor":         3,
		"Upper Hallway Motion Sensor": 4,
	}
)

func main() {
	flag.Parse()

	// No date on log messages
	log.SetFlags(0)

	// Command line processing
	if *flagAPIKey == "" {
		log.Fatalln("Need ThingSpeak write API key (--apikey)")
	}

	// Retrieve token
	config := gosmart.NewOAuthConfig(*flagClient, *flagSecret)
	token, err := gosmart.GetToken(tokenFile, config)
	if err != nil {
		log.Fatalln(err)
	}

	// Create a client with token.
	ctx := context.Background()
	client := config.Client(ctx, token)

	// Retrieve Endpoints URI.
	endpoint, err := gosmart.GetEndPointsURI(client)
	if err != nil {
		log.Fatalln(err)
		return
	}
	temps, err := fetchTemperature(client, endpoint)
	if err != nil {
		log.Fatalln(err)
	}
	if err = updateThingSpeak(tscmap, temps, *flagAPIKey); err != nil {
		log.Fatalln(err)
	}
}

// fetchTemperature retrieves the temperature from all sensors in SmartThings.
func fetchTemperature(client *http.Client, endpoint string) ([]TempCapability, error) {
	// Fetch temperature from ST
	resp, err := client.Get(endpoint + "/temperature")
	if err != nil {
		return nil, err
	}

	contents, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	// Convert to JSON
	var temps []TempCapability
	err = json.Unmarshal(contents, &temps)
	if err != nil {
		return nil, fmt.Errorf("error decoding JSON: %q", err)
	}
	return temps, nil
}

// updateThingSpeak updates a thingspeak channel with the relevant data.
func updateThingSpeak(tscmap tsChannelMap, temps []TempCapability, apikey string) error {
	// Thingspeak uses fieldN fieldnames in their channels.  We use
	// tsChannelMap to retrieve the correspondence between sensor name and
	// ThingSpeak channel field number.
	req := ""
	for _, t := range temps {
		fieldno, ok := tscmap[t.Name]
		if !ok {
			log.Printf("Unable to find ThingSpeak field for %q", t.Name)
			continue
		}
		req += fmt.Sprintf("&field%d=%d", fieldno, t.Value)
	}
	// Make request
	url := thingSpeakBaseURL + apikey + req
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check for application level errors
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("Got HTTP return code %d for %q\n", resp.StatusCode, url)
	}
	return nil
}
