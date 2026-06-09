package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

func Auth(secret string, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if tokenStr == "" {
			tokenStr = c.Query("token")
		}
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		if rdb != nil {
			n, _ := rdb.Exists(c.Request.Context(), "blacklist:"+tokenStr).Result()
			if n > 0 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token revoked"})
				return
			}
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		userID, _ := claims["user_id"].(string)
		playerID, _ := claims["player_id"].(string)
		username, _ := claims["username"].(string)
		c.Set("user_id", userID)
		c.Set("player_id", playerID)
		c.Set("username", username)
		c.Set("token_str", tokenStr)
		c.Next()
	}
}

// BlacklistToken adds a token to Redis blacklist, expiring when the JWT would expire.
func BlacklistToken(ctx context.Context, rdb *redis.Client, tokenStr string) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return
	}
	exp, _ := claims["exp"].(float64)
	if exp == 0 {
		return
	}
	rdb.Do(ctx, "SET", "blacklist:"+tokenStr, "1", "EXAT", int64(exp))
}
