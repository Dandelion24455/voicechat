package handler

import (
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	msProxyDialTimeout   = 5 * time.Second
	msProxyIdleTimeout   = 30 * time.Second
	msProxyMaxLifetime   = 15 * time.Minute
)

var msUpgrader = websocket.Upgrader{
	CheckOrigin:  func(r *http.Request) bool { return true },
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

var msDialer = websocket.Dialer{
	HandshakeTimeout: msProxyDialTimeout,
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

	msConn, _, err := msDialer.Dial(targetURL.String(), nil)
	if err != nil {
		log.Printf("[ms-proxy] dial mediasoup: %v", err)
		return
	}
	defer msConn.Close()

	log.Printf("[ms-proxy] proxying room=%s", roomID)

	deadline := time.Now().Add(msProxyMaxLifetime)
	clientConn.SetReadDeadline(deadline)
	msConn.SetReadDeadline(deadline)

	errCh := make(chan error, 2)

	go func() {
		for {
			if err := clientConn.SetReadDeadline(time.Now().Add(msProxyIdleTimeout)); err != nil {
				errCh <- err
				return
			}
			msgType, msg, err := clientConn.ReadMessage()
			if err != nil {
				errCh <- err
				return
			}
			if err := msConn.SetWriteDeadline(time.Now().Add(msProxyIdleTimeout)); err != nil {
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
			if err := msConn.SetReadDeadline(time.Now().Add(msProxyIdleTimeout)); err != nil {
				errCh <- err
				return
			}
			msgType, msg, err := msConn.ReadMessage()
			if err != nil {
				errCh <- err
				return
			}
			if err := clientConn.SetWriteDeadline(time.Now().Add(msProxyIdleTimeout)); err != nil {
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
