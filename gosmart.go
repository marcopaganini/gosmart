// This file is part of gosmart, a set of libraries to communicate with
// the Samsumg SmartThings API using Go (golang).
//
// http://github.com/marcopaganini/gosmart
// (C) 2016 by Marco Paganini <paganini@paganini.net>

package gosmart

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"os/user"
	"path/filepath"
	"strconv"
)

const (
	authDone  = "<html><body>Authentication Completed.</body></html>"
	authError = "<html><body>AUthentication error. Please see terminal output for details.</body></html>"

	// Endpoints URL
	endPointsURI = "https://graph.api.smartthings.com/api/smartapps/endpoints"

	// URL paths used for Oauth authentication on localhost
	callbackPath = "/OAuthCallback"
	donePath     = "/OauthDone"
	rootPath     = "/"

	// Token save file
	defaultTokenFile = ".st_token.json"

	// default local HTTP server port
	defaultPort = 4567
)

// Auth contains the SmartThings authentication related data.
type Auth struct {
	port             int
	config           *oauth2.Config
	rchan            chan oauthReturn
	oauthStateString string
}

// oauthReturn contains the values returned by the OAuth callback handler.
type oauthReturn struct {
	token *oauth2.Token
	err   error
}

// endpoints holds the values returned by the SmartThings endpoints URI.
type endpoints struct {
	OauthClient struct {
		ClientID string `json:"clientId"`
	}
	Location struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	URI     string `json:"uri"`
	BaseURL string `json:"base_url"`
	URL     string `json:"url"`
}

// NewOAuthConfig creates a new oauth2.config structure with the
// correct parameters to use smartthings.
func NewOAuthConfig(client, secret string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     client,
		ClientSecret: secret,
		Scopes:       []string{"app"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://graph.api.smartthings.com/oauth/authorize",
			TokenURL: "https://graph.api.smartthings.com/oauth/token",
		},
	}
}

// NewAuth creates a new Auth struct
func NewAuth(port int, config *oauth2.Config) (*Auth, error) {
	rnd, err := randomString(16)
	if err != nil {
		return nil, err
	}

	return &Auth{
		port:             port,
		config:           config,
		rchan:            make(chan oauthReturn),
		oauthStateString: rnd,
	}, nil
}

// FetchOAuthToken sets up the handler and a local HTTP server and fetches an
// Oauth token from the smartthings website.
func (g *Auth) FetchOAuthToken() (*oauth2.Token, error) {
	http.HandleFunc(rootPath, g.handleMain)
	http.HandleFunc(donePath, g.handleDone)
	http.HandleFunc(callbackPath, g.handleOAuthCallback)

	go http.ListenAndServe(":"+strconv.Itoa(g.port), nil)

	// Block on the return channel (this is set by handleOauthCallback)
	ret := <-g.rchan
	return ret.token, ret.err
}

// handleMain redirects the user to the main authentication page.
func (g *Auth) handleMain(w http.ResponseWriter, r *http.Request) {
	url := g.config.AuthCodeURL(g.oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// handleError shows a page indicating the authentication has failed.
func (g *Auth) handleError(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, authError)
}

// handleDone shows a page indicating the authentication is finished.
func (g *Auth) handleDone(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, authDone)
}

// handleOauthCallback fetches the callback from the OAuth provider and parses
// the URL, extracting the code and then exchanging it for a token.
func (g *Auth) handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	// Make sure we have the same "state" as our request.
	state := r.FormValue("state")
	if state != g.oauthStateString {
		g.rchan <- oauthReturn{
			token: nil,
			err:   fmt.Errorf("invalid oauth state, expected %q, got %q", g.oauthStateString, state),
		}
		return
	}

	// Retrieve the code from the URL, and exchange for a token
	code := r.FormValue("code")
	token, err := g.config.Exchange(oauth2.NoContext, code)
	if err != nil {
		g.rchan <- oauthReturn{
			token: nil,
			err:   fmt.Errorf("code exchange failed: %q", err),
		}
		return
	}

	// Return token.
	g.rchan <- oauthReturn{
		token: token,
		err:   nil,
	}
	// Redirect user to "Authentication done" page
	http.Redirect(w, r, donePath, http.StatusTemporaryRedirect)
	return
}

// GetEndPointsURI returns the smartthing endpoints URI. The endpoints
// URI is the base for all app requests.
func GetEndPointsURI(client *http.Client) (string, error) {
	// Fetch the JSON containing our endpoint URI
	resp, err := client.Get(endPointsURI)
	if err != nil {
		return "", fmt.Errorf("error getting endpoints URI %q", err)
	}
	contents, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if string(contents) == "[]" {
		return "", fmt.Errorf("endpoint URI returned no content")
	}

	// Only URI is fetched from JSON string.
	var ep []endpoints
	err = json.Unmarshal(contents, &ep)
	if err != nil {
		return "", fmt.Errorf("error decoding JSON: %q", err)
	}
	return ep[0].URI, nil
}

// LoadToken loads the token from a file on disk. If nil is used for filename
// a default filename user the user's directory is used.
func LoadToken(fname string) (*oauth2.Token, error) {
	// Generate token filename
	fname, err := makeTokenFile(fname)
	if err != nil {
		return nil, err
	}

	// Read & Decode JSON
	blob, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	token := &oauth2.Token{}
	if err := json.Unmarshal(blob, token); err != nil {
		return nil, err
	}

	return token, nil
}

// SaveToken saves the token to a file on disk. If nil is used for filename
// a default filename user the user's directory is used.
func SaveToken(fname string, token *oauth2.Token) error {
	// Generate token filename
	fname, err := makeTokenFile(fname)
	if err != nil {
		return err
	}

	// Encode & Save
	blob, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fname, blob, 0600)
}

// randomString generates a random string of bytes of the specified size
// and returns the its hexascii representation.
func randomString(size int) (string, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}

// GetToken returns the token for the ClientID and Secret specified in config.
// The function attempts to load the token from tokenFile first, and failing
// that, starts a full token authentication cycle with SmartThings. If
// tokenFile is blank, the function uses a default name under the current
// user's home directory. The token is saved to local disk before being
// returned to the caller.
//
// This function represents the most common (and possibly convenient) way to
// retrieve a token for a given ClientID and Secret.
func GetToken(tokenFile string, config *oauth2.Config) (*oauth2.Token, error) {
	// Attempt to load token from local storage. Fallback to full auth cycle.
	token, err := LoadToken(tokenFile)
	if err != nil || !token.Valid() {
		if config.ClientID == "" || config.ClientSecret == "" {
			return nil, errors.New("Need ClientID and Secret to generate new Token")
		}
		gst, err := NewAuth(defaultPort, config)
		if err != nil {
			return nil, err
		}

		fmt.Printf("Please login by visiting http://localhost:%d\n", defaultPort)
		token, err = gst.FetchOAuthToken()
		if err != nil {
			return nil, err
		}

		// Once we have the token, save it locally for future use.
		err = SaveToken(tokenFile, token)
		if err != nil {
			return nil, err
		}
	}
	return token, nil
}

// tokenFile generates a filename to store the token.
func makeTokenFile(fname string) (string, error) {
	// If filename is an absolute path, return it as is.
	// If filename != "", return user_home/filename
	// Otherwise, return user_home/defaultTokenFile

	if filepath.IsAbs(fname) {
		return fname, nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	if fname != "" {
		return filepath.Join(usr.HomeDir, fname), nil
	}
	return filepath.Join(usr.HomeDir, defaultTokenFile), nil
}
