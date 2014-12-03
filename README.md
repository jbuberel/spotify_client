spotify_client
==============

A Spotify Web Service API client written in Go. The spotify_client project aims to become a full featured, robust library for Golang that can be used to build cloud-based applications that interact with the Spotify developer APIs.

## Installation

This is still a work in progress, and the library does not support the full Spotify Web Service API yet. 
Before you get too far, you'll need to create a Spotify Developer account. You can do that by following
the instructions at:

https://developer.spotify.com/my-applications/

Once you've created your developer account and registered your spotify sample application, you'll need
to provide the 'client_id' and 'client_secret' data in order for this library to talk to the Spotify servers.
In the server.go file, you'll find these two lines where you can insert those values:

```
var ClientId s.ClientId = ""
var ClientSecret s.ClientSecret = ""
```
All of the interfaces you'll need are in the lib.go file. I have also included a small example web server
that uses all of the existing API features.

Running the example web server will start an HTTP listener on your local host at port 8080. You can start 
the server with:

```
$> go run server.go
```

You can then point your local browser at http://localhost:8080/. When that page loads you will be redirected 
to the Spotify OAuth user login page. Enter your spotify credentials. Once authenticated, the example
server will then retrieve a list of your playlists. Clicking on the linked playlist will then retrieve
the track list and render that to simple (ugly) HTML.


## API Reference

For reference, this library is built to use the official Spotify Web API. More details are visible here: https://developer.spotify.com/web-api/

## Contributors

If you're a Spotify fan who'd like to develop applications in Go, I'd love to hear from you! Contact the author - jason@buberel.org
