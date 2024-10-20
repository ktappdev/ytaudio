package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kkdai/youtube/v2"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/pflag"
)

const (
	youtubeAPIURL = "https://www.googleapis.com/youtube/v3/search"
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
}

// Video represents a YouTube video with its ID and Title
type Video struct {
	ID    string
	Title string
}

func main() {
	// Set up logging to include timestamps
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Println("Starting YouTube audio downloader")

	cfg := parseFlags()
	log.Printf("Parsed configuration: %+v", cfg)

	if err := run(cfg); err != nil {
		log.Fatalf("Error: %v", err)
	}

	log.Println("Program completed successfully")
}

// parseFlags parses command-line flags and loads the API key from environment
func parseFlags() Config {
	var cfg Config

	pflag.StringVarP(&cfg.Query, "query", "q", "", "Search query or YouTube URL")
	pflag.BoolVarP(&cfg.ListMode, "list", "l", false, "List videos instead of downloading")
	pflag.StringVarP(&cfg.FilePath, "file", "f", "", "Path to file containing queries or URLs")
	pflag.StringVarP(&cfg.PlaylistID, "playlist", "p", "", "YouTube playlist ID to download")
	pflag.IntVarP(&cfg.ConcurrentDownloads, "concurrent", "c", 3, "Number of concurrent downloads")

	var songQuery string
	pflag.StringVarP(&songQuery, "song", "s", "", "Search for a song using 'artist - song name' format")

	pflag.Parse()

	cfg.APIKey = os.Getenv("api_key")
	if cfg.APIKey == "" {
		log.Fatal("YouTube API key not found in environment variables")
	}
	log.Println("YouTube API key loaded from environment variables")

	if songQuery != "" {
		cfg.Query = songQuery
		cfg.SongMode = true
	}

	return cfg
}

// run executes the main program logic based on the provided configuration
func run(cfg Config) error {
	log.Println("Starting main program execution")
	switch {
	case cfg.PlaylistID != "":
		log.Printf("Downloading playlist: %s", cfg.PlaylistID)
		return downloadPlaylist(cfg)
	case cfg.FilePath != "":
		log.Printf("Processing file: %s", cfg.FilePath)
		return processFile(cfg)
	case cfg.Query == "":
		log.Println("No query provided")
		return fmt.Errorf("no query provided")
	case cfg.ListMode:
		log.Printf("Listing videos for query: %s", cfg.Query)
		return listVideos(cfg)
	case cfg.SongMode:
		log.Printf("Searching and downloading song: %s", cfg.Query)
		return searchAndDownloadSong(cfg)
	default:
		log.Printf("Downloading audio for query: %s", cfg.Query)
		return downloadAudio(cfg.Query)
	}
}

// processFile reads queries from a file and processes each one
func processFile(cfg Config) error {
	log.Printf("Reading file: %s", cfg.FilePath)
	content, err := os.ReadFile(cfg.FilePath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	queries := strings.Split(string(content), "\n")
	log.Printf("Found %d queries in file", len(queries))

	for i, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" {
			log.Printf("Skipping empty query at line %d", i+1)
			continue
		}
		log.Printf("Processing query %d: %s", i+1, query)
		videos, err := searchVideos(query, cfg.APIKey)
		if err != nil {
			log.Printf("Error searching for '%s': %v", query, err)
			continue
		}
		if len(videos) > 0 {
			log.Printf("Found %d videos for query '%s', downloading first result", len(videos), query)
			if err := downloadAudio(videos[0].ID); err != nil {
				log.Printf("Error processing '%s': %v", query, err)
			}
		} else {
			log.Printf("No videos found for query: %s", query)
		}
	}

	return nil
}

// listVideos searches for videos and displays the results
func listVideos(cfg Config) error {
	log.Printf("Searching for videos with query: %s", cfg.Query)
	videos, err := searchVideos(cfg.Query, cfg.APIKey)
	if err != nil {
		return fmt.Errorf("error searching videos: %w", err)
	}

	if len(videos) == 0 {
		log.Println("No videos found")
		fmt.Println("No videos found.")
		return nil
	}

	log.Printf("Found %d videos", len(videos))
	for i, video := range videos {
		log.Printf("Displaying video %d: %s", i+1, video.Title)
		fmt.Printf("Title: %s\nID: %s\nURL: https://www.youtube.com/watch?v=%s\n\n",
			video.Title, video.ID, video.ID)
	}

	return nil
}

// searchAndDownloadSong searches for a song and downloads the first result
func searchAndDownloadSong(cfg Config) error {
	log.Printf("Searching for song: %s", cfg.Query)
	videos, err := searchVideos(cfg.Query+" audio", cfg.APIKey)
	if err != nil {
		return fmt.Errorf("error searching for song: %w", err)
	}

	if len(videos) == 0 {
		log.Println("No videos found for the song")
		return fmt.Errorf("no videos found for the song")
	}

	log.Printf("Found %d videos, downloading the first result", len(videos))
	return downloadAudio(videos[0].ID)
}

// searchVideos performs a YouTube search using the YouTube Data API
func searchVideos(query string, apiKey string) ([]Video, error) {
	log.Printf("Searching YouTube for: %s", query)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	searchURL := fmt.Sprintf("%s?part=snippet&q=%s&key=%s&type=video&maxResults=5",
		youtubeAPIURL, url.QueryEscape(query), apiKey)
	log.Printf("Search URL: %s", searchURL)

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	client := &http.Client{}
	log.Println("Sending HTTP request to YouTube API")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	log.Println("Reading response body")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var searchResponse struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			} `json:"id"`
			Snippet struct {
				Title string `json:"title"`
			} `json:"snippet"`
		} `json:"items"`
	}

	log.Println("Parsing JSON response")
	err = json.Unmarshal(body, &searchResponse)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON response: %w", err)
	}

	var videos []Video
	for _, item := range searchResponse.Items {
		video := Video{
			ID:    item.ID.VideoID,
			Title: item.Snippet.Title,
		}
		videos = append(videos, video)
		log.Printf("Found video: %s (ID: %s)", video.Title, video.ID)
	}

	log.Printf("Found %d videos in total", len(videos))
	return videos, nil
}

func downloadAudio(query string) error {
	log.Printf("Initializing download for query: %s", query)
	client := youtube.Client{}

	log.Println("Fetching video information")
	video, err := client.GetVideo(query)
	if err != nil {
		return fmt.Errorf("error getting video info: %w", err)
	}
	log.Printf("Video information fetched for: %s", video.Title)

	// Find the audio format with the highest bitrate
	var format *youtube.Format
	maxBitrate := 0
	for _, f := range video.Formats.WithAudioChannels() {
		if f.AudioQuality != "" && f.AverageBitrate > maxBitrate {
			maxBitrate = f.AverageBitrate
			format = &f
		}
	}

	if format == nil {
		return fmt.Errorf("no suitable audio format found")
	}

	log.Printf("Selected format: Audio Quality: %s, Mime Type: %s, Bitrate: %d",
		format.AudioQuality, format.MimeType, format.AverageBitrate)

	log.Println("Getting stream")
	stream, size, err := client.GetStream(video, format)
	if err != nil {
		return fmt.Errorf("error getting stream: %w", err)
	}
	defer stream.Close()

	fileName := sanitizeFileName(fmt.Sprintf("%s.mp3", video.Title))
	filePath := filepath.Join(getDownloadPath(), fileName)
	log.Printf("Saving audio to: %s", filePath)

	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer out.Close()

	bar := progressbar.DefaultBytes(
		size,
		"Downloading",
	)

	log.Println("Copying audio data to file")
	startTime := time.Now()
	written, err := io.Copy(io.MultiWriter(out, bar), stream)
	if err != nil {
		return fmt.Errorf("error saving file: %w", err)
	}

	duration := time.Since(startTime)
	speed := float64(written) / duration.Seconds() / 1024 // KB/s

	log.Println("Download completed successfully")
	fmt.Printf("\nDownloaded: %s\n", filePath)
	fmt.Printf("Download speed: %.2f KB/s\n", speed)
	return nil
}

func downloadPlaylist(cfg Config) error {
	downloader := NewPlaylistDownloader(cfg.APIKey, cfg.ConcurrentDownloads, downloadAudio)
	return downloader.DownloadPlaylist(cfg.PlaylistID)
}

// getDownloadPath returns the path to save downloaded files
func getDownloadPath() string {
	log.Println("Determining download path")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting user home directory: %v", err)
	}

	downloadPath := filepath.Join(homeDir, "Downloads", "YouTubeAudio")
	log.Printf("Download path: %s", downloadPath)

	if err := os.MkdirAll(downloadPath, 0755); err != nil {
		log.Fatalf("Error creating download directory: %v", err)
	}

	return downloadPath
}

// sanitizeFileName removes or replaces characters that are invalid in file names
func sanitizeFileName(fileName string) string {
	log.Printf("Sanitizing file name: %s", fileName)
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalid {
		fileName = strings.ReplaceAll(fileName, char, "_")
	}
	log.Printf("Sanitized file name: %s", fileName)
	return fileName
}
