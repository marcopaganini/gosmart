// This file is part of gosmart, a set of libraries to communicate with
// the Samsumg SmartThings API using Go (golang).
//
// http://github.com/marcopaganini/gosmart
// (C) 2016 by Marco Paganini <paganini@paganini.net>

package gosmart

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestGetEndPointsURI(t *testing.T) {
	dummyURI := "http://localhost/response"
	dummyJSON := fmt.Sprintf(`[{"uri": "%s"}]`, dummyURI)

	caseTests := []struct {
		body                string
		wantURI             string
		useInvalidServerURL bool
		wantError           bool
	}{
		// Basic test. Valid URI and return. OK.
		{
			body:    dummyJSON,
			wantURI: dummyURI,
		},
		// Invalid URL (Error).
		{
			body:                dummyJSON,
			useInvalidServerURL: true,
			wantURI:             dummyURI,
			wantError:           true,
		},
		// Empty JSON content.
		{
			body:      "[]",
			wantURI:   dummyURI,
			wantError: true,
		},
		// Invalid JSON content.
		{
			body:      "InvalidJSON",
			wantURI:   dummyURI,
			wantError: true,
		},
	}

	for _, tt := range caseTests {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Write([]byte(tt.body))
		}))

		reqURL := server.URL
		if tt.useInvalidServerURL {
			reqURL = "InvalidURL"
		}
		uri, err := GetEndPointsURI(server.Client(), reqURL)
		server.Close()

		if tt.wantError {
			if err == nil {
				t.Errorf("Expected error, got nil")
			}
			continue
		}
		// We don't want errors from this point on.
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
			continue
		}

		if uri != tt.wantURI {
			t.Errorf("Expected URI: %v, got %v", dummyURI, uri)
		}
	}
}

// equals fails the test if exp is not equal to got.
func equals(tb testing.TB, exp, got interface{}) {
	if !reflect.DeepEqual(exp, got) {
		tb.Errorf("expected: %#v got: %#v\n\n", exp, got)
	}
}
