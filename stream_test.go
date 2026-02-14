package scdl

import (
	"fmt"
	"io"
	"net/http"
	"strings"
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
func TestGetStreamURL_Errors(t *testing.T) {
	client := &Client{
		clientID: "test",
		httpClient: &http.Client{
			Transport: &mockTransport{
				RoundTripFunc: func(req *http.Request) (*http.Response, error) {
					u := req.URL.String()
					if strings.Contains(u, "fail") {
						return nil, fmt.Errorf("fail")
					}
					if strings.Contains(u, "empty") {
						return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"url": ""}`))}, nil
					}
					if strings.Contains(u, "badjson") {
						return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{invalid`))}, nil
					}
					return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader("404"))}, nil
				},
			},
		},
	}

	t.Run("ParseFail", func(t *testing.T) {
		_, err := client.GetStreamURL(&Track{HLSURL: "invalid"})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("FetchFail", func(t *testing.T) {
		_, err := client.GetStreamURL(&Track{HLSURL: "http://api/soundcloud:tracks:123/fail/stream/hls"})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("EmptyURL", func(t *testing.T) {
		_, err := client.GetStreamURL(&Track{HLSURL: "http://api/soundcloud:tracks:123/empty/stream/hls"})
		if err == nil || !strings.Contains(err.Error(), "empty stream URL") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("BadJSON", func(t *testing.T) {
		_, err := client.GetStreamURL(&Track{HLSURL: "http://api/soundcloud:tracks:123/badjson/stream/hls"})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("BadTrackID", func(t *testing.T) {
		_, _, err := parseHLSURL("https://api-v2.soundcloud.com/media/soundcloud:tracks/ab-cd-ef/stream/hls")
		if err == nil {
			t.Error("expected error")
		}
	})
}
