package spotify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Track represents a Spotify track
type Track struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Artists []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"artists"`
}

// Artist represents a Spotify artist
type Artist struct {
	Genres []string `json:"genres"`
}

// GetLikedSongs fetches the user's liked songs
func GetLikedSongs(client *http.Client) ([]Track, error) {
	var likedSongs []Track
	limit := 50
	offset := 0

	for {
		url := fmt.Sprintf("https://api.spotify.com/v1/me/tracks?limit=%d&offset=%d", limit, offset)
		resp, err := client.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Read the raw response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %v", err)
		}

		// Check if the response is an error
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("API returned non-200 status: %d, response: %s", resp.StatusCode, string(body))
		}

		var result struct {
			Items []struct {
				Track Track `json:"track"`
			} `json:"items"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %v", err)
		}

		for _, item := range result.Items {
			likedSongs = append(likedSongs, item.Track)
		}

		if len(result.Items) < limit {
			break
		}
		offset += limit
	}

	return likedSongs, nil
}

// GetArtistDetails fetches details for a specific artist
func GetArtistDetails(client *http.Client, artistID string) (*Artist, error) {
	url := fmt.Sprintf("https://api.spotify.com/v1/artists/%s", artistID)
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch artist details: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read artist details response: %v", err)
	}

	var artist Artist
	if err := json.Unmarshal(body, &artist); err != nil {
		return nil, fmt.Errorf("failed to unmarshal artist details: %v", err)
	}

	return &artist, nil
}

// FilterSongsByGenre filters songs by a specific genre
func FilterSongsByGenre(client *http.Client, songs []Track, genre string) ([]string, error) {
	var filteredTrackIDs []string

	for _, song := range songs {
		// Fetch details for the first artist of the track
		artistID := song.Artists[0].ID
		artist, err := GetArtistDetails(client, artistID)
		if err != nil {
			return nil, fmt.Errorf("failed to get artist details: %v", err)
		}

		// Check if the artist's genres match the desired genre
		for _, g := range artist.Genres {
			if g == genre {
				filteredTrackIDs = append(filteredTrackIDs, song.ID)
				break
			}
		}
	}

	return filteredTrackIDs, nil
}

// CreatePlaylist creates a new playlist and adds tracks to it
func CreatePlaylist(client *http.Client, name string, trackIDs []string) error {
	// Get the current user's ID
	userResp, err := client.Get("https://api.spotify.com/v1/me")
	if err != nil {
		return fmt.Errorf("failed to fetch user details: %v", err)
	}
	defer userResp.Body.Close()

	var user struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(userResp.Body).Decode(&user); err != nil {
		return fmt.Errorf("failed to unmarshal user details: %v", err)
	}

	// Create the playlist
	playlistData := map[string]interface{}{
		"name":        name,
		"description": "Created by Spotify Playlist Manager",
		"public":      false,
	}

	playlistBody, err := json.Marshal(playlistData)
	if err != nil {
		return fmt.Errorf("failed to marshal playlist data: %v", err)
	}

	playlistResp, err := client.Post(
		fmt.Sprintf("https://api.spotify.com/v1/users/%s/playlists", user.ID),
		"application/json",
		bytes.NewReader(playlistBody),
	)
	if err != nil {
		return fmt.Errorf("failed to create playlist: %v", err)
	}
	defer playlistResp.Body.Close()

	var playlist struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(playlistResp.Body).Decode(&playlist); err != nil {
		return fmt.Errorf("failed to unmarshal playlist response: %v", err)
	}

	// Add tracks to the playlist
	trackURIs := make([]string, len(trackIDs))
	for i, id := range trackIDs {
		trackURIs[i] = fmt.Sprintf("spotify:track:%s", id)
	}

	addTracksData := map[string]interface{}{
		"uris": trackURIs,
	}
	addTracksBody, err := json.Marshal(addTracksData)
	if err != nil {
		return fmt.Errorf("failed to marshal track URIs: %v", err)
	}

	addTracksResp, err := client.Post(
		fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlist.ID),
		"application/json",
		bytes.NewReader(addTracksBody),
	)
	if err != nil {
		return fmt.Errorf("failed to add tracks to playlist: %v", err)
	}
	defer addTracksResp.Body.Close()

	// Log the response for debugging
	body, err := io.ReadAll(addTracksResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read add tracks response: %v", err)
	}
	log.Printf("Add tracks response: %s", string(body))

	return nil
}

// FindExistingPlaylist checks if a playlist with the same name exists
func FindExistingPlaylist(client *http.Client, name string) (string, error) {
	userResp, err := client.Get("https://api.spotify.com/v1/me")
	if err != nil {
		return "", fmt.Errorf("failed to fetch user details: %v", err)
	}
	defer userResp.Body.Close()

	var user struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(userResp.Body).Decode(&user); err != nil {
		return "", fmt.Errorf("failed to unmarshal user details: %v", err)
	}

	url := fmt.Sprintf("https://api.spotify.com/v1/users/%s/playlists?limit=50", user.ID)
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch playlists: %v", err)
	}
	defer resp.Body.Close()

	var playlists struct {
		Items []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&playlists); err != nil {
		return "", fmt.Errorf("failed to unmarshal playlists: %v", err)
	}

	for _, p := range playlists.Items {
		if p.Name == name {
			return p.ID, nil // Return existing playlist ID
		}
	}
	return "", nil // No existing playlist found
}

// AddTracksToPlaylist adds tracks to an existing playlist
func AddTracksToPlaylist(client *http.Client, playlistID string, trackIDs []string) error {
	trackURIs := make([]string, len(trackIDs))
	for i, id := range trackIDs {
		trackURIs[i] = fmt.Sprintf("spotify:track:%s", id)
	}

	addTracksData := map[string]interface{}{
		"uris": trackURIs,
	}
	addTracksBody, err := json.Marshal(addTracksData)
	if err != nil {
		return fmt.Errorf("failed to marshal track URIs: %v", err)
	}

	url := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlistID)
	req, err := http.NewRequest("POST", url, bytes.NewReader(addTracksBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to add tracks to playlist: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add tracks, status: %d, response: %s", resp.StatusCode, string(body))
	}

	log.Printf("Tracks added successfully to playlist %s", playlistID)
	return nil
}

// CreateOrUpdatePlaylist checks for an existing playlist and updates it, otherwise creates a new one
func CreateOrUpdatePlaylist(client *http.Client, name string, trackIDs []string) error {
	playlistID, err := FindExistingPlaylist(client, name)
	if err != nil {
		return err
	}

	if playlistID != "" {
		log.Printf("Updating existing playlist: %s", name)
		return AddTracksToPlaylist(client, playlistID, trackIDs)
	}

	// If no existing playlist, create a new one
	return CreatePlaylist(client, name, trackIDs)
}
