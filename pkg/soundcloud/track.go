package soundcloud

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// Track holds metadata for a SoundCloud track.
type Track struct {
	ID                 int
	Title              string
	Artist             string
	ArtworkURL         string
	Genre              string
	Description        string
	Duration           int // milliseconds
	TrackAuthorization string
	HLSURL             string // HLS transcoding URL for audio/mpeg
}

var hydrationRe = regexp.MustCompile(`window\.__sc_hydration\s*=\s*(\[.+?\]);`)

// GetTrack fetches a SoundCloud track page and extracts metadata from the
// hydration data embedded in the HTML.
func (c *Client) GetTrack(trackURL string) (*Track, error) {
	body, err := c.get(trackURL)
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
			ID                 int    `json:"id"`
			Title              string `json:"title"`
			Description        string `json:"description"`
			Genre              string `json:"genre"`
			Duration           int    `json:"duration"`
			ArtworkURL         string `json:"artwork_url"`
			TrackAuthorization string `json:"track_authorization"`
			User               struct {
				Username string `json:"username"`
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

		track := &Track{
			ID:                 data.ID,
			Title:              data.Title,
			Artist:             data.User.Username,
			ArtworkURL:         data.ArtworkURL,
			Genre:              data.Genre,
			Description:        data.Description,
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
