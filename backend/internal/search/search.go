package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Track struct {
	ID string `json:"id"`
	Title string `json:"title"`
	Artist string `json:"artist"`
	Duration int `json:"duration"`
	Thumbnail string `json:"thumbnail"`
	Source string `json:"source"`
	URL string `json:"url,omitempty"`
}

var rdb *redis.Client

func Init(redisAddr string) {
	if opts, err := redis.ParseURL(redisAddr); err == nil {
		rdb = redis.NewClient(opts)
	} else {
		rdb = redis.NewClient(&redis.Options{Addr: redisAddr})
	}
}

func Search(query string) ([]Track, error) {
	ctx := context.Background()
	cacheKey := "search:" + query
	if rdb != nil {
		if cached, err := rdb.Get(ctx, cacheKey).Result(); err == nil {
			var tracks []Track
			if json.Unmarshal([]byte(cached), &tracks) == nil { return tracks, nil }
		}
	}

	tracks, err := searchPiped(query)
	if err != nil || len(tracks) == 0 {
		tracks, err = searchJioSaavn(query)
	}
	if err == nil && len(tracks) > 0 && rdb != nil {
		if data, e := json.Marshal(tracks); e == nil { _ = rdb.Set(ctx, cacheKey, data, 10*time.Minute).Err() }
	}
	return tracks, err
}

func pipedBase() string {
	base := strings.TrimRight(os.Getenv("PIPED_API_URL"), "/")
	if base == "" { base = "https://pipedapi.kavin.rocks" }
	return base
}

func searchPiped(query string) ([]Track, error) {
	api := fmt.Sprintf("%s/search?q=%s&filter=music_songs", pipedBase(), url.QueryEscape(query))
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(api)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK { return nil, fmt.Errorf("piped search status %d", resp.StatusCode) }
	body, err := io.ReadAll(resp.Body)
	if err != nil { return nil, err }

	var result struct { Items []struct {
		URL string `json:"url"`; Title string `json:"title"`; UploaderName string `json:"uploaderName"`; Duration int `json:"duration"`; Thumbnail string `json:"thumbnail"`
	} `json:"items"` }
	if err := json.Unmarshal(body, &result); err != nil { return nil, err }

	tracks := make([]Track, 0, len(result.Items))
	for _, item := range result.Items {
		id := strings.TrimPrefix(item.URL, "/")
		tracks = append(tracks, Track{ID:id, Title:item.Title, Artist:item.UploaderName, Duration:item.Duration, Thumbnail:item.Thumbnail, Source:"youtube"})
	}
	return tracks, nil
}

func searchJioSaavn(query string) ([]Track, error) {
	api := fmt.Sprintf("https://saavn.dev/api/search/songs?query=%s", url.QueryEscape(query))
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(api)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK { return nil, fmt.Errorf("jiosaavn search status %d", resp.StatusCode) }
	body, err := io.ReadAll(resp.Body)
	if err != nil { return nil, err }

	var result struct { Data struct { Results []struct {
		ID string `json:"id"`; Name string `json:"name"`; PrimaryArtists string `json:"primaryArtists"`; Duration string `json:"duration"`; Image []struct { URL string `json:"url"` } `json:"image"`
	} `json:"results"` } `json:"data"` }
	if err := json.Unmarshal(body, &result); err != nil { return nil, err }

	tracks := make([]Track, 0, len(result.Data.Results))
	for _, item := range result.Data.Results {
		thumb := ""; if len(item.Image) > 0 { thumb = item.Image[len(item.Image)-1].URL }
		dur := 0; _, _ = fmt.Sscanf(item.Duration, "%d", &dur)
		tracks = append(tracks, Track{ID:item.ID, Title:item.Name, Artist:item.PrimaryArtists, Duration:dur, Thumbnail:thumb, Source:"jiosaavn"})
	}
	return tracks, nil
}
