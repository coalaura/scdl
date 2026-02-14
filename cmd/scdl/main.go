// Package main is the entry point for the scdl command line tool.
package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/hellsontime/scdl"
	"github.com/urfave/cli/v2"
)

var version = "dev"

func init() {
	if version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok {
			if info.Main.Version != "" && info.Main.Version != "(devel)" {
				version = info.Main.Version
			}
		}
	}
}

func main() {
	app := &cli.App{
		Name:    "scdl",
		Version: version,
		Usage:   "SoundCloud Downloader",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output directory",
				Value:   ".",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf("missing SoundCloud URL")
			}

			trackURL := c.Args().First()
			if !strings.Contains(trackURL, "soundcloud.com/") {
				return fmt.Errorf("not a valid SoundCloud URL: %s", trackURL)
			}

			outputDir := c.String("output")

			if err := os.MkdirAll(outputDir, 0750); err != nil {
				return fmt.Errorf("create output directory: %w", err)
			}

			ctx := c.Context

			client, err := scdl.NewClient(ctx)
			if err != nil {
				return fmt.Errorf("create client: %w", err)
			}

			track, err := client.GetTrack(ctx, trackURL)
			if err != nil {
				return fmt.Errorf("get track: %w", err)
			}

			if _, err := client.Download(ctx, track, outputDir, nil); err != nil {
				return fmt.Errorf("download: %w", err)
			}

			fmt.Printf("downloaded: %s - %s\n", track.Artist, track.Title)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
