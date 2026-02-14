package scdl

import (
	"encoding/json"
	"fmt"
	"strings"
)

// GetStreamURL resolves the M3U8 playlist URL for a track.
func (c *Client) GetStreamURL(track *Track) (string, error) {
	trackID, streamToken, err := parseHLSURL(track.HLSURL)
	if err != nil {
		return "", err
	}

	apiURL := fmt.Sprintf(
		"https://api-v2.soundcloud.com/media/soundcloud:tracks:%s/%s/stream/hls?client_id=%s&track_authorization=%s",
		trackID, streamToken, c.clientID, track.TrackAuthorization,
	)

	body, err := c.get(apiURL)
	if err != nil {
		return "", fmt.Errorf("fetch stream URL: %w", err)
	}

	var resp struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("parse stream response: %w", err)
	}

	if resp.URL == "" {
		return "", fmt.Errorf("empty stream URL in API response")
	}

	return resp.URL, nil
}

// parseHLSURL extracts the track ID and stream token from an HLS URL like:
// https://api-v2.soundcloud.com/media/soundcloud:tracks:123456/ab-cd-ef/stream/hls
func parseHLSURL(hlsURL string) (trackID, streamToken string, err error) {
	parts := strings.Split(hlsURL, "/")
	if len(parts) < 6 {
		return "", "", fmt.Errorf("invalid HLS URL format: %s", hlsURL)
	}

	// Find the part containing "soundcloud:tracks:"
	for i, part := range parts {
		if strings.HasPrefix(part, "soundcloud:tracks:") {
			idParts := strings.Split(part, ":")
			if len(idParts) < 3 {
				return "", "", fmt.Errorf("invalid track ID format in URL")
			}
			trackID = idParts[2]
			if i+1 < len(parts) {
				streamToken = parts[i+1]
			}
			return trackID, streamToken, nil
		}
	}

	return "", "", fmt.Errorf("track ID not found in HLS URL: %s", hlsURL)
}
