// 实现 JWT 认证中间件，保护需要认证的路由

package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		if strings.EqualFold(c.GetHeader("Upgrade"), "websocket") {
			tokenString = c.Query("token")
			log.Printf("WebSocket request, token from query: %s", tokenString)
			if tokenString == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token in query"})
				c.Abort()
				return
			}
		} else {
			tokenString = c.GetHeader("Authorization")
			log.Printf("HTTP request, Authorization header: %s", tokenString)
			if !strings.HasPrefix(tokenString, "Bearer ") {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization header"})
				c.Abort()
				return
			}
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte("secret_key"), nil
		})
		if err != nil || !token.Valid {
			log.Printf("Invalid token: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			userID := claims["user_id"]
			log.Printf("Authenticated user_id: %v", userID)
			c.Set("user_id", userID)
		}

		c.Next()
	}
}
