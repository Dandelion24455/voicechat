package handler

import (
	"fmt"
	"net/http"
	"voicechat-server/config"

	"github.com/gin-gonic/gin"
)

type MediasoupHandler struct {
	Cfg *config.Config
}

func (h *MediasoupHandler) GetConnectionInfo(c *gin.Context) {
	roomID := c.Param("id")

	// Extract the raw JWT token from Authorization header for the client to use
	// when connecting to the mediasoup WebSocket server
	token := c.Query("token")
	if token == "" {
		token = c.GetHeader("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
	}

	wsHost := h.Cfg.MediasoupWSHost
	wsPort := h.Cfg.MediasoupWSPort
	wsURL := fmt.Sprintf("ws://%s:%s", wsHost, wsPort)

	c.JSON(http.StatusOK, gin.H{
		"room":             roomID,
		"mediasoup_ws_url": wsURL,
		"token":            token,
	})
}
