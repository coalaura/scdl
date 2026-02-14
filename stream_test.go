package scdl

import (
	"testing"
)

func TestParseHLSURL(t *testing.T) {
	tests := []struct {
		name            string
		url             string
		wantTrackID     string
		wantStreamToken string
		wantErr         bool
	}{
		{
			name:            "Valid URL",
			url:             "https://api-v2.soundcloud.com/media/soundcloud:tracks:123456/ab-cd-ef/stream/hls",
			wantTrackID:     "123456",
			wantStreamToken: "ab-cd-ef",
			wantErr:         false,
		},
		{
			name:            "Valid URL with query params",
			url:             "https://api-v2.soundcloud.com/media/soundcloud:tracks:987654/xyz/stream/hls?foo=bar",
			wantTrackID:     "987654",
			wantStreamToken: "xyz",
			wantErr:         false,
		},
		{
			name:    "Invalid URL - too short",
			url:     "https://api-v2.soundcloud.com/media",
			wantErr: true,
		},
		{
			name:    "Invalid URL - missing track ID",
			url:     "https://api-v2.soundcloud.com/media/something/ab-cd-ef/stream/hls",
			wantErr: true,
		},
		{
			name:    "Invalid URL - bad track format",
			url:     "https://api-v2.soundcloud.com/media/soundcloud:tracks/ab-cd-ef/stream/hls",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTrackID, gotStreamToken, err := parseHLSURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseHLSURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotTrackID != tt.wantTrackID {
					t.Errorf("parseHLSURL() gotTrackID = %v, want %v", gotTrackID, tt.wantTrackID)
				}
				if gotStreamToken != tt.wantStreamToken {
					t.Errorf("parseHLSURL() gotStreamToken = %v, want %v", gotStreamToken, tt.wantStreamToken)
				}
			}
		})
	}
}
