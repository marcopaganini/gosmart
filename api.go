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

// GetDevices returns the list of devices from smartthings using
// the specified http.client and endpoint URI.
func GetDevices(client *http.Client, endpoint string) ([]DeviceList, error) {
	// For some reason, the /devices handler returns an array of one element,
	// containing another array with the list of devices.  We avoid the extra
	// confusion by returning only the first element in the slice.
	ret := [][]DeviceList{}

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
	return ret[0], nil
}
