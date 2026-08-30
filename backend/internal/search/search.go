package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/redis/go-redis/v9"
)

type Track struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Artist    string `json:"artist"`
	Duration  int    `json:"duration"`
	Thumbnail string `json:"thumbnail"`
	Source    string `json:"source"` // youtube | jiosaavn
	URL       string `json:"url,omitempty"`
}

var rdb *redis.Client

func Init(redisAddr string) {
	rdb = redis.NewClient(&redis.Options{Addr: redisAddr})
}

// Search performs multi-source search with Redis cache
func Search(query string) ([]Track, error) {
	ctx := context.Background()
	cacheKey := "search:" + query

	if rdb != nil {
		if cached, err := rdb.Get(ctx, cacheKey).Result(); err == nil {
			var tracks []Track
			if json.Unmarshal([]byte(cached), &tracks) == nil {
				return tracks, nil
			}
		}
	}

	// Primary: Piped (YouTube privacy frontend)
	tracks, err := searchPiped(query)
	if err != nil || len(tracks) == 0 {
		// Fallback could be Invidious or JioSaavn
		tracks, err = searchJioSaavn(query)
	}

	if err == nil && len(tracks) > 0 && rdb != nil {
		data, _ := json.Marshal(tracks)
		rdb.Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return tracks, err
}

func searchPiped(query string) ([]Track, error) {
	// Public Piped instance (replace with your own for production)
	api := fmt.Sprintf("https://pipedapi.kavin.rocks/search?q=%s&filter=music_songs", url.QueryEscape(query))

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(api)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Items []struct {
			URL         string `json:"url"`
			Title       string `json:"title"`
			UploaderName string `json:"uploaderName"`
			Duration    int    `json:"duration"`
			Thumbnail   string `json:"thumbnail"`
		} `json:"items"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var tracks []Track
	for _, item := range result.Items {
		id := item.URL
		if len(id) > 0 && id[0] == '/' {
			id = id[1:]
		}
		tracks = append(tracks, Track{
			ID:        id,
			Title:     item.Title,
			Artist:    item.UploaderName,
			Duration:  item.Duration,
			Thumbnail: item.Thumbnail,
			Source:    "youtube",
		})
	}
	return tracks, nil
}

func searchJioSaavn(query string) ([]Track, error) {
	// Public unofficial endpoint example (replace with stable source)
	api := fmt.Sprintf("https://saavn.dev/api/search/songs?query=%s", url.QueryEscape(query))

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(api)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Data struct {
			Results []struct {
				ID     string `json:"id"`
				Name   string `json:"name"`
				PrimaryArtists string `json:"primaryArtists"`
				Duration string `json:"duration"`
				Image  []struct {
					URL string `json:"url"`
				} `json:"image"`
			} `json:"results"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var tracks []Track
	for _, item := range result.Data.Results {
		thumb := ""
		if len(item.Image) > 0 {
			thumb = item.Image[len(item.Image)-1].URL
		}
		dur := 0
		fmt.Sscanf(item.Duration, "%d", &dur)

		tracks = append(tracks, Track{
			ID:        item.ID,
			Title:     item.Name,
			Artist:    item.PrimaryArtists,
			Duration:  dur,
			Thumbnail: thumb,
			Source:    "jiosaavn",
		})
	}
	return tracks, nil
}
