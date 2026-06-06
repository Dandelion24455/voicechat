package handler

import (
	"crypto/rand"
	"log"
	"math/big"
	"net/http"
	"strings"
	"voicechat-server/config"
	"voicechat-server/model"
	"voicechat-server/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RoomHandler struct {
	DB  *store.DB
	Cfg *config.Config
}

func generateInviteCode() string {
	for {
		b := make([]byte, 4)
		for i := range b {
			n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(playerIDChars))))
			b[i] = playerIDChars[n.Int64()]
		}
		code := string(b)
		if !strings.ContainsAny(code, "OI01") {
			return code
		}
	}
}

func (h *RoomHandler) Create(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required,min=1,max=64"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	room := &model.Room{
		ID:         uuid.NewString(),
		Name:       req.Name,
		CreatorID:  c.GetString("user_id"),
		InviteCode: generateInviteCode(),
	}
	if err := h.DB.CreateRoom(c.Request.Context(), room); err != nil {
		log.Printf("create room error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create room failed"})
		return
	}
	c.JSON(http.StatusCreated, room)
}

func (h *RoomHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	rooms, err := h.DB.ListRoomsByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list rooms failed"})
		return
	}
	if rooms == nil {
		rooms = []model.Room{}
	}
	c.JSON(http.StatusOK, rooms)
}

func (h *RoomHandler) Delete(c *gin.Context) {
	roomID := c.Param("id")
	userID := c.GetString("user_id")
	if err := h.DB.DeleteRoom(c.Request.Context(), roomID, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found or not yours"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (h *RoomHandler) Join(c *gin.Context) {
	roomID := c.Param("id")
	room, err := h.DB.GetRoomByID(c.Request.Context(), roomID)
	if err != nil || room == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"room": room})
}

func (h *RoomHandler) JoinByCode(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required,len=4"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invite code must be 4 characters"})
		return
	}

	code := strings.ToUpper(req.Code)
	room, err := h.DB.GetRoomByInviteCode(c.Request.Context(), code)
	if err != nil || room == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"room": room})
}
