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
			&cli.StringFlag{
				Name:    "author",
				Aliases: []string{"a"},
				Usage:   "Override artist/author name",
			},
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "Override track title",
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

			if author := c.String("author"); author != "" {
				track.Artist = author
			}
			if name := c.String("name"); name != "" {
				track.Title = name
			}

			if _, err := client.Download(ctx, track, outputDir, nil); err != nil {
				return fmt.Errorf("download: %w", err)
			}

			fmt.Printf("downloaded: %s - %s\n", track.Artist, track.Title)
			return nil
		},
	}

	if err := app.Run(reorderArgs(os.Args)); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// reorderArgs moves flags and their values before positional arguments so that
// urfave/cli parses them correctly regardless of where the user places them.
func reorderArgs(args []string) []string {
	if len(args) < 2 {
		return args
	}

	flags := []string{args[0]}
	var positional []string

	for i := 1; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			flags = append(flags, args[i])
			// If the flag doesn't use --flag=value form and has a next arg, treat it as the flag's value.
			if !strings.Contains(args[i], "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				flags = append(flags, args[i])
			}
		} else {
			positional = append(positional, args[i])
		}
	}

	return append(flags, positional...)
}
