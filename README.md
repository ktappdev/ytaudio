# ytaudio

A powerful YouTube audio downloader with support for single videos, playlists, batch processing, and concurrent downloads.

## Features

- Download audio from YouTube videos and playlists
- Search and download songs by artist and title
- Batch processing from CSV files or text files
- Concurrent downloads for faster processing
- High-quality audio extraction
- Progress bars and detailed logging
- Cross-platform support

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd ytaudio
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
go build -o ytaudio
```

## Setup

You need a YouTube Data API key to use this tool:

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the YouTube Data API v3
4. Create credentials (API key)
5. Set the environment variable:

```bash
export api_key="your-youtube-api-key-here"
```

## Usage

### Help
```bash
./ytaudio -h
./ytaudio --help
./ytaudio          # Shows help when no arguments provided
```

### Download Single Video
```bash
./ytaudio -d "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
```

### Search and Download Song
```bash
./ytaudio -s "Rick Astley - Never Gonna Give You Up"
```

### List Videos (without downloading)
```bash
./ytaudio -l -d "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
./ytaudio -l -s "Rick Astley - Never Gonna Give You Up"
```

### Download Entire Playlist
```bash
./ytaudio -p "PLrAXtmRdnEQy4Qy9RMp-3X30f3gWD1CUr"
```

### Batch Download from List
```bash
./ytaudio -m "Song 1, Song 2, Song 3"
```

### Batch Download from CSV File
```bash
./ytaudio --csv-file songs.csv
```

CSV file format (with optional header):
```csv
Artist,Song
Rick Astley,Never Gonna Give You Up
Queen,Bohemian Rhapsody
The Beatles,Hey Jude
```

### Process Queries from Text File
```bash
./ytaudio -f queries.txt
```

Text file format (one query per line):
```
Rick Astley - Never Gonna Give You Up
https://www.youtube.com/watch?v=dQw4w9WgXcQ
Queen Bohemian Rhapsody
```

### Concurrent Downloads
Use the `-c` flag to specify the number of concurrent downloads (default: 3):

```bash
./ytaudio --csv-file songs.csv -c 5
./ytaudio -p "playlist-id" -c 2
./ytaudio -m "Song 1, Song 2, Song 3" -c 4
```

## Command Line Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--query` | `-d` | Download audio from YouTube URL |
| `--song` | `-s` | Search and download song using 'artist - song name' format |
| `--list` | `-l` | List videos instead of downloading |
| `--file` | `-f` | Process queries from file (one per line) |
| `--playlist` | `-p` | Download entire YouTube playlist |
| `--songs` | `-m` | Download comma-separated list of songs |
| `--csv-file` | | Download songs from CSV file (Artist,Song format) |
| `--concurrent` | `-c` | Number of concurrent downloads (default: 3) |
| `--help` | `-h` | Show help message |

## Examples

### Basic Usage
```bash
# Download single video
./ytaudio -d "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

# Search and download
./ytaudio -s "Imagine Dragons - Believer"

# List search results
./ytaudio -l -s "Coldplay - Yellow"
```

### Batch Processing
```bash
# From CSV file with 2 concurrent downloads
./ytaudio --csv-file my-playlist.csv -c 2

# From comma-separated list
./ytaudio -m "Song 1, Song 2, Song 3" -c 5

# From text file
./ytaudio -f my-queries.txt
```

### Playlist Downloads
```bash
# Download entire playlist with 4 concurrent downloads
./ytaudio -p "PLrAXtmRdnEQy4Qy9RMp-3X30f3gWD1CUr" -c 4
```

## Output

Downloaded files are saved to:
- **Windows**: `%USERPROFILE%\Downloads\YouTubeAudio\`
- **macOS/Linux**: `~/Downloads/YouTubeAudio/`

Files are saved as MP3 format with the video title as the filename (sanitized for filesystem compatibility).

## Requirements

- Go 1.19 or later
- YouTube Data API v3 key
- Internet connection

## Dependencies

- `github.com/kkdai/youtube/v2` - YouTube video extraction
- `github.com/schollz/progressbar/v3` - Progress bars
- `github.com/spf13/pflag` - Command line flag parsing
- `google.golang.org/api/youtube/v3` - YouTube Data API

## License

[Add your license information here]

## Contributing

[Add contributing guidelines here]
