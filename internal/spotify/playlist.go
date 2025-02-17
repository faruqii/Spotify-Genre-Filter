package spotify

import "github.com/zmb3/spotify"


func CreatePlaylist(client *spotify.Client, name string, trackIDs []spotify.ID) error {
	user, err := client.CurrentUser()
	if err != nil {
		return err
	}

	playlist, err := client.CreatePlaylistForUser(user.ID, name, "Created by Spotify Playlist Manager", false)
	if err != nil {
		return err
	}

	_, err = client.AddTracksToPlaylist(playlist.ID, trackIDs...)
	return err
}