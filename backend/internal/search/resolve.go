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

type StreamInfo struct {
	URL string `json:"url"`
	Quality string `json:"quality"`
	MimeType string `json:"mimeType"`
	Thumbnail string `json:"thumbnail,omitempty"`
	Title string `json:"title,omitempty"`
	Artist string `json:"artist,omitempty"`
	Duration int `json:"duration,omitempty"`
}

func ResolveStream(trackID, source string) (*StreamInfo, error) {
	source = strings.ToLower(strings.TrimSpace(source))
	switch source {
	case "jiosaavn":
		return resolveJioSaavn(trackID)
	case "youtube":
		return resolvePiped(trackID)
	default:
		if info, err := resolvePiped(trackID); err == nil { return info, nil }
		return resolveJioSaavn(trackID)
	}
}

func resolvePiped(videoID string) (*StreamInfo, error) {
	id := videoID
	if strings.Contains(id, "watch?v=") { id = strings.SplitN(id, "watch?v=", 2)[1] }
	id = strings.TrimPrefix(id, "/")
	if idx := strings.Index(id, "&"); idx >= 0 { id = id[:idx] }
	if idx := strings.Index(id, "?"); idx >= 0 { id = id[:idx] }
	if id == "" { return nil, fmt.Errorf("invalid youtube video id") }

	api := fmt.Sprintf("%s/streams/%s", pipedBase(), url.PathEscape(id))
	client := &http.Client{Timeout: 12 * time.Second}
	resp, err := client.Get(api)
	if err != nil { return nil, fmt.Errorf("piped request failed: %w", err) }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK { return nil, fmt.Errorf("piped status %d", resp.StatusCode) }
	body, err := io.ReadAll(resp.Body)
	if err != nil { return nil, err }

	var result struct {
		Title string `json:"title"`
		Uploader string `json:"uploader"`
		Duration int `json:"duration"`
		Thumbnail string `json:"thumbnailUrl"`
		AudioStreams []struct {
			URL string `json:"url"`
			Quality string `json:"quality"`
			MimeType string `json:"mimeType"`
			Bitrate int `json:"bitrate"`
			VideoOnly bool `json:"videoOnly"`
		} `json:"audioStreams"`
	}
	if err := json.Unmarshal(body, &result); err != nil { return nil, err }
	if len(result.AudioStreams) == 0 { return nil, fmt.Errorf("no audio streams found") }

	best := result.AudioStreams[0]
	for _, s := range result.AudioStreams[1:] {
		if !s.VideoOnly && s.Bitrate > best.Bitrate { best = s }
	}
	if best.URL == "" { return nil, fmt.Errorf("audio stream url missing") }
	return &StreamInfo{URL:best.URL, Quality:best.Quality, MimeType:best.MimeType, Thumbnail:result.Thumbnail, Title:result.Title, Artist:result.Uploader, Duration:result.Duration}, nil
}

func resolveJioSaavn(songID string) (*StreamInfo, error) {
	api := fmt.Sprintf("https://saavn.dev/api/songs/%s", url.PathEscape(songID))
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(api)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK { return nil, fmt.Errorf("jiosaavn status %d", resp.StatusCode) }
	body, err := io.ReadAll(resp.Body)
	if err != nil { return nil, err }

	var result struct { Data []struct {
		Name string `json:"name"`; Duration string `json:"duration"`
		Artists struct { Primary []struct { Name string `json:"name"` } `json:"primary"` } `json:"artists"`
		Image []struct { URL string `json:"url"` } `json:"image"`
		DownloadURL []struct { Quality string `json:"quality"`; URL string `json:"url"` } `json:"downloadUrl"`
	} `json:"data"` }
	if err := json.Unmarshal(body, &result); err != nil { return nil, err }
	if len(result.Data) == 0 { return nil, fmt.Errorf("song not found") }
	song := result.Data[0]
	if len(song.DownloadURL) == 0 { return nil, fmt.Errorf("no download url") }

	best := song.DownloadURL[len(song.DownloadURL)-1]
	for _, d := range song.DownloadURL { if d.Quality == "320kbps" { best = d; break } }
	artist := ""; if len(song.Artists.Primary) > 0 { artist = song.Artists.Primary[0].Name }
	thumb := ""; if len(song.Image) > 0 { thumb = song.Image[len(song.Image)-1].URL }
	dur := 0; _, _ = fmt.Sscanf(song.Duration, "%d", &dur)
	return &StreamInfo{URL:best.URL, Quality:best.Quality, MimeType:"audio/mpeg", Thumbnail:thumb, Title:song.Name, Artist:artist, Duration:dur}, nil
}
