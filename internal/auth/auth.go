package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"

	"github.com/zmb3/spotify"
)

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

var ch = make(chan *spotify.Client)

func Authenticate(clientID, clientSecret, redirectURI string) (*spotify.Client, error) {
	auth := spotify.NewAuthenticator(redirectURI, spotify.ScopeUserLibraryRead, spotify.ScopePlaylistModifyPrivate)

	state := generateState()

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		tok, err := auth.Token(state, r)
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusForbidden)
			log.Fatal(err)
		}
		if st := r.FormValue("state"); st != state {
			http.NotFound(w, r)
			log.Fatalf("State mismatch: %s != %s\n", st, state)
		}

		client := auth.NewClient(tok)
		fmt.Fprint(w, "Login Completed!")
		ch <- &client
	})
	go http.ListenAndServe(":8080", nil)

	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	client := <-ch
	return client, nil
}
