/**

An example HTTP server that demonstrates how to use the Go Spotify Library.

**/

package main

import (
	"fmt"
	s "github.com/jbuberel/spotify_client"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type ResponseUserAgent struct {
	Useragent string `json:"user-agent"`
}

// A Params is a map containing URL parameters.
type Params map[string]string

// These are the three variables that you'll need to provide,
// based on the configuration of your Spotify Developer App:
//   see https://developer.spotify.com/my-applications/
// for more information
//
// The init funciton on this sample server will look for
// environment variables - client_id and client_secret - and use
// those values if it finds them. Or you can you supply them
// directly in source code here:
var ClientId s.ClientId = ""
var ClientSecret s.ClientSecret = ""
var RedirectUri s.RedirectUri = "http://localhost:8080/callback/"

// The init function will look through environment variables
// to find the client_id and client_secret, which need to come from your
// Spotify Developer Applications settings - see https://developer.spotify.com/my-applications/
// for more information.
func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)

	log.Printf("Looking through env vars\n")
	for _, e := range os.Environ() {
		parts := strings.Split(e, "=")
		if len(parts) == 2 {
			if parts[0] == "client_id" {
				ClientId = s.ClientId(parts[1])
				log.Printf("Client ID set from environ to: %v\n", ClientId)
			} else if parts[0] == "client_secret" {
				ClientSecret = s.ClientSecret(parts[1])
				log.Printf("Client Secret set from environ to: %v\n", ClientSecret)
			}
		}
	}
}

// This function is required to start the end-user visible
// authentication process. In this example, this function
// is configured to handle '/login/' - as in:
//    http://myhost.mydomain.com/login/
// It will generate an HTTP 302 response, sending the use to the Spotify Account
// authentication page. If the user approves the request, they'll be sent back to the
// redirectUri value, which MUST exactly match one of the whitelisted callback URIs
// that you configured in your Spotify Application - see
// https://developer.spotify.com/my-applications/ for more information.
func sendLogion(w http.ResponseWriter, r *http.Request) {
	redirectUri := url.QueryEscape("http://localhost:8080/callback/")
	scopes := url.QueryEscape("playlist-read-private playlist-modify-private user-read-private")
	http.Redirect(w, r, "https://accounts.spotify.com/authorize?client_id="+string(ClientId)+"&scope="+scopes+"&response_type=code&redirect_uri="+redirectUri, 302)

}

// This is the method that the user's browser session will be
// redirected to after they complete the spotify authentication and approval.
// It is currently configured to respond to '/callback/' - as in:
//   http://myhost.mydomain.com/callback/
// This can be modified in the main function below. The spotify server
// will append a 'code' query parameter to the URL, which needs to be
// used by the library to retrieve the authentication token.
//
// The method will then connect directly to the spotify API service
// using the library methods to retrieve detailed information about the user
//

func authCallback(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "text/html")
	log.Printf("Callback %s", r.URL.Path[1:])
	var code string = r.URL.Query()["code"][0]
	log.Printf("Code: %v\n", code)

	tokenResponse, err := s.GetAccessToken(code, ClientId, ClientSecret, RedirectUri)
	if err != nil {
		log.Println(err)
	}
	log.Println(tokenResponse.AccessToken)

	userInfoResponse, err := s.GetUserInfo(tokenResponse.AccessToken)
	if err != nil {
		log.Println(err)
	}

	username := userInfoResponse.Id
	log.Printf("Username: %v\n", username)

	http.Redirect(w, r, "/listplaylists/"+ string(username)+"/" +string(tokenResponse.AccessToken), http.StatusFound)
}
	

// This method will response to requests starting with /listplaylists/ and it expects the URL path to include:
// 		/listplaylists/{username}/{access_token}
// For each playlist retrieved, it will generate an <a href...> tag that links
// back to this server with information about the playlist encoded in the URL path
// information in this format:
//   /tracks/{username}/{access_token}/{playlist_id}
func listPlaylists(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	username := s.Username(parts[2])
	accessToken := s.AccessToken(parts[3])

	playlistItems, err := s.GetUserPlaylists(accessToken, username)
	if err != nil {
		log.Println(err)
	}

	for _, i := range playlistItems {
		log.Printf(" [%v]:[%v]\n", i.Id, i.Name)
		fmt.Fprintf(w, "<a href=\"/tracks/%v/%v/%v\">List tracks - %v</a> - \n", i.Owner.Id, accessToken, i.Id, i.Name)
		fmt.Fprintf(w, "<a href=\"/duplicate/%v/%v/%v/%v\">Duplicate - %v</a><br/>\n", i.Owner.Id, username, accessToken, i.Id, i.Name)
	}
}

// This function handles calls to URLs starting with /tracks/ and it expects
// that the playlist information is encoded into the URL in the following format:
//    /tracks/{username}/{access_token}/{playlist_id}
//
// Using the information in that URL, it will retrieve the contents of the playlist
// and list them on the page.
func showTracks(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "text/html")
	log.Printf("URL: %v\n", r.URL)
	parts := strings.Split(r.URL.Path, "/")
	username := s.Username(parts[2])
	accessToken := s.AccessToken(parts[3])
	playlistId := parts[4]
	for n, p := range parts {
		fmt.Fprintf(w, "Tracks path %v %v!<br/>", n, p)
	}

	tracks, err := s.GetTracksForPlaylist(accessToken, username, playlistId)
	if err != nil {
		log.Println(err)
		return
	}
	for _, track := range tracks {
		fmt.Fprintf(w, "<p>%v - %v - %v </p><br/>\n", track.Id, track.Name, track.Album.Name)
		for _, artist := range track.Artists {
			fmt.Fprintf(w, "<p>%v</p><br/>\n", artist.Name)
		}
	}

}

// This function handles calls to URLs starting with /duplicate/ and it
// expects the playlist information is encoded in the URL as follows:
//    /tracks/{playlist_owner_username}/{playlist_creator_username}/{access_token}/{playlist_id}
//
// Using the information in that URL, it will retrieve the contents of the playlist
// and create a new playlist named "Copy of $OLD_NAME" contianing the same set of tracks.
func duplicatePlaylist(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "text/html")
	log.Printf("URL: %v\n", r.URL)
	parts := strings.Split(r.URL.Path, "/")
	playlistOwner := s.Username(parts[2])
	playlistCreator := s.Username(parts[3])
	accessToken := s.AccessToken(parts[4])
	playlistId := parts[5]
	for n, p := range parts {
		fmt.Fprintf(w, "Tracks path %v %v!<br/>", n, p)
	}

	// important to get information on the existing playlist before creating the copy
	playlist, err := s.GetPlaylistInfo(accessToken, playlistOwner, playlistId)
	if err != nil {
		log.Println(err)
		return
	}
	
	tracks, err := s.GetTracksForPlaylist(accessToken, playlistOwner, playlistId)
	
	fmt.Fprintf(w, "<p>Original: %v-%v </p><br/>\n", playlist.Id, playlist.Name)

	
	duplicatePlaylist, err := s.CreatePlaylist(accessToken,playlistCreator , "Copy of " + playlist.Name, false)
	
	fmt.Fprintf(w, "<p>Copy: %v-%v </p><br/>\n", duplicatePlaylist.Id, duplicatePlaylist.Name)
	
	addTracksResponse, err := s.AddTracksToPlaylist(accessToken, playlistCreator , duplicatePlaylist, tracks )
	if err != nil {
		log.Printf("Error adding tracks to playlist %v\n", duplicatePlaylist.Id, err)
		return
	}
	fmt.Fprintf(w, "<p>Snapshot ID: %v </p><br/>\n", addTracksResponse.SnapshotId)


}

// Here you can configure the handler functions for each of the three request types.
//
// It is also recommended that you invoke the server with the following style of command in
// order to securely provide the client_id and client_secret values:
//
// $ client_secret=76c...eeb client_id=f72...125 bash -c 'go run jbuberel/spotify_client/example/server.go'
func main() {
	http.HandleFunc("/login/", sendLogion)
	http.HandleFunc("/callback/", authCallback)
	http.HandleFunc("/tracks/", showTracks)
	http.HandleFunc("/duplicate/", duplicatePlaylist)
	http.HandleFunc("/listplaylists/", listPlaylists)
	http.ListenAndServe(":8080", nil)
}
