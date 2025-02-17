package spotify

import "github.com/zmb3/spotify"

func GetLikedSongs(client *spotify.Client) ([]spotify.SavedTrack, error) {
	var likedSongs []spotify.SavedTrack
	limit := 50
	offset := 0

	for {
		result, err := client.CurrentUsersTracksOpt(&spotify.Options{Limit: &limit, Offset: &offset})
		if err != nil {
			return nil, err
		}
		likedSongs = append(likedSongs, result.Tracks...)
		if len(result.Tracks) < limit {
			break
		}
		offset += limit
	}

	return likedSongs, nil
}

func FilterSongsByGenre(client *spotify.Client, songs []spotify.SavedTrack, genre string) ([]spotify.ID, error) {
	var filteredTracksIDs []spotify.ID

	for _, song := range songs {
		artist, err := client.GetArtist(song.Artists[0].ID)
		if err != nil {
			return nil, err
		}

		for _, genre := range artist.Genres {
			if genre == genre {
				filteredTracksIDs = append(filteredTracksIDs, song.ID)
				break
			}
		}
	}

	return filteredTracksIDs, nil
}
