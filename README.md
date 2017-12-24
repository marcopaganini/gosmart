# gosmart
[![Go Report Card](https://goreportcard.com/badge/github.com/marcopaganini/gosmart)](https://goreportcard.com/report/github.com/marcopaganini/gosmart)
[![GoDoc](https://godoc.org/github.com/marcopaganini/gosmart?status.svg)](https://godoc.org/github.com/marcopaganini/gosmart)

A Go (golang) library to interface with the Samsung SmartThings (IoT) API.

## Introduction

GoSmart (or gosmart, lowercase) is a Go Library that allows easy interfacing
between a Go program and the SmartThings API. I've written this library out of
frustration with the many incomplete examples of how to properly authenticate
using Oauth2 in Go, and the inability to read the wealth of data reported by my
home sensors without dealing with Groovy and a mobile device.

The library is under heavy development, but can already perform an OAuth login
to SmartThings, save and restore the token locally, and fetch device names and
sensor values. Looking at the `examples` directory should provide a good
starting point on how to use it.

The library contains two parts: The Go library itself, and a Groovy "handler"
app. All calls to the SmartThings API are made with an authenticated HTTP
request to the SmartThings API website. The request is then served by the
Groovy app and the results converted back from JSON to a native Go structure
for regular use.

## Installation

### Local

We assume a working Go environment. This includes a recent version of Go (currently
using 1.7.3, check with `go version`) and properly set GOPATH/GOROOT environment
variables.

First, grab the latest version of gosmart with:

    go get -u -v github.com/marcopaganini/gosmart

This will install the package under `$GOPATH/src/github.com/marcopaganini/gosmart`

Change to that directory and locate the `endpoint.groovy` file under the `apismartapp`
directory. We'll need this file momentarily.

### SmartThings API setup

This only needs to be done *once*.

* Navigate to the [SmartThings API website](https://graph.api.smartthings.com/). Register
a new account (or login if you already have an account).

* Once logged in, click on `My SmartApps`. This will show a list of the current SmartApps
installed (it could be blank for new accounts). 

* Click the `New SmartApp` button. The "New SmartApp" form page will appear.

* Enter a name for your SmartApp (suggestion: GoSmart Webservices Handler)

* Enter your GitHub user Id in the namespace (probably doesn't matter much).

* Enter the author, name and category.

* Under "Oauth", click on `Enable OAuth in Smart App`. More fields will be added to the
current form.

* Take note of the "Client ID" and "Client Secret". These will be used to authenticate and
retrieve a token. Once the token is saved locally by the library, authentication can proceed
without user intervention.

* On `Redirect URI`, enter `http://localhost:4567/OAuthCallback`. *Case is important here*

* The application editor will open with a basic App template. Completely delete the editor
contents and replace it with the contents of `endpoint.groovy` in this package (copy & paste
are your friends).

* Click the `Save` Button to save your changes.

* Click the `Publish` Button and choose the `For Me` option. A green banner indicating success
should appear at the top of the screen.

* Click the `My SmartApps` link again (at the top of the screen). Make sure the new App shows with
status set to "Published" and OAuth set to "True".

At this point, the smartthings part of the installation should be ready.

## Running an example

You can easily run the examples in the `examples` directory and see some output:

* Change the current directory (cd) to `examples/simple` under the installation tree.

* Type `go build` to build the `simple` application. If everything goes right, a binary
called `simple` will be created under the current directory.

* Now run `simple` using the "Client ID" and "Client Secret" for our app. Replace
`client_id` and `client_secret` below with the values obtained at the time we created the App.
Type:

```
  ./simple --client client_id --secret client_secret
```

* Since this is the first time we try to authenticate this particular
Client ID and Client Secret pair, the authentication process will require
user intervention. A message will display on the screen asking the user to
visit [http://localhost:4567](http://localhost:4567) to complete the authentication
process. Visit this link with your favorite browser.  You'll be redirected to the
smartthings.com API website. Proceed to log in normally and indicate which
devices should be "seen" by this App (normally, all). Confirm your choices.

* At this point, the `simple` program will proceed and show a (crude) output showing
some information about your sensors.

* Try running simple again. This time, omit the `--secret` command line option. Notice how the full
login process is bypassed. This happens because the library retrieved the token from local storage.
The `--client` option is only needed because `simple` uses it to form the name of the file containing
the token. You can also specify a `--tokenfile` option to force saving the token to a specific file.

**Important**: A third party will have  *full access* to your SmartThings IoT network if they
obtain your token. Make sure to save the token file in a safe location. By default the library saves the
token file under the current user's directory. You can change this behavior easily by specifying a
token filename during authentication time. Look at the examples for more details.
