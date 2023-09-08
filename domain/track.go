package domain

type Track struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Artist string `json:"artist"`
}

func NewTrack(id, name, artist string) *Track {
	return &Track{
		ID:     id,
		Name:   name,
		Artist: artist,
	}
}
