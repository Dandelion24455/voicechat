package handler

import (
	"net/http"
	"voicechat-server/config"

	"github.com/gin-gonic/gin"
	"github.com/livekit/protocol/auth"
)

type LiveKitHandler struct {
	Cfg *config.Config
}

func (h *LiveKitHandler) GetToken(c *gin.Context) {
	roomID := c.Param("id")
	userID := c.GetString("user_id")
	username := c.GetString("username")

	at := auth.NewAccessToken(h.Cfg.LiveKitKey, h.Cfg.LiveKitSecret)
	at.SetName(username)
	at.SetIdentity(userID)
	at.AddGrant(&auth.VideoGrant{
		RoomJoin: true,
		Room:     roomID,
		CanPublish:     &[]bool{true}[0],
		CanSubscribe:   &[]bool{true}[0],
	})

	token, err := at.ToJWT()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"livekit_token": token,
		"livekit_url":   h.Cfg.LiveKitURL,
		"room":          roomID,
	})
}
