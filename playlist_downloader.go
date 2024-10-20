package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type PlaylistDownloader struct {
	APIKey           string
	ConcurrentLimit  int
	DownloadFunction func(string) error
}

func NewPlaylistDownloader(apiKey string, concurrentLimit int, downloadFunc func(string) error) *PlaylistDownloader {
	return &PlaylistDownloader{
		APIKey:           apiKey,
		ConcurrentLimit:  concurrentLimit,
		DownloadFunction: downloadFunc,
	}
}

func (pd *PlaylistDownloader) DownloadPlaylist(playlistID string) error {
	ctx := context.Background()
	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(pd.APIKey))
	if err != nil {
		return fmt.Errorf("error creating YouTube client: %w", err)
	}

	videos, err := pd.getPlaylistVideos(youtubeService, playlistID)
	if err != nil {
		return err
	}

	log.Printf("Found %d videos in playlist", len(videos))

	jobs := make(chan string, len(videos))
	results := make(chan error, len(videos))

	var wg sync.WaitGroup
	for w := 1; w <= pd.ConcurrentLimit; w++ {
		wg.Add(1)
		go pd.worker(jobs, results, &wg)
	}

	for _, video := range videos {
		jobs <- video
	}
	close(jobs)

	wg.Wait()
	close(results)

	for err := range results {
		if err != nil {
			log.Printf("Error downloading video: %v", err)
		}
	}

	return nil
}

func (pd *PlaylistDownloader) getPlaylistVideos(service *youtube.Service, playlistID string) ([]string, error) {
	var videos []string
	nextPageToken := ""

	for {
		call := service.PlaylistItems.List([]string{"snippet"}).
			PlaylistId(playlistID).
			MaxResults(50).
			PageToken(nextPageToken)

		response, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("error fetching playlist items: %w", err)
		}

		for _, item := range response.Items {
			videos = append(videos, item.Snippet.ResourceId.VideoId)
		}

		nextPageToken = response.NextPageToken
		if nextPageToken == "" {
			break
		}
	}

	return videos, nil
}

func (pd *PlaylistDownloader) worker(jobs <-chan string, results chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	for videoID := range jobs {
		log.Printf("Downloading video: %s", videoID)
		err := pd.DownloadFunction(videoID)
		results <- err
	}
}
