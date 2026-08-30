package handlers

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/redis/go-redis/v9"

	"github.com/TEAMUNKNOW/Music/internal/search"
	"github.com/TEAMUNKNOW/Music/internal/stream"
	"github.com/TEAMUNKNOW/Music/internal/sync"
)

var redisClient *redis.Client

func InitRedis(addr string) {
	if opts, err := redis.ParseURL(addr); err == nil { redisClient=redis.NewClient(opts) } else { redisClient=redis.NewClient(&redis.Options{Addr:addr}) }
}

func Search(c *fiber.Ctx) error {
	q:=strings.TrimSpace(c.Query("q")); if q=="" { return c.Status(400).JSON(fiber.Map{"error":"query required"}) }
	tracks,err:=search.Search(q); if err!=nil { return c.Status(502).JSON(fiber.Map{"error":"search failed","details":err.Error()}) }
	return c.JSON(fiber.Map{"query":q,"results":tracks})
}

func publicBaseURL() string {
	if u:=strings.TrimRight(strings.TrimSpace(os.Getenv("BASE_URL")),"/"); u!="" && !strings.Contains(u,"localhost") { return u }
	if d:=strings.TrimSpace(os.Getenv("RAILWAY_PUBLIC_DOMAIN")); d!="" { return "https://"+strings.TrimRight(d,"/") }
	if u:=strings.TrimRight(strings.TrimSpace(os.Getenv("BASE_URL")),"/"); u!="" { return u }
	return "http://localhost:3000"
}

func GetTrack(c *fiber.Ctx) error {
	id:=c.Params("id"); if id=="" { return c.Status(400).JSON(fiber.Map{"error":"id required"}) }
	info,err:=search.ResolveStream(id,c.Query("source")); if err!=nil { return c.Status(502).JSON(fiber.Map{"error":"failed to resolve stream","details":err.Error()}) }
	if redisClient!=nil { _=redisClient.Set(context.Background(),"src:"+id,info.URL,5*time.Minute).Err() }
	return c.JSON(fiber.Map{"id":id,"title":info.Title,"artist":info.Artist,"duration":info.Duration,"thumbnail":info.Thumbnail,"quality":info.Quality,"streamUrl":stream.GenerateSignedURL(id,publicBaseURL()),"expiresIn":90})
}

func Stream(c *fiber.Ctx) error {
	id,exp,sig:=c.Params("id"),c.Query("exp"),c.Query("sig")
	if !stream.ValidateSignedRequest(id,exp,sig) { return c.Status(403).JSON(fiber.Map{"error":"invalid or expired signature"}) }
	var sourceURL string
	if redisClient!=nil { sourceURL,_=redisClient.Get(context.Background(),"src:"+id).Result() }
	if sourceURL=="" { info,err:=search.ResolveStream(id,""); if err!=nil { return c.Status(404).JSON(fiber.Map{"error":"stream source not found"}) }; sourceURL=info.URL }
	return stream.ServeStream(c,id,sourceURL)
}

func WebSocketUpgrade(c *fiber.Ctx) error { if websocket.IsWebSocketUpgrade(c) { return websocket.New(handleWS)(c) }; return fiber.ErrUpgradeRequired }
func handleWS(conn *websocket.Conn) {
	defer conn.Close(); roomID,userID:=conn.Query("room"),conn.Query("user"); if roomID=="" { roomID="default" }; if userID=="" { userID="anon" }
	room:=sync.GetOrCreateRoom(roomID); room.Join(conn,userID); defer room.Leave(conn)
	for { _,msg,err:=conn.ReadMessage(); if err!=nil { break }; room.HandleMessage(msg) }
}
