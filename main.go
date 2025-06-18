package main

import (
	"log"

	"github.com/ktappdev/ytaudio/config"
	"github.com/ktappdev/ytaudio/downloader"
	"github.com/ktappdev/ytaudio/playlist"
	"github.com/ktappdev/ytaudio/youtube"
)

func main() {
	// Set up logging to include timestamps
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	cfg := config.ParseFlags()

	if err := run(cfg); err != nil {
		log.Fatalf("Error: %v", err)
	}

	log.Println("Program completed successfully")
}

// run executes the main program logic based on the provided configuration
func run(cfg *config.Config) error {
	// Check if help flag is set or no command is provided
	if cfg.ShowHelp {
		config.ShowHelp()
		return nil
	}

	// Check if no command is provided
	if cfg.Query == "" && cfg.FilePath == "" && cfg.PlaylistID == "" && !cfg.SongListMode {
		config.ShowHelp()
		return nil
	}

	switch {
	case cfg.PlaylistID != "":
		log.Printf("Downloading playlist: %s", cfg.PlaylistID)
		return playlist.DownloadPlaylist(cfg)
	case cfg.SongListMode:
		if cfg.SongCSVFile != "" {
			log.Printf("Downloading songs from CSV file: %s", cfg.SongCSVFile)
		} else {
			log.Printf("Downloading song list: %s", cfg.SongList)
		}
		return downloader.DownloadSongList(cfg)
	case cfg.FilePath != "":
		log.Printf("Processing file: %s", cfg.FilePath)
		return downloader.ProcessFile(cfg)
	case cfg.Query == "":
		log.Println("No query provided")
		config.ShowHelp()
		return nil
	case cfg.ListMode:
		log.Printf("Listing videos for query: %s", cfg.Query)
		return youtube.ListVideos(cfg)
	case cfg.SongMode:
		log.Printf("Searching and downloading song: %s", cfg.Query)
		videoID, err := youtube.SearchAndDownloadSong(cfg)
		if err != nil {
			return err
		}
		return downloader.DownloadAudio(videoID)
	default:
		log.Printf("Downloading audio for query: %s", cfg.Query)
		return downloader.DownloadAudio(cfg.Query)
	}
}