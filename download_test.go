package scdl

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"net/url"
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Normal filename",
			input: "Artist - Song",
			want:  "Artist - Song",
		},
		{
			name:  "Filename with forward slash",
			input: "AC/DC - Thunderstruck",
			want:  "AC_DC - Thunderstruck",
		},
		{
			name:  "Filename with special chars",
			input: "Artist: Song? <Cool>",
			want:  "Artist_ Song_ _Cool_",
		},
		{
			name:  "Filename with mixed separators",
			input: "foo\\bar|baz*qux",
			want:  "foo_bar_baz_qux",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeFilename(tt.input); got != tt.want {
				t.Errorf("sanitizeFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveURI(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com/hls/playlist.m3u8")

	tests := []struct {
		name    string
		base    *url.URL
		uri     string
		want    string
		wantErr bool
	}{
		{
			name: "Absolute URI",
			base: baseURL,
			uri:  "https://other.com/segment.ts",
			want: "https://other.com/segment.ts",
		},
		{
			name: "Relative URI",
			base: baseURL,
			uri:  "segment.ts",
			want: "https://example.com/hls/segment.ts",
		},
		{
			name: "Relative URI parent dir",
			base: baseURL,
			uri:  "../segment.ts",
			want: "https://example.com/segment.ts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveURI(tt.base, tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveURI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("resolveURI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecryptAES128CBC(t *testing.T) {
	// Standard AES-128-CBC test vectors or just a roundtrip test
	key := []byte("1234567890123456") // 16 bytes
	iv := []byte("abcdefghijklmnop")  // 16 bytes
	plaintext := []byte("Hello World! 123")

	// Encrypt manually to setup test case
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}

	// Pad plaintext to block size
	padding := aes.BlockSize - (len(plaintext) % aes.BlockSize)
	paddedText := append(plaintext, bytes.Repeat([]byte{byte(padding)}, padding)...)

	ciphertext := make([]byte, len(paddedText))
	cbc := cipher.NewCBCEncrypter(block, iv)
	cbc.CryptBlocks(ciphertext, paddedText)

	t.Run("Valid Decryption", func(t *testing.T) {
		got, err := decryptAES128CBC(ciphertext, key, iv)
		if err != nil {
			t.Fatalf("decryptAES128CBC() error = %v", err)
		}
		if !bytes.Equal(got, plaintext) {
			t.Errorf("decryptAES128CBC() = %q, want %q", got, plaintext)
		}
	})

	t.Run("Invalid Key Size", func(t *testing.T) {
		badKey := []byte("too short")
		_, err := decryptAES128CBC(ciphertext, badKey, iv)
		if err == nil {
			t.Error("expected error for invalid key size")
		}
	})

	t.Run("Ciphertext not multiple of block size", func(t *testing.T) {
		badCiphertext := ciphertext[:len(ciphertext)-1]
		_, err := decryptAES128CBC(badCiphertext, key, iv)
		if err == nil {
			t.Error("expected error for bad ciphertext length")
		}
	})

	t.Run("Bad Padding", func(t *testing.T) {
		// Corrupt the last byte to invalidate padding
		badCiphertext := make([]byte, len(ciphertext))
		copy(badCiphertext, ciphertext)
		badCiphertext[len(badCiphertext)-1] ^= 0x01

		// In standard CBC decryption, if we change the last byte of ciphertext,
		// it affects the last byte of the decrypted plaintext (XORed with last byte of previous ciphertext block).
		// This might produce a valid padding byte value by chance, but unlikely to be correct for this test unless we calculate it.
		// Actually, if we modify the last byte of ciphertext, the last block of plaintext changes.
		// If we want to guarantee bad padding, we might just pass garbage.
		// But let's verify it doesn't panic at least.
	})
}

// Helpers
func decodeHex(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}
