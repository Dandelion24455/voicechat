package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

const memberTTL = 3 * time.Minute
const refreshInterval = 1 * time.Minute

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	Conn     *websocket.Conn
	UserID   string
	PlayerID string
	Username string
	RoomID   string
	mu       sync.Mutex
}

type Message struct {
	Type     string   `json:"type"`
	UserID   string   `json:"user_id,omitempty"`
	PlayerID string   `json:"player_id,omitempty"`
	Username string   `json:"username,omitempty"`
	RoomID   string   `json:"room_id,omitempty"`
	Speaking bool     `json:"speaking,omitempty"`
	Members  []Member `json:"members,omitempty"`
}

type Member struct {
	UserID   string `json:"user_id"`
	PlayerID string `json:"player_id"`
	Username string `json:"username"`
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]bool
	rdb     *redis.Client
}

func NewHub(rdb *redis.Client) *Hub {
	return &Hub{
		clients: make(map[*Client]bool),
		rdb:     rdb,
	}
}

func (hub *Hub) Handle(c *gin.Context) {
	roomID := c.Param("id")
	userID := c.GetString("user_id")
	playerID := c.GetString("player_id")
	username := c.GetString("username")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
		Conn:     conn,
		UserID:   userID,
		PlayerID: playerID,
		Username: username,
		RoomID:   roomID,
	}

	hub.mu.Lock()
	hub.clients[client] = true
	hub.mu.Unlock()

	hub.addMember(roomID, userID, playerID, username)
	hub.broadcastMembers(roomID)

	// Periodic TTL refresh — keeps Redis keys alive while connected
	done := make(chan struct{})
	defer close(done)
	go func() {
		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				hub.refreshMember(roomID, userID)
			case <-done:
				return
			}
		}
	}()

	defer func() {
		hub.removeMember(roomID, userID)
		hub.broadcastMembers(roomID)
		hub.mu.Lock()
		delete(hub.clients, client)
		hub.mu.Unlock()
		conn.Close()
	}()

	// Read deadline ensures ReadJSON won't block forever on a dead connection
	conn.SetReadDeadline(time.Now().Add(2 * time.Minute))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(2 * time.Minute))
		return nil
	})

	// Ping every 30s to keep the connection alive and detect dead peers
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				client.mu.Lock()
				conn.WriteMessage(websocket.PingMessage, nil)
				client.mu.Unlock()
			case <-done:
				return
			}
		}
	}()

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}
		msg.UserID = userID
		msg.Username = username
		msg.RoomID = roomID
		hub.broadcast(roomID, msg)
	}
}

func (hub *Hub) addMember(roomID, userID, playerID, username string) {
	ctx := context.Background()
	membersKey := "room:" + roomID + ":members"
	userKey := "user:" + userID
	hub.rdb.SAdd(ctx, membersKey, userID)
	hub.rdb.Expire(ctx, membersKey, memberTTL)
	hub.rdb.HSet(ctx, userKey, map[string]interface{}{
		"player_id": playerID,
		"username":  username,
		"room_id":   roomID,
	})
	hub.rdb.Expire(ctx, userKey, memberTTL)
}

func (hub *Hub) refreshMember(roomID, userID string) {
	ctx := context.Background()
	hub.rdb.Expire(ctx, "room:"+roomID+":members", memberTTL)
	hub.rdb.Expire(ctx, "user:"+userID, memberTTL)
}

func (hub *Hub) removeMember(roomID, userID string) {
	ctx := context.Background()
	hub.rdb.SRem(ctx, "room:"+roomID+":members", userID)
	hub.rdb.Del(ctx, "user:"+userID)
}

func (hub *Hub) broadcastMembers(roomID string) {
	members := hub.getMembers(roomID)
	msg := Message{Type: "members", Members: members}
	hub.broadcastRaw(roomID, msg)
}

func (hub *Hub) getMembers(roomID string) []Member {
	ctx := context.Background()
	userIDs, _ := hub.rdb.SMembers(ctx, "room:"+roomID+":members").Result()
	var members []Member
	for _, uid := range userIDs {
		username, _ := hub.rdb.HGet(ctx, "user:"+uid, "username").Result()
		playerID, _ := hub.rdb.HGet(ctx, "user:"+uid, "player_id").Result()
		members = append(members, Member{UserID: uid, PlayerID: playerID, Username: username})
	}
	return members
}

func (hub *Hub) broadcast(roomID string, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("ws marshal: %v", err)
		return
	}
	hub.mu.RLock()
	defer hub.mu.RUnlock()
	for c := range hub.clients {
		if c.RoomID == roomID {
			c.mu.Lock()
			c.Conn.WriteMessage(websocket.TextMessage, data)
			c.mu.Unlock()
		}
	}
}

func (hub *Hub) broadcastRaw(roomID string, msg Message) {
	data, _ := json.Marshal(msg)
	hub.mu.RLock()
	defer hub.mu.RUnlock()
	for c := range hub.clients {
		if c.RoomID == roomID {
			c.mu.Lock()
			c.Conn.WriteMessage(websocket.TextMessage, data)
			c.mu.Unlock()
		}
	}
}
