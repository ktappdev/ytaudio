package downloader

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ktappdev/ytaudio/config"
	"github.com/ktappdev/ytaudio/youtube"
)

// ProcessFile reads queries from a file and processes each one
func ProcessFile(cfg *config.Config) error {
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
		videos, err := youtube.SearchVideos(query, cfg.APIKey)
		if err != nil {
			log.Printf("Error searching for '%s': %v", query, err)
			continue
		}
		if len(videos) > 0 {
			log.Printf("Found %d videos for query '%s', downloading first result", len(videos), query)
			if err := DownloadAudio(videos[0].ID); err != nil {
				log.Printf("Error processing '%s': %v", query, err)
			}
		} else {
			log.Printf("No videos found for query: %s", query)
		}
	}

	return nil
}

// checkYtDlpInstalled verifies that yt-dlp is available on the system
func checkYtDlpInstalled() error {
	cmd := exec.Command("yt-dlp", "--version")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("yt-dlp not found. Please install it with: brew install yt-dlp")
	}
	return nil
}

// DownloadAudio downloads audio using yt-dlp (much more reliable than the Go library)
func DownloadAudio(videoID string) error {
	log.Printf("Initializing yt-dlp download for video ID: %s", videoID)

	// Check if yt-dlp is installed
	if err := checkYtDlpInstalled(); err != nil {
		return err
	}

	// Construct YouTube URL from video ID
	videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
	downloadPath := getDownloadPath()

	log.Printf("Downloading from: %s", videoURL)
	log.Printf("Download path: %s", downloadPath)

	// yt-dlp command with options for audio-only download (more efficient)
	cmd := exec.Command("yt-dlp",
		"-f", "bestaudio", // Download only audio stream (more efficient)
		"--extract-audio", // Extract audio only
		"--audio-format", "mp3", // Convert to MP3
		"--audio-quality", "0", // Best quality
		"--output", filepath.Join(downloadPath, "%(title)s.%(ext)s"), // Output template
		"--no-playlist", // Don't download playlists
		"--embed-metadata", // Embed metadata
		"--add-metadata", // Add metadata
		videoURL,
	)

	// Create a pipe to capture output for progress monitoring
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error creating stderr pipe: %w", err)
	}

	// Start the command
	log.Println("Starting yt-dlp download...")
	startTime := time.Now()
	if startErr := cmd.Start(); startErr != nil {
		return fmt.Errorf("error starting yt-dlp: %w", err)
	}

	// Monitor progress from stderr (yt-dlp outputs progress to stderr)
	go func() {
		scanner := bufio.NewScanner(stderr)
		progressRegex := regexp.MustCompile(`\[(\d+\.\d+)%\]`)

		for scanner.Scan() {
			line := scanner.Text()
			log.Printf("yt-dlp: %s", line)

			// Extract progress percentage
			if matches := progressRegex.FindStringSubmatch(line); len(matches) > 1 {
				if progress, parseErr := strconv.ParseFloat(matches[1], 64); parseErr == nil {
					fmt.Printf("\rProgress: %.1f%%", progress)
				}
			}
		}
	}()

	// Also capture stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			log.Printf("yt-dlp stdout: %s", scanner.Text())
		}
	}()

	// Wait for the command to complete
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("yt-dlp download failed: %w", err)
	}

	duration := time.Since(startTime)
	log.Printf("Download completed successfully in %v", duration)
	fmt.Printf("\nDownload completed in %v\n", duration)
	fmt.Printf("Files saved to: %s\n", downloadPath)

	return nil
}

// DownloadSongList downloads multiple songs from a comma-separated list or CSV file with concurrency
func DownloadSongList(cfg *config.Config) error {
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
		videos, err := youtube.SearchVideos(song+" audio", apiKey)
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
		err = DownloadAudio(videos[0].ID)
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