package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/ktappdev/ytaudio/config"
)

const (
	youtubeAPIURL = "https://www.googleapis.com/youtube/v3/search"
)

// Video represents a YouTube video with its ID and Title
type Video struct {
	ID    string
	Title string
}

// ListVideos searches for videos and displays the results
func ListVideos(cfg *config.Config) error {
	log.Printf("Searching for videos with query: %s", cfg.Query)
	videos, err := SearchVideos(cfg.Query, cfg.APIKey)
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

// SearchAndDownloadSong searches for a song and returns the first result's video ID
func SearchAndDownloadSong(cfg *config.Config) (string, error) {
	log.Printf("Searching for song: %s", cfg.Query)
	videos, err := SearchVideos(cfg.Query+" audio", cfg.APIKey)
	if err != nil {
		return "", fmt.Errorf("error searching for song: %w", err)
	}

	if len(videos) == 0 {
		log.Println("No videos found for the song")
		return "", fmt.Errorf("no videos found for the song")
	}

	log.Printf("Found %d videos, returning the first result", len(videos))
	return videos[0].ID, nil
}

// SearchVideos performs a YouTube search using the YouTube Data API
func SearchVideos(query string, apiKey string) ([]Video, error) {
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