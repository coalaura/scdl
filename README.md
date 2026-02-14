# scdl

A simple command-line & library downloader for SoundCloud tracks written in Go.

## Installation

```bash
go install github.com/hellsontime/scdl/cmd/scdl@latest
```

Or build from source:

```bash
go build -o scdl cmd/scdl/main.go
```
or 
```bash
make build-os-arch
```

## Usage

```bash
scdl <soundcloud-url> [options]
```

### Options

- `-o, --output`: Specify the output directory (defaults to current directory).

## Examples

Download to the current directory:

```bash
scdl https://soundcloud.com/cowboyclicker/stay
```

Download to a specific directory:

```bash
scdl https://soundcloud.com/cowboyclicker/stay --output ~/Music/
```

## Library Usage

### Installation

```bash
go get github.com/hellsontime/scdl
```

### Simple Usage

```go
import "github.com/hellsontime/scdl"

// ...

client, err := scdl.NewClient()
if err != nil {
    return err
}

// Fetch track metadata
track, err := client.GetTrack("https://soundcloud.com/cowboyclicker/stay")
if err != nil {
    return err
}

// Download the track to the current directory
path, err := client.Download(track, ".", nil) // progress callback is optional
if err != nil {
    return err
}
```
