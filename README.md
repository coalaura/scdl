# scdl

A simple command-line SoundCloud downloader written in Go.

## Installation

```bash
go install github.com/hellsontime/scdl/cmd/scdl@latest
```

Or build from source:

```bash
go build -o scdl cmd/scdl/main.go
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
