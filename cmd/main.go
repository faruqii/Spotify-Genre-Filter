package main

import (
	"fmt"
	"log"

	"github.com/faruqii/Spotify-Genre-Filter/internal/auth"
	"github.com/faruqii/Spotify-Genre-Filter/internal/spotify"
	"github.com/faruqii/Spotify-Genre-Filter/pkg/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	client, err := auth.Authenticate(cfg.ClientID, cfg.ClientSecret, cfg.RedirectURI)
	if err != nil {
		log.Fatalf("failed to authenticate: %v", err)
	}

	likedSongs, err := spotify.GetLikedSongs(client)
	if err != nil {
		log.Fatalf("failed to get liked songs: %v", err)
	}

	genre := "shoegaze"
	filteredTracksIDs, err := spotify.FilterSongsByGenre(client, likedSongs, genre)
	if err != nil {
		log.Fatalf("failed to filter songs by genre: %v", err)
	}

	// create playlist
	playlistName := genre + " playlist"
	err = spotify.CreatePlaylist(client, playlistName, filteredTracksIDs)
	if err != nil {
		log.Fatalf("failed to create playlist: %v", err)
	}

	fmt.Printf("Playlist created successfully: %s\n", playlistName)
}
