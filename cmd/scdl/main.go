// Package main provides the command line interface for the downloader.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/hellsontime/scdl/pkg/soundcloud"
)

func main() {
	outputDir := flag.String("o", ".", "Output directory")
	flag.StringVar(outputDir, "output", ".", "Output directory")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: scdl [options] <soundcloud-url>\n\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	trackURL := flag.Arg(0)
	if !strings.Contains(trackURL, "soundcloud.com/") {
		fmt.Fprintf(os.Stderr, "Error: not a valid SoundCloud URL: %s\n", trackURL)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Resolving client ID...\n")
	client, err := soundcloud.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Fetching track info...\n")
	track, err := client.GetTrack(trackURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Downloading: %s - %s\n", track.Artist, track.Title)

	progress := func(downloaded, total int) {
		fmt.Fprintf(os.Stderr, "\rDownloading segments: %d/%d", downloaded, total)
	}

	outPath, err := client.Download(track, *outputDir, progress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "\nDone: %s\n", outPath)
}
