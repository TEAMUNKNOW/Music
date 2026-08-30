package search

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// StreamInfo holds the playable audio URL + metadata
type StreamInfo struct {
	URL       string `json:"url"`
	Quality   string `json:"quality"`
	MimeType  string `json:"mimeType"`
	Thumbnail string `json:"thumbnail,omitempty"`
	Title     string `json:"title,omitempty"`
	Artist    string `json:"artist,omitempty"`
	Duration  int    `json:"duration,omitempty"`
}

// ResolveStream returns a direct playable audio URL for a track ID + source
func ResolveStream(trackID, source string) (*StreamInfo, error) {
	source = strings.ToLower(source)
	switch {
	case source == "youtube" || strings.HasPrefix(trackID, "watch") || strings.Contains(trackID, "youtube") || len(trackID) == 11:
		return resolvePiped(trackID)
	case source == "jiosaavn":
		return resolveJioSaavn(trackID)
	default:
		// Try Piped first (most common)
		if info, err := resolvePiped(trackID); err == nil {
			return info, nil
		}
		return resolveJioSaavn(trackID)
	}
}

func resolvePiped(videoID string) (*StreamInfo, error) {
	// Clean ID
	id := videoID
	if strings.Contains(id, "watch?v=") {
		parts := strings.Split(id, "watch?v=")
		id = parts[len(parts)-1]
	}
	id = strings.TrimPrefix(id, "/")
	if idx := strings.Index(id, "&"); idx != -1 {
		id = id[:idx]
	}

	// Public Piped instance — replace with self-hosted for production
	api := fmt.Sprintf("https://pipedapi.kavin.rocks/streams/%s", url.PathEscape(id))

	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Get(api)
	if err != nil {
		return nil, fmt.Errorf("piped request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("piped status %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Title     string `json:"title"`
		Uploader  string `json:"uploader"`
		Duration  int    `json:"duration"`
		Thumbnail string `json:"thumbnailUrl"`
		AudioStreams []struct {
			URL       string `json:"url"`
			Quality   string `json:"quality"`
			MimeType  string `json:"mimeType"`
			Bitrate   int    `json:"bitrate"`
		} `json:"audioStreams"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if len(result.AudioStreams) == 0 {
		return nil, fmt.Errorf("no audio streams found")
	}

	// Prefer highest bitrate audio/mp4 or audio/webm
	best := result.AudioStreams[0]
	for _, s := range result.AudioStreams {
		if s.Bitrate > best.Bitrate {
			best = s
		}
	}

	return &StreamInfo{
		URL:       best.URL,
		Quality:   best.Quality,
		MimeType:  best.MimeType,
		Thumbnail: result.Thumbnail,
		Title:     result.Title,
		Artist:    result.Uploader,
		Duration:  result.Duration,
	}, nil
}

func resolveJioSaavn(songID string) (*StreamInfo, error) {
	api := fmt.Sprintf("https://saavn.dev/api/songs/%s", url.PathEscape(songID))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(api)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Data []struct {
			Name     string `json:"name"`
			Duration string `json:"duration"`
			Artists  struct {
				Primary []struct {
					Name string `json:"name"`
				} `json:"primary"`
			} `json:"artists"`
			Image []struct {
				URL string `json:"url"`
			} `json:"image"`
			DownloadURL []struct {
				Quality string `json:"quality"`
				URL     string `json:"url"`
			} `json:"downloadUrl"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("song not found")
	}

	song := result.Data[0]
	if len(song.DownloadURL) == 0 {
		return nil, fmt.Errorf("no download url")
	}

	// Prefer 320kbps if available
	best := song.DownloadURL[len(song.DownloadURL)-1]
	for _, d := range song.DownloadURL {
		if d.Quality == "320kbps" {
			best = d
			break
		}
	}

	artist := ""
	if len(song.Artists.Primary) > 0 {
		artist = song.Artists.Primary[0].Name
	}
	thumb := ""
	if len(song.Image) > 0 {
		thumb = song.Image[len(song.Image)-1].URL
	}
	dur := 0
	fmt.Sscanf(song.Duration, "%d", &dur)

	return &StreamInfo{
		URL:       best.URL,
		Quality:   best.Quality,
		MimeType:  "audio/mpeg",
		Thumbnail: thumb,
		Title:     song.Name,
		Artist:    artist,
		Duration:  dur,
	}, nil
}
