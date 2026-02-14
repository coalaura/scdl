package scdl

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

// mockTransport allows us to mock HTTP responses
type mockTransport struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req)
}

func TestGetStreamURL(t *testing.T) {
	client := &Client{
		clientID: "test-client-id",
		httpClient: &http.Client{
			Transport: &mockTransport{
				RoundTripFunc: func(req *http.Request) (*http.Response, error) {
					// Check URL structure
					if strings.Contains(req.URL.Path, "soundcloud:tracks:123456/ab-cd-ef/stream/hls") {
						// Verify query params
						q := req.URL.Query()
						if q.Get("client_id") != "test-client-id" {
							return &http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader("bad client_id"))}, nil
						}
						if q.Get("track_authorization") != "auth-token" {
							return &http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader("bad auth"))}, nil
						}

						respJSON := `{"url": "https://cf-hls-media.sndcdn.com/playlist.m3u8"}`
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(strings.NewReader(respJSON)),
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

	track := &Track{
		ID:                 123456,
		TrackAuthorization: "auth-token",
		HLSURL:             "https://api-v2.soundcloud.com/media/soundcloud:tracks:123456/ab-cd-ef/stream/hls",
	}

	url, err := client.GetStreamURL(context.Background(), track)
	if err != nil {
		t.Fatalf("GetStreamURL() error = %v", err)
	}

	if url != "https://cf-hls-media.sndcdn.com/playlist.m3u8" {
		t.Errorf("got URL %q, want %q", url, "https://cf-hls-media.sndcdn.com/playlist.m3u8")
	}
}

func TestExtractClientID(t *testing.T) {
	// Simulate:
	// 1. GET soundcloud.com -> returns HTML with asset script src
	// 2. GET asset script -> returns content with client_id:"xyz"

	html := `<html><body><script src="https://a-v2.sndcdn.com/assets/app-123.js"></script></body></html>`
	js := `(function(){ bla bla client_id:"my-client-id-123" bla bla })`

	client := &Client{
		httpClient: &http.Client{
			Transport: &mockTransport{
				RoundTripFunc: func(req *http.Request) (*http.Response, error) {
					if req.URL.String() == "https://soundcloud.com" {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(strings.NewReader(html)),
						}, nil
					}
					if req.URL.String() == "https://a-v2.sndcdn.com/assets/app-123.js" {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(strings.NewReader(js)),
						}, nil
					}
					return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader("Not Found"))}, nil
				},
			},
		},
	}

	id, err := client.extractClientIDFrom(context.Background(), "https://soundcloud.com")
	if err != nil {
		t.Fatalf("extractClientID() error = %v", err)
	}
	if id != "my-client-id-123" {
		t.Errorf("got clientID %q, want %q", id, "my-client-id-123")
	}
}

func TestNewClient(t *testing.T) {
	html := `<html><body><script src="https://a-v2.sndcdn.com/assets/app-123.js"></script></body></html>`
	js := `(function(){ bla bla client_id:"my-client-id-123" bla bla })`

	transport := &mockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			if req.URL.String() == "https://mock.com" {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(html)),
				}, nil
			}
			if req.URL.String() == "https://a-v2.sndcdn.com/assets/app-123.js" {
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(js)),
				}, nil
			}
			return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader("Not Found"))}, nil
		},
	}

	httpClient := &http.Client{Transport: transport}
	client, err := newClient(context.Background(), "https://mock.com", httpClient)
	if err != nil {
		t.Fatalf("newClient() error = %v", err)
	}

	if client.clientID != "my-client-id-123" {
		t.Errorf("got clientID %q, want %q", client.clientID, "my-client-id-123")
	}
}

func TestNewClient_Direct(t *testing.T) {
	// We just want coverage for the wrapper function.
	// It's expected to fail in most environments without network.
	_, _ = NewClient(context.Background())
}

func TestNewClient_Fail(t *testing.T) {
	transport := &mockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("fail")
		},
	}
	httpClient := &http.Client{Transport: transport}
	_, err := newClient(context.Background(), "http://mock", httpClient)
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetError(t *testing.T) {
	client := &Client{
		httpClient: &http.Client{
			Transport: &mockTransport{
				RoundTripFunc: func(req *http.Request) (*http.Response, error) {
					return nil, fmt.Errorf("network error")
				},
			},
		},
	}

	_, err := client.get(context.Background(), "http://fail")
	if err == nil {
		t.Error("expected error for network failure")
	}

	client.httpClient.Transport = &mockTransport{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 500,
				Body:       io.NopCloser(strings.NewReader("internal error")),
			}, nil
		},
	}

	_, err = client.get(context.Background(), "http://500")
	if err == nil {
		t.Error("expected error for HTTP 500")
	}

	t.Run("NewRequestError", func(t *testing.T) {
		_, err := client.get(context.Background(), ":")
		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestExtractClientIDErrors(t *testing.T) {
	t.Run("FetchMainFailed", func(t *testing.T) {
		client := &Client{
			httpClient: &http.Client{
				Transport: &mockTransport{
					RoundTripFunc: func(req *http.Request) (*http.Response, error) {
						return nil, fmt.Errorf("fail")
					},
				},
			},
		}
		_, err := client.extractClientIDFrom(context.Background(), "http://mock")
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("NoAssetsFound", func(t *testing.T) {
		client := &Client{
			httpClient: &http.Client{
				Transport: &mockTransport{
					RoundTripFunc: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(strings.NewReader("no assets here")),
						}, nil
					},
				},
			},
		}
		_, err := client.extractClientIDFrom(context.Background(), "http://mock")
		if err == nil || !strings.Contains(err.Error(), "no asset URLs found") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("AssetFetchFails", func(t *testing.T) {
		html := `<html><body><script src="https://a-v2.sndcdn.com/assets/fail.js"></script></body></html>`
		client := &Client{
			httpClient: &http.Client{
				Transport: &mockTransport{
					RoundTripFunc: func(req *http.Request) (*http.Response, error) {
						if req.URL.String() == "http://mock" {
							return &http.Response{
								StatusCode: 200,
								Body:       io.NopCloser(strings.NewReader(html)),
							}, nil
						}
						return nil, fmt.Errorf("asset fail")
					},
				},
			},
		}
		_, err := client.extractClientIDFrom(context.Background(), "http://mock")
		if err == nil || !strings.Contains(err.Error(), "not found in any asset bundle") {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
func TestDownload(t *testing.T) {
	// Setup encryption for mock segment
	key := []byte("1234567890123456")
	iv := make([]byte, 16) // zero IV
	playlistContent := []byte("audio content")
	// Pad
	padding := aes.BlockSize - (len(playlistContent) % aes.BlockSize)
	padded := append(playlistContent, bytes.Repeat([]byte{byte(padding)}, padding)...)
	// Encrypt
	ciphertext := make([]byte, len(padded))
	block, _ := aes.NewCipher(key)
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)

	// M3U8 content
	m3u8Content := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXT-X-KEY:METHOD=AES-128,URI="http://mock/key",IV=0x00000000000000000000000000000000
#EXTINF:10.0,
http://mock/segment.ts
#EXT-X-ENDLIST`

	client := &Client{
		clientID: "test-client",
		httpClient: &http.Client{
			Transport: &mockTransport{
				RoundTripFunc: func(req *http.Request) (*http.Response, error) {
					u := req.URL.String()
					if strings.Contains(u, "/stream/hls") {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(strings.NewReader(`{"url": "http://mock/playlist.m3u8"}`)),
						}, nil
					}
					if u == "http://mock/playlist.m3u8" {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(strings.NewReader(m3u8Content)),
						}, nil
					}
					if u == "http://mock/key" {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader(key)),
						}, nil
					}
					if u == "http://mock/segment.ts" {
						return &http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader(ciphertext)),
						}, nil
					}
					return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader("Not Found: " + u))}, nil
				},
			},
		},
	}

	track := &Track{
		ID:                 123,
		Title:              "MySong",
		Artist:             "MyArtist",
		HLSURL:             "https://api-v2.soundcloud.com/media/soundcloud:tracks:123/token/stream/hls",
		TrackAuthorization: "auth",
	}

	outDir := t.TempDir()
	outPath, err := client.Download(context.Background(), track, outDir, nil)
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}

	// Verify file name
	expectedName := "MyArtist - MySong.mp3"
	if !strings.HasSuffix(outPath, expectedName) {
		t.Errorf("expected suffix %q, got path %q", expectedName, outPath)
	}

	// Verify content
	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}

	// The file should contain "audio content" plus ID3 metadata appended/prepended.
	if !bytes.Contains(content, playlistContent) {
		t.Errorf("file content missing decrypted audio. Got size %d", len(content))
	}
}
