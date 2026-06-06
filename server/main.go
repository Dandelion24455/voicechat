package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"voicechat-server/config"
	"voicechat-server/handler"
	"voicechat-server/middleware"
	"voicechat-server/store"
	"voicechat-server/ws"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	cfg := config.Load()
	gin.SetMode(gin.ReleaseMode)

	db, err := store.NewDB(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	if err := db.Migrate(ctx()); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})

	authH := &handler.AuthHandler{DB: db, Cfg: cfg}
	roomH := &handler.RoomHandler{DB: db, Cfg: cfg}
	lkH := &handler.LiveKitHandler{Cfg: cfg}

	hub := ws.NewHub(rdb)

	r := gin.Default()
	r.Use(middleware.CORS())
	r.StaticFile("/", "/client/index.html")

	api := r.Group("/api")
	{
		api.POST("/register", authH.Register)
		api.POST("/login", authH.Login)

		auth := api.Group("", middleware.Auth(cfg.JWTSecret))
		{
			auth.POST("/rooms", roomH.Create)
			auth.GET("/rooms", roomH.List)
			auth.DELETE("/rooms/:id", roomH.Delete)
			auth.POST("/rooms/:id/join", roomH.Join)
			auth.POST("/rooms/join-by-code", roomH.JoinByCode)
			auth.GET("/rooms/:id/token", lkH.GetToken)
			auth.GET("/ws/room/:id", func(c *gin.Context) {
				hub.Handle(c)
			})
		}
	}

	if cfg.Domain != "" {
		// Also serve HTTP on 8080 as fallback for users without HTTPS access
		go func() {
			log.Printf("HTTP fallback on :%s", cfg.Port)
			if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
				log.Fatalf("http fallback: %v", err)
			}
		}()

		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(cfg.Domain),
			Cache:      autocert.DirCache("/data/certs"),
		}

		tlsConfig := &tls.Config{
			GetCertificate: certManager.GetCertificate,
			MinVersion:     tls.VersionTLS12,
		}

		httpsSrv := &http.Server{
			Addr:      ":443",
			Handler:   r,
			TLSConfig: tlsConfig,
		}

		go func() {
			log.Printf("HTTP→HTTPS redirect on :80")
			err := http.ListenAndServe(":80", certManager.HTTPHandler(nil))
			if err != nil {
				log.Fatalf("http redirect: %v", err)
			}
		}()

		log.Printf("server starting on :443 (TLS for %s)", cfg.Domain)
		if err := httpsSrv.ListenAndServeTLS("", ""); err != nil {
			log.Fatalf("https: %v", err)
		}
	} else {
		log.Printf("server starting on :%s", cfg.Port)
		r.Run(":" + cfg.Port)
	}
}

func ctx() context.Context {
	return context.Background()
}
