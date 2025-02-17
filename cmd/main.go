package main

import (
	"fmt"
	"log"

	"github.com/faruqii/Spotify-Genre-Filter/internal/auth"
	"github.com/faruqii/Spotify-Genre-Filter/internal/spotify"
	"github.com/faruqii/Spotify-Genre-Filter/pkg/config"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("failed to load .env file: %v", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Authenticate with Spotify
	client, err := auth.Authenticate(cfg.ClientID, cfg.ClientSecret, cfg.RedirectURI)
	if err != nil {
		log.Fatalf("failed to authenticate: %v", err)
	}

	// Fetch liked songs
	likedSongs, err := spotify.GetLikedSongs(client)
	if err != nil {
		log.Fatalf("failed to get liked songs: %v", err)
	}

	// Filter songs by genre
	genre := "shoegaze"
	filteredTrackIDs, err := spotify.FilterSongsByGenre(client, likedSongs, genre)
	if err != nil {
		log.Fatalf("failed to filter songs by genre: %v", err)
	}

	// Create playlist
	playlistName := genre + " playlist"
	err = spotify.CreateOrUpdatePlaylist(client, playlistName, filteredTrackIDs)
	if err != nil {
		log.Fatalf("failed to create playlist: %v", err)
	}

	fmt.Printf("Playlist created successfully: %s\n", playlistName)
}
