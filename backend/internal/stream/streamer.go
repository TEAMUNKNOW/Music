package stream

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func InitRedis(addr string) {
	if opts, err := redis.ParseURL(addr); err == nil { rdb = redis.NewClient(opts) } else { rdb = redis.NewClient(&redis.Options{Addr:addr}) }
}

func ServeStream(c *fiber.Ctx, trackID, sourceURL string) error {
	if !ValidateSignedRequest(trackID, c.Query("exp"), c.Query("sig")) { return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error":"invalid or expired signature"}) }
	req, err := http.NewRequest(http.MethodGet, sourceURL, nil)
	if err != nil { return c.Status(500).SendString("failed to create request") }
	if rangeHeader := c.Get("Range"); rangeHeader != "" { req.Header.Set("Range", rangeHeader) }
	resp, err := (&http.Client{Timeout:30*time.Second}).Do(req)
	if err != nil { return c.Status(502).SendString("upstream error") }
	defer resp.Body.Close()
	if resp.StatusCode >= 400 { return c.Status(resp.StatusCode).SendString("upstream stream error") }
	contentType := resp.Header.Get("Content-Type"); if contentType=="" { contentType="audio/mpeg" }
	c.Set("Content-Type",contentType)
	for _, h := range []string{"Content-Length","Accept-Ranges","Content-Range","Cache-Control"} { if v:=resp.Header.Get(h); v!="" { c.Set(h,v) } }
	c.Status(resp.StatusCode)
	_, err = io.Copy(c.Response().BodyWriter(), resp.Body)
	return err
}

func serveCached(c *fiber.Ctx, data []byte) error {
	c.Set("Content-Type","audio/mpeg"); c.Set("Accept-Ranges","bytes")
	rangeHeader := c.Get("Range")
	if rangeHeader == "" { c.Set("Content-Length",strconv.Itoa(len(data))); return c.Send(data) }
	parts := strings.Split(strings.TrimPrefix(rangeHeader,"bytes="),"-")
	if len(parts)!=2 { return c.Status(416).SendString("invalid range") }
	start, err := strconv.Atoi(parts[0]); if err!=nil || start<0 { return c.Status(416).SendString("invalid range") }
	end := len(data)-1
	if parts[1]!="" { end,err=strconv.Atoi(parts[1]); if err!=nil { return c.Status(416).SendString("invalid range") } }
	if start>end || start>=len(data) || end>=len(data) { return c.Status(416).SendString("invalid range") }
	c.Status(206); c.Set("Content-Range",fmt.Sprintf("bytes %d-%d/%d",start,end,len(data))); c.Set("Content-Length",strconv.Itoa(end-start+1)); return c.Send(data[start:end+1])
}
