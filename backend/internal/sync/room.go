package sync

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
)

type RoomState struct {
	TrackID   string  `json:"trackId"`
	Title     string  `json:"title"`
	Artist    string  `json:"artist"`
	Position  float64 `json:"position"` // seconds
	IsPlaying bool    `json:"isPlaying"`
	UpdatedAt int64   `json:"updatedAt"`
}

type Client struct {
	Conn   *websocket.Conn
	UserID string
}

type Room struct {
	ID      string
	State   RoomState
	Clients map[*websocket.Conn]*Client
	mu      sync.RWMutex
}

var (
	rooms   = make(map[string]*Room)
	roomsMu sync.RWMutex
)

func GetOrCreateRoom(id string) *Room {
	roomsMu.Lock()
	defer roomsMu.Unlock()
	if r, ok := rooms[id]; ok {
		return r
	}
	r := &Room{
		ID:      id,
		Clients: make(map[*websocket.Conn]*Client),
		State: RoomState{
			UpdatedAt: time.Now().UnixMilli(),
		},
	}
	rooms[id] = r
	return r
}

func (r *Room) Join(conn *websocket.Conn, userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Clients[conn] = &Client{Conn: conn, UserID: userID}

	// Send current state to new client
	data, _ := json.Marshal(map[string]interface{}{
		"type":  "state",
		"state": r.State,
	})
	_ = conn.WriteMessage(websocket.TextMessage, data)
}

func (r *Room) Leave(conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Clients, conn)
}

func (r *Room) Broadcast(msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	for conn := range r.Clients {
		_ = conn.WriteMessage(websocket.TextMessage, data)
	}
}

func (r *Room) UpdateState(newState RoomState) {
	r.mu.Lock()
	r.State = newState
	r.State.UpdatedAt = time.Now().UnixMilli()
	r.mu.Unlock()

	r.Broadcast(map[string]interface{}{
		"type":  "state",
		"state": r.State,
	})
}

// HandleMessage processes play / pause / seek / track change
func (r *Room) HandleMessage(raw []byte) {
	var msg struct {
		Type     string  `json:"type"`
		TrackID  string  `json:"trackId"`
		Title    string  `json:"title"`
		Artist   string  `json:"artist"`
		Position float64 `json:"position"`
	}
	if err := json.Unmarshal(raw, &msg); err != nil {
		return
	}

	r.mu.Lock()
	switch msg.Type {
	case "play":
		r.State.IsPlaying = true
		r.State.Position = msg.Position
	case "pause":
		r.State.IsPlaying = false
		r.State.Position = msg.Position
	case "seek":
		r.State.Position = msg.Position
	case "track":
		r.State.TrackID = msg.TrackID
		r.State.Title = msg.Title
		r.State.Artist = msg.Artist
		r.State.Position = 0
		r.State.IsPlaying = true
	}
	r.State.UpdatedAt = time.Now().UnixMilli()
	state := r.State
	r.mu.Unlock()

	r.Broadcast(map[string]interface{}{
		"type":  "state",
		"state": state,
	})
}
