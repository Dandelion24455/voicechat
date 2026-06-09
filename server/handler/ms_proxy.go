package handler

import (
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var msUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (h *MediasoupHandler) ProxyMediasoupWS(c *gin.Context) {
	roomID := c.Param("id")
	token := c.Query("token")

	clientConn, err := msUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer clientConn.Close()

	targetURL := url.URL{
		Scheme: "ws",
		Host:   h.Cfg.MediasoupWSHost + ":" + h.Cfg.MediasoupWSPort,
		Path:   "/",
	}
	q := targetURL.Query()
	q.Set("token", token)
	q.Set("roomId", roomID)
	targetURL.RawQuery = q.Encode()

	msConn, _, err := websocket.DefaultDialer.Dial(targetURL.String(), nil)
	if err != nil {
		log.Printf("[ms-proxy] dial mediasoup: %v", err)
		return
	}
	defer msConn.Close()

	log.Printf("[ms-proxy] proxying room=%s", roomID)

	errCh := make(chan error, 2)
	go func() {
		for {
			msgType, msg, err := clientConn.ReadMessage()
			if err != nil {
				errCh <- err
				return
			}
			if err := msConn.WriteMessage(msgType, msg); err != nil {
				errCh <- err
				return
			}
		}
	}()
	go func() {
		for {
			msgType, msg, err := msConn.ReadMessage()
			if err != nil {
				errCh <- err
				return
			}
			if err := clientConn.WriteMessage(msgType, msg); err != nil {
				errCh <- err
				return
			}
		}
	}()

	<-errCh
}
