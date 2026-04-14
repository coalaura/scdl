package scdl

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Track holds metadata for a SoundCloud track.
type Track struct {
	ID                 int
	Title              string
	Album              string
	Artist             string
	ArtworkURL         string
	ArtistAvatarURL    string
	Genre              string
	Description        string
	Year               string
	Duration           int // milliseconds
	TrackAuthorization string
	HLSURL             string // HLS transcoding URL for audio/mpeg
}

var hydrationRe = regexp.MustCompile(`window\.__sc_hydration\s*=\s*(\[.+?]);`)

// GetTrack fetches a SoundCloud track page and extracts metadata from the
// hydration data embedded in the HTML.
func (c *Client) GetTrack(ctx context.Context, trackURL string) (*Track, error) {
	body, err := c.get(ctx, trackURL)
	if err != nil {
		return nil, fmt.Errorf("fetch track page: %w", err)
	}

	matches := hydrationRe.FindSubmatch(body)
	if len(matches) < 2 {
		return nil, fmt.Errorf("hydration data not found on page")
	}

	var hydration []struct {
		Hydratable string          `json:"hydratable"`
		Data       json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(matches[1], &hydration); err != nil {
		return nil, fmt.Errorf("parse hydration JSON: %w", err)
	}

	for _, entry := range hydration {
		if entry.Hydratable != "sound" {
			continue
		}

		var data struct {
			ID                 int       `json:"id"`
			Title              string    `json:"title"`
			CreatedAt          time.Time `json:"created_at"`
			ReleaseDate        time.Time `json:"release_date"`
			Description        string    `json:"description"`
			Genre              string    `json:"genre"`
			Duration           int       `json:"duration"`
			ArtworkURL         string    `json:"artwork_url"`
			TrackAuthorization string    `json:"track_authorization"`
			PublisherMetadata  struct {
				Artist       string `json:"artist"`
				AlbumTitle   string `json:"album_title"`
				ReleaseTitle string `json:"release_title"`
			} `json:"publisher_metadata"`
			User struct {
				AvatarURL string `json:"avatar_url"`
				Username  string `json:"username"`
			} `json:"user"`
			Media struct {
				Transcodings []struct {
					URL    string `json:"url"`
					Format struct {
						Protocol string `json:"protocol"`
						MimeType string `json:"mime_type"`
					} `json:"format"`
				} `json:"transcodings"`
			} `json:"media"`
		}
		if err := json.Unmarshal(entry.Data, &data); err != nil {
			return nil, fmt.Errorf("parse track data: %w", err)
		}

		artist := data.User.Username
		if data.PublisherMetadata.Artist != "" {
			artist = data.PublisherMetadata.Artist
		}

		title := data.Title
		if data.PublisherMetadata.ReleaseTitle != "" {
			title = data.PublisherMetadata.ReleaseTitle
		} else {
			title = cleanupTrackTitle(title, artist)
		}

		var year string
		if !data.ReleaseDate.IsZero() {
			year = data.ReleaseDate.Format("2006")
		} else if !data.CreatedAt.IsZero() {
			year = data.CreatedAt.Format("2006")
		}

		track := &Track{
			ID:                 data.ID,
			Title:              title,
			Album:              data.PublisherMetadata.AlbumTitle,
			Artist:             artist,
			ArtworkURL:         data.ArtworkURL,
			ArtistAvatarURL:    data.User.AvatarURL,
			Genre:              data.Genre,
			Description:        data.Description,
			Year:               year,
			Duration:           data.Duration,
			TrackAuthorization: data.TrackAuthorization,
		}

		for _, t := range data.Media.Transcodings {
			if t.Format.MimeType == "audio/mpeg" && t.Format.Protocol == "hls" {
				track.HLSURL = t.URL
				break
			}
		}

		if track.HLSURL == "" {
			return nil, fmt.Errorf("no HLS audio/mpeg transcoding found")
		}

		return track, nil
	}

	return nil, fmt.Errorf("no sound entry found in hydration data")
}

// cleanupTrackTitle removes the artist name from the title (e.g. "CapoBlanco - Love In The Rain")
func cleanupTrackTitle(title, artist string) string {
	if artist == "" || !strings.HasPrefix(title, artist) {
		return title
	}

	// Remove artist prefix and any surrounding whitespace/hyphen
	res := strings.TrimPrefix(title, artist)
	res = strings.TrimSpace(res)
	if strings.HasPrefix(res, "-") {
		return strings.TrimSpace(res[1:])
	}

	return title
}
