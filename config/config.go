package config

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/pflag"
)

// Config holds the command-line configuration and API key
type Config struct {
	Query               string
	ListMode            bool
	FilePath            string
	APIKey              string
	SongMode            bool
	PlaylistID          string
	ConcurrentDownloads int
	SongListMode        bool
	SongList            string
	SongCSVFile         string
	ShowHelp            bool
}

// ParseFlags parses command-line flags and loads the API key from environment
func ParseFlags() *Config {
	var cfg Config

	pflag.StringVarP(&cfg.Query, "query", "d", "", "Download YouTube URL")
	pflag.BoolVarP(&cfg.ListMode, "list", "l", false, "List videos instead of downloading")
	pflag.StringVarP(&cfg.FilePath, "file", "f", "", "Path to file containing queries or URLs")
	pflag.StringVarP(&cfg.PlaylistID, "playlist", "p", "", "YouTube playlist ID to download")
	pflag.IntVarP(&cfg.ConcurrentDownloads, "concurrent", "c", 3, "Number of concurrent downloads")
	pflag.StringVarP(&cfg.SongList, "songs", "m", "", "Comma-separated list of songs to download")
	pflag.StringVar(&cfg.SongCSVFile, "csv-file", "", "Path to CSV file with Artist,Song format")
	pflag.BoolVarP(&cfg.ShowHelp, "help", "h", false, "Show help message")

	var songQuery string
	pflag.StringVarP(&songQuery, "song", "s", "", "Search for a song using 'artist - song name' format")

	pflag.Parse()

	cfg.APIKey = os.Getenv("api_key")
	if cfg.APIKey == "" {
		log.Fatal("YouTube API key not found in environment variables")
	}

	if songQuery != "" {
		cfg.Query = songQuery
		cfg.SongMode = true
	}

	if cfg.SongList != "" {
		cfg.SongListMode = true
	}

	if cfg.SongCSVFile != "" {
		cfg.SongListMode = true
	}

	return &cfg
}

// ShowHelp displays the help message with all available commands and flags
func ShowHelp() {
	fmt.Println("YouTube Audio Downloader")
	fmt.Println("========================")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  ytaudio [flags]")
	fmt.Println()
	fmt.Println("FLAGS:")
	fmt.Println("  -d, --query <url>           Download audio from YouTube URL")
	fmt.Println("  -s, --song <query>          Search and download song using 'artist - song name' format")
	fmt.Println("  -l, --list                  List videos instead of downloading")
	fmt.Println("  -f, --file <path>           Process queries from file (one per line)")
	fmt.Println("  -p, --playlist <id>         Download entire YouTube playlist")
	fmt.Println("  -m, --songs <list>          Download comma-separated list of songs")
	fmt.Println("      --csv-file <path>       Download songs from CSV file (Artist,Song format)")
	fmt.Println("  -c, --concurrent <num>      Number of concurrent downloads (default: 3)")
	fmt.Println("  -h, --help                  Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  ytaudio -d \"https://www.youtube.com/watch?v=dQw4w9WgXcQ\"")
	fmt.Println("  ytaudio -s \"Rick Astley - Never Gonna Give You Up\"")
	fmt.Println("  ytaudio -p \"PLrAXtmRdnEQy4Qy9RMp-3X30f3gWD1CUr\"")
	fmt.Println("  ytaudio -m \"Song 1, Song 2, Song 3\" -c 5")
	fmt.Println("  ytaudio --csv-file songs.csv -c 2")
	fmt.Println("  ytaudio -f queries.txt")
	fmt.Println()
	fmt.Println("ENVIRONMENT:")
	fmt.Println("  api_key                     YouTube Data API key (required)")
}