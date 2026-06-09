package middleware

import "github.com/gin-gonic/gin"

func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowMap := make(map[string]bool)
	for _, o := range allowedOrigins {
		allowMap[o] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		// Empty allowMap = allow all (dev mode fallback)
		if len(allowMap) == 0 || allowMap[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
