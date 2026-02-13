package middleware

import (
	"github.com/gin-gonic/gin"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.AbortWithStatusJSON(403, gin.H{
				"error": gin.H{"message": "관리자 권한이 필요합니다.", "code": "AUTH_003"},
			})
			return
		}
		c.Next()
	}
}
