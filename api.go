// This file is part of gosmart, a set of libraries to communicate with
// the Samsumg SmartThings API using Go (golang).
//
// http://github.com/marcopaganini/gosmart
// (C) 2016 by Marco Paganini <paganini@paganini.net>

package gosmart

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// DeviceList holds the list of devices returned by /devices
type DeviceList struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

// DeviceInfo holds information about a specific device.
type DeviceInfo struct {
	DeviceList
	Attributes map[string]interface{} `json:"attributes"`
}

// DeviceCommand holds one command a device can accept.
type DeviceCommand struct {
	Command string                 `json:"command"`
	Params  map[string]interface{} `json:"params"`
}

// GetDevices returns the list of devices from smartthings using
// the specified http.client and endpoint URI.
func GetDevices(client *http.Client, endpoint string) ([]DeviceList, error) {
	ret := []DeviceList{}

	uri := endpoint + "/devices"
	resp, err := client.Get(uri)
	if err != nil {
		return nil, err
	}
	contents, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(contents, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// GetDeviceInfo returns device specific information about a particular device.
func GetDeviceInfo(client *http.Client, endpoint string, id string) (*DeviceInfo, error) {
	ret := &DeviceInfo{}

	uri := endpoint + "/devices/" + id
	resp, err := client.Get(uri)
	if err != nil {
		return nil, err
	}
	contents, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(contents, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

// GetDeviceCommands returns a slice of commands a specific device accepts.
func GetDeviceCommands(client *http.Client, endpoint string, id string) ([]DeviceCommand, error) {
	ret := []DeviceCommand{}

	uri := endpoint + "/devices/" + id + "/commands"
	resp, err := client.Get(uri)
	if err != nil {
		return nil, err
	}
	contents, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(contents, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}
