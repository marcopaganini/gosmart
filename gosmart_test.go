// This file is part of gosmart, a set of libraries to communicate with
// the Samsumg SmartThings API using Go (golang).
//
// http://github.com/marcopaganini/gosmart
// (C) 2016 by Marco Paganini <paganini@paganini.net>

package gosmart

import (
	"encoding/json"
	"fmt"
	"github.com/go-test/deep"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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

		if !checkerr(t, err, tt.wantError, tt) {
			continue
		}

		if uri != tt.wantURI {
			t.Errorf("Expected URI: %v, got %v (%+v)", dummyURI, uri, tt)
		}
	}
}

func TestSaveToken(t *testing.T) {
	const (
		testFilename = "/tmp/gosmart_test.data"
	)

	dummyToken := &oauth2.Token{
		AccessToken:  "access",
		TokenType:    "tokentype",
		RefreshToken: "refresh",
		Expiry:       time.Now().Add(time.Hour * 72),
	}

	caseTests := []struct {
		fname     string
		token     *oauth2.Token
		wantError bool
	}{
		// Basic test. Return OK.
		{
			fname: testFilename,
			token: dummyToken,
		},
		// Invalid Filename, Error.
		{
			fname:     "/tmp/non/existing/dir/test",
			token:     dummyToken,
			wantError: true,
		},
		// Invalid Token (nil), Error.
		{
			fname:     testFilename,
			wantError: true,
		},
		// Invalid Token (empty), Error.
		{
			fname:     testFilename,
			token:     &oauth2.Token{},
			wantError: true,
		},
	}

	for _, tt := range caseTests {
		err := SaveToken(tt.fname, tt.token)
		checkerr(t, err, tt.wantError, tt)
	}
}

func TestLoadToken(t *testing.T) {
	const (
		testFilename = "/tmp/gosmart_test.data"
	)

	dummyToken := &oauth2.Token{
		AccessToken:  "access",
		TokenType:    "tokentype",
		RefreshToken: "refresh",
		Expiry:       time.Now().Add(time.Hour * 72),
	}
	dummyTokenJSON, err := json.Marshal(dummyToken)
	if err != nil {
		t.Errorf("unable to generate dummy token: %v", err)
	}

	caseTests := []struct {
		fname     string
		saveToken []byte
		wantToken []byte
		wantError bool
	}{
		// Basic test. Return OK.
		{
			fname:     testFilename,
			saveToken: dummyTokenJSON,
			wantToken: dummyTokenJSON,
		},
		// Invalid Filename, Error.
		{
			fname:     "/tmp/non/existing/dir/test",
			saveToken: dummyTokenJSON,
			wantError: true,
		},
	}

	for _, tt := range caseTests {
		// Manually save token. We can't use SaveToken as it will refuse
		// to save invalid tokens.
		fname, err := makeTokenFile(testFilename)
		if !checkerr(t, err, false, "Error creating token file.") {
			continue
		}
		err = ioutil.WriteFile(fname, dummyTokenJSON, 0600)
		if !checkerr(t, err, false, tt) {
			continue
		}
		token, err := LoadToken(tt.fname)
		if !checkerr(t, err, tt.wantError, tt) {
			continue
		}
		if d := deep.Equal(dummyToken, token); d != nil {
			t.Error(d)
		}
	}
}

// checkerr checks common error return conditions (wantError true/false vs
// error true/false). Return false to signal the caller to stop processing of
// the current case (either an error happened or it makes no sense).  Return
// true otherwise.
func checkerr(t *testing.T, err error, wantError bool, v interface{}) bool {
	// If we want an error, always return false since processing can't
	// continue on the caller.
	if wantError {
		if err == nil {
			t.Errorf("expected error, got nil: %+v", v)
		}
		return false
	}
	// Don't want error, got error.
	if err != nil {
		t.Errorf("expected no error, got %v, %v", err, v)
		return false
	}
	// Don't want error, didn't get error.
	return true
}
