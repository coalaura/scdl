package scdl

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestGetTrack(t *testing.T) {
	// Sample hydration data (minified)
	hydrationData := `[{"hydratable":"user","data":{}},{"hydratable":"sound","data":{"id":123456,"title":"Test Title","description":"Test Description","genre":"Rock","duration":60000,"artwork_url":"https://i1.sndcdn.com/artworks-000-large.jpg","track_authorization":"auth-token","user":{"username":"Test Artist"},"media":{"transcodings":[{"url":"https://api-v2.soundcloud.com/media/soundcloud:tracks:123456/stream/hls","format":{"protocol":"hls","mime_type":"audio/mpeg"}},{"url":"https://api-v2.soundcloud.com/media/soundcloud:tracks:123456/stream/progressive","format":{"protocol":"progressive","mime_type":"audio/mpeg"}}]}}}]`

	htmlResponse := `
	<html>
	<body>
	<script>
	window.__sc_hydration = ` + hydrationData + `;
	</script>
	</body>
	</html>
	`

	client := &Client{
		httpClient: &http.Client{
			Transport: &mockTransport{
				RoundTripFunc: func(req *http.Request) (*http.Response, error) {
					if req.URL.String() == "https://soundcloud.com/artist/song" {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(strings.NewReader(htmlResponse)),
							Header:     make(http.Header),
						}, nil
					}
					return &http.Response{
						StatusCode: 404,
						Body:       io.NopCloser(strings.NewReader("Not Found")),
					}, nil
				},
			},
		},
	}

	track, err := client.GetTrack("https://soundcloud.com/artist/song")
	if err != nil {
		t.Fatalf("GetTrack() error = %v", err)
	}

	if track.ID != 123456 {
		t.Errorf("got ID %d, want 123456", track.ID)
	}
	if track.Title != "Test Title" {
		t.Errorf("got Title %q, want %q", track.Title, "Test Title")
	}
	if track.Artist != "Test Artist" {
		t.Errorf("got Artist %q, want %q", track.Artist, "Test Artist")
	}
	if track.HLSURL != "https://api-v2.soundcloud.com/media/soundcloud:tracks:123456/stream/hls" {
		t.Errorf("got HLSURL %q, want %q", track.HLSURL, "https://api-v2.soundcloud.com/media/soundcloud:tracks:123456/stream/hls")
	}
}
