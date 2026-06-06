package handler

import (
	"crypto/rand"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"
	"voicechat-server/config"
	"voicechat-server/model"
	"voicechat-server/store"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	DB  *store.DB
	Cfg *config.Config
}

const playerIDChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generatePlayerID() string {
	for {
		b := make([]byte, 4)
		for i := range b {
			n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(playerIDChars))))
			b[i] = playerIDChars[n.Int64()]
		}
		id := string(b)
		// Validate: no ambiguous chars (O/0, I/1), not all same
		if strings.ContainsAny(id, "OI") {
			continue
		}
		return id
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=2,max=32"`
		Password string `json:"password" binding:"required,min=6,max=128"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username 2-32 chars. Password 6-128 chars."})
		return
	}

	existing, _ := h.DB.GetUserByUsername(c.Request.Context(), req.Username)
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username taken"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hash failed"})
		return
	}

	playerID := generatePlayerID()
	for {
		taken, err := h.DB.IsPlayerIDTaken(c.Request.Context(), playerID)
		if err != nil {
			log.Printf("player_id check error: %v", err)
		}
		if !taken {
			break
		}
		playerID = generatePlayerID()
	}

	user := &model.User{
		ID:           uuid.NewString(),
		PlayerID:     playerID,
		Username:     req.Username,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}
	if err := h.DB.CreateUser(c.Request.Context(), user); err != nil {
		log.Printf("create user error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create user failed"})
		return
	}

	token, _ := h.generateToken(user)
	c.JSON(http.StatusCreated, gin.H{"token": token, "player_id": playerID, "user": user})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.DB.GetUserByUsername(c.Request.Context(), req.Username)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, _ := h.generateToken(user)
	c.JSON(http.StatusOK, gin.H{"token": token, "user": user})
}

func (h *AuthHandler) generateToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":   user.ID,
		"player_id": user.PlayerID,
		"username":  user.Username,
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.Cfg.JWTSecret))
}
