/**
	A client library for the Spotify Web Service API, written in the Go
	programming language (golang). Use of this library requires
	a Spotify Developer Account. Start here to set one up:
	
	https://developer.spotify.com/my-applications/
	
	See the README.md for more information.
	
	
**/
package spotify_client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmcvetta/napping"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

func init() {
	log.SetFlags(log.Ltime | log.Lshortfile)

}

type ResponseUserAgent struct {
	Useragent string `json:"user-agent"`
}

// A Params is a map containing URL parameters.
type Params map[string]string
type TokenResponse struct {
	AccessToken  AccessToken `json:"access_token"`
	TokenType    string      `json:"token_type"`
	ExpiresIn    int64       `json:"expires_in"`
	RefreshToken string      `json:"refresh_token"`
}

type PlaylistResponse struct {
	Href     string     `json:"href"`
	Limit    int32      `json:"limit"`
	Offset   int32      `json:"offset"`
	Next     string     `json:"next"`
	Previous string     `json:"previous"`
	Total    int32      `json:"total"`
	Items    []Playlist `json:"items"`
}


type CreatePlaylistRequest struct {
	Name     string     `json:"name"`
	Public	 bool 		`json:"public"`
}

type AddTrackToPlaylistRequest struct {
	Uris	[]string	`json:"uris"`
}

type AddTrackToPlaylistResponse struct {
	SnapshotId	string	`json:"snapshot_id"`
}

type Playlist struct {
	Href string `json:"href"`
	Id   string `json:"id"`
	Name string `json:"name"`
	Owner	PlaylistOwner	`json:"owner"`
}

type PlaylistOwner struct {
	Href	string `json:"href"`
	Id		string `json:"id"`
}

type TracklistResponse struct {
	Href     string          `json:"href"`
	Items    []PlaylistTrack `json:"items"`
	Limit    int32           `json:"limit"`
	Offset   int32           `json:"offset"`
	Next     string          `json:"next"`
	Previous string          `json:"previous"`
	Total    int32           `json:"total"`
}

type PlaylistTrack struct {
	Track Track `json:"track"`
}

type Track struct {
	Id      string   `json:"id"`
	Href    string   `json:"href"`
	Name    string   `json:"name"`
	Album   Album    `json:"album"`
	Artists []Artist `json:"artists"`
}

type Artist struct {
	Href string `json:"href"`
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Album struct {
	AlbumType string `json:"album_type"`
	Href      string `json:"href"`
	Id        string `json:"id"`
	Name      string `json:"name"`
}

type UserInfoResponse struct {
	Id          Username `json:"id"`
	URI         string   `json:"uri"`
	DisplayName string   `json:"display_name"`
	Email       string   `json:"email"`
}

// Type overrides to ensure various string-like values don't get mixed up.
type AccessToken string
type Username string
type ClientId string
type ClientSecret string
type RedirectUri string

// Given an access code returned by the spotify web server, along with the
// Client ID and Client Secret for your spotify app (see: https://developer.spotify.com/my-applications/)
// this method will retrieve an access token, returned as type TokenResponse
func GetAccessToken(accessCode string, clientId ClientId, clientSecret ClientSecret, redirectUri RedirectUri) (*TokenResponse, error) {

	resp, err := http.PostForm("https://accounts.spotify.com/api/token",
		url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {accessCode},
			"redirect_uri":  {string(redirectUri)},
			"client_id":     {string(clientId)},
			"client_secret": {string(clientSecret)},
		})

	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	log.Printf("Status code %v\n", resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	if body != nil {
		var tokenResponse = new(TokenResponse)
		err := json.Unmarshal(body, &tokenResponse)
		if err == nil {
			log.Printf("JSON: %v+\n", tokenResponse)
			log.Printf("Access Token: %v\n", tokenResponse.AccessToken)
			return tokenResponse, nil
		} else {
			log.Println(err)
			return nil, err
		}
	}

	log.Println("Empty response body")
	return nil, errors.New("Empty response body")

}

// Given an AccessToken returned by the GetAccessToken method, this
// function will retrieve information about the authenticated user.
// This information is used to retrieve their playlists later.
func GetUserInfo(accessToken AccessToken) (*UserInfoResponse, error) {

	s := napping.Session{}
	header := http.Header{}
	header.Add("Authorization", "Bearer "+string(accessToken))
	s.Header = &header

	res := ResponseUserAgent{}
	url := "https://api.spotify.com/v1/me"

	res = ResponseUserAgent{}
	resp, err := s.Get(url, nil, &res, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var userInfoResponse = new(UserInfoResponse)
	err = resp.Unmarshal(&userInfoResponse)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return userInfoResponse, nil
}

// Given an AccessToken and a UserName (retrieved using the GetAccessToken and GetUserInfo functions),
// this function will extract all of the user's playlists as a slice of Playlist objects.
func GetUserPlaylists(accessToken AccessToken, username Username) ([]Playlist, error) {

	s := napping.Session{}
	header := http.Header{}
	header.Add("Authorization", "Bearer "+string(accessToken))
	s.Header = &header

	res := ResponseUserAgent{}

	offset := 0
	limit := 5

	playlistItems := make([]Playlist, 0, 1)
	for done := false; done == false; {

		res = ResponseUserAgent{}
		url := fmt.Sprintf("https://api.spotify.com/v1/users/%v/playlists?limit=%v&offset=%v", username, limit, offset)

		resp, err := s.Get(url, nil, &res, nil)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Printf("Response URL: %v\n", resp.Url)
		var playlistResponse = new(PlaylistResponse)

		err = resp.Unmarshal(&playlistResponse)
		if err != nil {
			log.Println(err)
		}
		log.Printf("Items Length: %v\n", len(playlistResponse.Items))
		log.Printf("Total items: %v\n", playlistResponse.Total)

		for _, item := range playlistResponse.Items {
			playlistItems = append(playlistItems, item)
		}
		if int32(len(playlistItems)) >= playlistResponse.Total {
			done = true
		} else {
			offset += limit
		}
		log.Printf("Accumulated Items: %v\n", len(playlistItems))
	}

	return playlistItems, nil
}


// Given an AccessToken, UserName (retrieved using the GetAccessToken and GetUserInfo functions),
// and Playlist ID, this function will return a simplified Playlist object
func GetPlaylistInfo(accessToken AccessToken, username Username, playlistId string) (Playlist, error) {

	s := napping.Session{}
	header := http.Header{}
	header.Add("Authorization", "Bearer "+string(accessToken))
	s.Header = &header

	res := ResponseUserAgent{}

	res = ResponseUserAgent{}
	reqUrl := fmt.Sprintf("https://api.spotify.com/v1/users/%v/playlists/%v", username, playlistId)

	resp, err := s.Get(reqUrl, nil, &res, nil)
	if err != nil {
		log.Println(err)
		return Playlist{}, err
	}
	log.Printf("Response URL: %v\n", resp.Url)
	var playlist = new(Playlist)

	err = resp.Unmarshal(&playlist)
	if err != nil {
		log.Println(err)
	}

	return *playlist, nil
}

// Given an AccessToken, UserName (retrieved using the GetAccessToken and GetUserInfo functions),
// and new playlist name, this method will create a new playlist.
func CreatePlaylist(accessToken AccessToken, username Username, playlistName string, playlistPublic bool) (Playlist, error) {

	s := napping.Session{}
	header := http.Header{}
	header.Add("Authorization", "Bearer "+string(accessToken))
	s.Header = &header

	res := ResponseUserAgent{}

	res = ResponseUserAgent{}
	reqUrl := fmt.Sprintf("https://api.spotify.com/v1/users/%v/playlists", username)

	s.Log = false
	payload := &CreatePlaylistRequest{ 
			Name: playlistName, 
			Public: playlistPublic,
			}
	resp, err := s.Post(reqUrl, payload, &res, nil)
	s.Log = false
	if err != nil {
		log.Println(err)
		return Playlist{}, err
	}
	log.Printf("Response URL: %v\n", resp.Url)
	log.Printf("Response Text: %v\n", resp.RawText())
	var playlist = new(Playlist)

	err = resp.Unmarshal(&playlist)
	if err != nil {
		log.Println(err)
	}

	return *playlist, nil
}

// Method will add tracks to an existing playlist
func AddTracksToPlaylist(accessToken AccessToken, username Username, playlist Playlist, tracks []Track) (AddTrackToPlaylistResponse, error) {
		
	s := napping.Session{}
	header := http.Header{}
	header.Add("Authorization", "Bearer "+string(accessToken))
	s.Header = &header

	res := ResponseUserAgent{}
	res = ResponseUserAgent{}
	reqUrl := fmt.Sprintf("https://api.spotify.com/v1/users/%v/playlists/%v/tracks", string(username), playlist.Id)
	
	// Spotify's Add Tracks API only allows you to add 100 tracks at a time, 
	// so we need to paginate our POST's if the list of tracks > 100:
	var addTrackResponse = new (AddTrackToPlaylistResponse)
	for i := 0; i < len(tracks); i += 100 {
		addTracksRequest := new(AddTrackToPlaylistRequest)
		x := len(tracks[i:])
		if x > 100 {
			x = i +  100
		} else {
			x = i + len(tracks[i:])
		}
		for _, t := range tracks[i:x] {
			log.Printf("Adding track to playlist: %v-%v\n", t.Id, t.Name)
			addTracksRequest.Uris = append(addTracksRequest.Uris,  "spotify:track:" + t.Id)
		}
		
		resp, err := s.Post(reqUrl, addTracksRequest, &res, nil)
		if err != nil {
			log.Println(err)
			return AddTrackToPlaylistResponse{}, err
		}
		
		addTrackResponse = new (AddTrackToPlaylistResponse)
		err = resp.Unmarshal(&addTrackResponse)
		if err != nil {
			log.Println(err)
			return AddTrackToPlaylistResponse{}, err
		}

	}
	
	
	return *addTrackResponse, nil
	
}

// For a given user and playlist, this method will return track listings for
// each entry in the selected playlist as a slice of Track objects.
func GetTracksForPlaylist(accessToken AccessToken, owner Username, playlistId string) ([]Track, error) {

	s := napping.Session{}
	header := http.Header{}
	header.Add("Authorization", "Bearer "+string(accessToken))
	s.Header = &header

	res := ResponseUserAgent{}

	offset := 0
	limit := 5

	tracks := make([]Track, 0, 1)
	for done := false; done == false; {

		res = ResponseUserAgent{}
		url := fmt.Sprintf("https://api.spotify.com/v1/users/%v/playlists/%v/tracks?limit=%v&offset=%v", owner, playlistId, limit, offset)

		resp, err := s.Get(url, nil, &res, nil)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		log.Printf("Response URL: %v\n", resp.Url)
		log.Printf("Body: %v\n", resp.RawText())
		var tracklistResponse = new(TracklistResponse)

		err = resp.Unmarshal(&tracklistResponse)
		if err != nil {
			log.Println(err)
		}
		log.Printf("Items Length: %v\n", len(tracklistResponse.Items))
		log.Printf("Total items: %v\n", tracklistResponse.Total)

		for _, item := range tracklistResponse.Items {
			tracks = append(tracks, item.Track)
		}
		if int32(len(tracks)) >= tracklistResponse.Total {
			done = true
		} else {
			offset += limit
		}
		log.Printf("Accumulated Items: %v\n", len(tracks))
	}
	return tracks, nil
}
