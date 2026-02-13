package middleware

import (
	"github.com/coby/colight/apps/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AuthMiddleware(tokenService *service.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractBearerToken(c.GetHeader("Authorization"))
		if tokenString == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"error": gin.H{"message": "인증이 필요합니다.", "code": "AUTH_001"},
			})
			return
		}

		claims, err := tokenService.ValidateAccessToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{
				"error": gin.H{"message": "유효하지 않은 토큰입니다.", "code": "AUTH_002"},
			})
			return
		}

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{
				"error": gin.H{"message": "유효하지 않은 토큰입니다.", "code": "AUTH_002"},
			})
			return
		}

		c.Set("user_id", userID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func extractBearerToken(header string) string {
	if len(header) > 7 && header[:7] == "Bearer " {
		return header[7:]
	}
	return ""
}
