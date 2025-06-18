package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
	SongListMode        bool
	SongList            string
	SongCSVFile         string
	ShowHelp            bool
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
	log.Println("YouTube API key loaded from environment variables")

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

	return cfg
}

// showHelp displays the help message with all available commands and flags
func showHelp() {
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

// run executes the main program logic based on the provided configuration
func run(cfg Config) error {
	log.Println("Starting main program execution")
	
	// Check if help flag is set or no command is provided
	if cfg.ShowHelp {
		showHelp()
		return nil
	}
	
	// Check if no command is provided
	if cfg.Query == "" && cfg.FilePath == "" && cfg.PlaylistID == "" && !cfg.SongListMode {
		showHelp()
		return nil
	}
	
	switch {
	case cfg.PlaylistID != "":
		log.Printf("Downloading playlist: %s", cfg.PlaylistID)
		return downloadPlaylist(cfg)
	case cfg.SongListMode:
		if cfg.SongCSVFile != "" {
			log.Printf("Downloading songs from CSV file: %s", cfg.SongCSVFile)
		} else {
			log.Printf("Downloading song list: %s", cfg.SongList)
		}
		return downloadSongList(cfg)
	case cfg.FilePath != "":
		log.Printf("Processing file: %s", cfg.FilePath)
		return processFile(cfg)
	case cfg.Query == "":
		log.Println("No query provided")
		showHelp()
		return nil
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

// downloadSongList downloads multiple songs from a comma-separated list or CSV file with concurrency
func downloadSongList(cfg Config) error {
	log.Printf("Parsing song list with %d concurrent downloads", cfg.ConcurrentDownloads)
	
	var cleanSongs []string
	var err error
	
	if cfg.SongCSVFile != "" {
		// Read songs from CSV file
		cleanSongs, err = readSongsFromCSV(cfg.SongCSVFile)
		if err != nil {
			return fmt.Errorf("error reading CSV file: %w", err)
		}
	} else {
		// Split the comma-separated list and clean up each song
		songs := strings.Split(cfg.SongList, ",")
		for _, song := range songs {
			song = strings.TrimSpace(song)
			if song != "" {
				cleanSongs = append(cleanSongs, song)
			}
		}
	}
	
	if len(cleanSongs) == 0 {
		return fmt.Errorf("no valid songs found in the list")
	}
	
	log.Printf("Found %d songs to download", len(cleanSongs))
	
	// Create channels for job distribution
	jobs := make(chan string, len(cleanSongs))
	results := make(chan error, len(cleanSongs))
	
	// Start worker goroutines
	var wg sync.WaitGroup
	for w := 1; w <= cfg.ConcurrentDownloads; w++ {
		wg.Add(1)
		go songWorker(jobs, results, &wg, cfg.APIKey)
	}
	
	// Send jobs
	for _, song := range cleanSongs {
		jobs <- song
	}
	close(jobs)
	
	// Wait for all workers to finish
	wg.Wait()
	close(results)
	
	// Collect and report results
	var errors []error
	for err := range results {
		if err != nil {
			log.Printf("Error downloading song: %v", err)
			errors = append(errors, err)
		}
	}
	
	log.Printf("Completed downloading %d songs with %d errors", len(cleanSongs), len(errors))
	
	if len(errors) > 0 {
		return fmt.Errorf("encountered %d errors during download", len(errors))
	}
	
	return nil
}

// songWorker processes individual songs from the job queue
func songWorker(jobs <-chan string, results chan<- error, wg *sync.WaitGroup, apiKey string) {
	defer wg.Done()
	for song := range jobs {
		log.Printf("Processing song: %s", song)
		
		// Search for the song
		videos, err := searchVideos(song+" audio", apiKey)
		if err != nil {
			log.Printf("Error searching for '%s': %v", song, err)
			results <- fmt.Errorf("search failed for '%s': %w", song, err)
			continue
		}
		
		if len(videos) == 0 {
			log.Printf("No videos found for song: %s", song)
			results <- fmt.Errorf("no videos found for '%s'", song)
			continue
		}
		
		// Download the first result
		log.Printf("Downloading first result for '%s': %s", song, videos[0].Title)
		err = downloadAudio(videos[0].ID)
		if err != nil {
			log.Printf("Error downloading '%s': %v", song, err)
			results <- fmt.Errorf("download failed for '%s': %w", song, err)
		} else {
			log.Printf("Successfully downloaded: %s", song)
			results <- nil
		}
	}
}

// readSongsFromCSV reads songs from a CSV file with Artist,Song format
func readSongsFromCSV(filePath string) ([]string, error) {
	log.Printf("Reading songs from CSV file: %s", filePath)
	
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening CSV file: %w", err)
	}
	defer file.Close()
	
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV file: %w", err)
	}
	
	var songs []string
	for i, record := range records {
		// Skip header row if it exists
		if i == 0 && len(record) >= 2 && (strings.ToLower(record[0]) == "artist" || strings.ToLower(record[1]) == "song") {
			log.Println("Skipping header row")
			continue
		}
		
		if len(record) >= 2 {
			artist := strings.TrimSpace(record[0])
			song := strings.TrimSpace(record[1])
			if artist != "" && song != "" {
				songQuery := fmt.Sprintf("%s - %s", artist, song)
				songs = append(songs, songQuery)
				log.Printf("Added song: %s", songQuery)
			}
		}
	}
	
	log.Printf("Successfully read %d songs from CSV file", len(songs))
	return songs, nil
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
