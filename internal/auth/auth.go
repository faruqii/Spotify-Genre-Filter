package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

var (
	oauthConfig *oauth2.Config
	state       string
)

// generateState generates a random state string for OAuth2
func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Authenticate sets up the OAuth2 configuration and starts the authentication flow
func Authenticate(clientID, clientSecret, redirectURI string) (*http.Client, error) {
	// Set up OAuth2 configuration
	oauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Scopes:       []string{"user-library-read", "playlist-modify-private"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.spotify.com/authorize",
			TokenURL: "https://accounts.spotify.com/api/token",
		},
	}

	// Generate a random state
	state = generateState()

	// Start an HTTP server to handle the callback
	http.HandleFunc("/callback", handleCallback)
	go http.ListenAndServe(":8080", nil)

	// Redirect the user to Spotify's authorization page
	authURL := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", authURL)

	// Wait for the callback to complete and return the authenticated HTTP client
	client := <-ch
	return client, nil
}

// handleCallback handles the OAuth2 callback from Spotify
func handleCallback(w http.ResponseWriter, r *http.Request) {
	// Check for errors in the callback
	if err := r.FormValue("error"); err != "" {
		http.Error(w, "Error from Spotify: "+err, http.StatusBadRequest)
		return
	}

	// Verify the state parameter
	if r.FormValue("state") != state {
		http.Error(w, "State mismatch", http.StatusBadRequest)
		return
	}

	// Exchange the authorization code for an access token
	code := r.FormValue("code")
	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create an authenticated HTTP client
	client := oauthConfig.Client(context.Background(), token)
	fmt.Fprint(w, "Login Completed!")
	ch <- client
}

var ch = make(chan *http.Client)
