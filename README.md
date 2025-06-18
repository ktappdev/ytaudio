# ytaudio

A powerful YouTube audio downloader with a Terminal User Interface (TUI), support for playlists, batch processing from CSV or text files, and concurrent downloads.

## Features

-   Interactive TUI for searching and downloading songs.
-   Download audio from YouTube playlists.
-   Batch processing of song queries from CSV files or plain text files.
-   Concurrent downloads for faster processing.
-   High-quality audio extraction using `yt-dlp`.
-   Real-time progress updates in the TUI.
-   Cross-platform support (macOS, Linux, Windows).

## Installation

1.  **Install Go**: Ensure you have Go (version 1.19 or later) installed. You can download it from [golang.org](https://golang.org/dl/).
2.  **Install yt-dlp**: This tool relies on `yt-dlp` for downloading and extracting audio. Install it using a package manager like Homebrew (macOS) or pip (Python):
    ```bash
    # macOS (using Homebrew)
    brew install yt-dlp

    # Other systems (using pip)
    pip install yt-dlp
    ```
3.  **Clone the repository**:
    ```bash
    git clone https://github.com/ktappdev/ytaudio.git
    cd ytaudio
    ```
4.  **Install Go dependencies**:
    ```bash
    go mod tidy
    ```
5.  **Build the application**:
    ```bash
    go build -o ytaudio
    ```

## Setup

You need a YouTube Data API key to use the search functionality:

1.  Go to the [Google Cloud Console](https://console.cloud.google.com/).
2.  Create a new project or select an existing one.
3.  Enable the **YouTube Data API v3**.
4.  Create credentials (API key).
5.  Set the environment variable `api_key` with your API key:
    ```bash
    export api_key="YOUR_YOUTUBE_API_KEY"
    # or
    export youtube_api_key="YOUR_YOUTUBE_API_KEY"
    ```
    The application will first check for `api_key`. If it's not set, it will then check for `youtube_api_key`.
    Alternatively, you can pass the API key directly using the `--api-key` flag (not recommended for security reasons if sharing your command history).

## Usage

### Interactive TUI (Recommended)

To start the application in its interactive TUI mode, simply run:

```bash
./ytaudio
```

Follow the on-screen prompts to search for songs and download them.

### Command-Line Mode

While the TUI is the primary way to interact with `ytaudio`, you can also use command-line flags for specific batch operations.

**Help**

```bash
./ytaudio -h
./ytaudio --help
```

**Download Entire Playlist**

```bash
./ytaudio -p "YOUR_PLAYLIST_ID"
```

**Batch Download from Song List (Comma-separated)**

```bash
./ytaudio -m "Artist1 - Song Title1, Artist2 - Song Title2, Artist3 - Song Title3"
```

**Batch Download from CSV File**

```bash
./ytaudio --csv-file songs.csv
```

The CSV file should contain song queries. The expected format is two columns: `Artist` and `Song`. A header row is optional and will be skipped if present.

*Example `songs.csv` content:*
```csv
Artist,Song
Rick Astley,Never Gonna Give You Up
Queen,Bohemian Rhapsody
The Beatles,Hey Jude
Journey,Don't Stop Believin'
```

If the CSV has only one column, each line will be treated as a full search query.

*Example single-column `songs.csv` content:*
```csv
Rick Astley - Never Gonna Give You Up
Queen - Bohemian Rhapsody
The Beatles - Hey Jude
```

**Process Queries from Text File**

Each line in the text file is treated as a separate search query.

```bash
./ytaudio -f queries.txt
```

*Example `queries.txt` content:*
```
Rick Astley - Never Gonna Give You Up
Queen Bohemian Rhapsody
The Beatles Hey Jude
```

**Concurrent Downloads**

Use the `-c` or `--concurrent` flag to specify the number of concurrent downloads for batch operations (default is 3):

```bash
./ytaudio --csv-file songs.csv -c 5
./ytaudio -p "YOUR_PLAYLIST_ID" -c 2
./ytaudio -m "Song1, Song2" -c 4
```

## Command Line Flags

| Flag         | Short | Description                                                                 |
|--------------|-------|-----------------------------------------------------------------------------|
| `--playlist`   | `-p`  | Download entire YouTube playlist by providing the Playlist ID.              |
| `--songs`      | `-m`  | Download a comma-separated list of songs (e.g., "Artist - Song, ...").    |
| `--csv-file` |       | Download songs from a CSV file. Expected format: `Artist,Song` (header optional) or a single column of search queries. |
| `--file`       | `-f`  | Process search queries from a text file (one query per line).               |
| `--concurrent` | `-c`  | Number of concurrent downloads for batch operations (default: 3).           |
| `--api-key`    |       | Your YouTube Data API v3 key (overrides `api_key` environment variable).    |
| `--help`       | `-h`  | Show this help message.                                                     |

*Note: If no flags are provided, `ytaudio` will launch its interactive TUI.*

## Output

Downloaded audio files are saved as MP3s in the following directory:

-   **Windows**: `%USERPROFILE%\Downloads\YouTubeAudio\`
-   **macOS/Linux**: `~/Downloads/YouTubeAudio/`

Filenames are based on the video title as provided by `yt-dlp`.

## Requirements

-   Go 1.19 or later
-   `yt-dlp` installed and accessible in your system's PATH
-   A YouTube Data API v3 key (for search functionality)
-   Internet connection

## Dependencies

This project uses several Go modules, including:

-   `github.com/charmbracelet/bubbletea` - For the Terminal User Interface
-   `github.com/charmbracelet/bubbles` - UI components for Bubble Tea
-   `github.com/spf13/pflag` - Command line flag parsing
-   `google.golang.org/api/youtube/v3` - Google API Client for YouTube Data API

(yt-dlp is an external dependency, not a Go module)

## License

This project is open-source. Please add your preferred license (e.g., MIT, Apache 2.0).

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues.
