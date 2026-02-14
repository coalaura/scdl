// Package main is the entry point for the scdl command line tool.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/hellsontime/scdl"
	"github.com/urfave/cli/v2"
)

var version = "dev"

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

			// Ensure output directory exists (0750 per linter requirement)
			if err := os.MkdirAll(outputDir, 0750); err != nil {
				return fmt.Errorf("create output directory: %w", err)
			}

			client, err := scdl.NewClient()
			if err != nil {
				return fmt.Errorf("create client: %w", err)
			}

			track, err := client.GetTrack(trackURL)
			if err != nil {
				return fmt.Errorf("get track: %w", err)
			}

			if _, err := client.Download(track, outputDir, nil); err != nil {
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

// reorderArgs moves flags to the front of the arguments list to support
// <arg> <flag> style usage (e.g. scdl URL --output /tmp).
func reorderArgs(args []string) []string {
	if len(args) < 2 {
		return args
	}

	var newArgs = make([]string, 0, len(args))
	var flags []string
	var positional []string

	// Keep the program name
	newArgs = append(newArgs, args[0])

	// Iterate through the rest
	skipNext := false
	for i, arg := range args[1:] {
		if skipNext {
			skipNext = false
			continue
		}

		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
			// Check if this flag takes an argument
			if (arg == "-o" || arg == "--output") && i+1 < len(args[1:]) {
				// The next argument is the value for this flag
				flags = append(flags, args[1:][i+1])
				skipNext = true
			}
		} else {
			positional = append(positional, arg)
		}
	}

	newArgs = append(newArgs, flags...)
	newArgs = append(newArgs, positional...)

	return newArgs
}
