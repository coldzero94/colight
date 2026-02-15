package middleware

import (
	"github.com/gin-gonic/gin"
)

// RoleLevel returns the numeric level for a role string.
// Higher level = more privilege.
func RoleLevel(role string) int {
	switch role {
	case "user":
		return 1
	case "manager":
		return 2
	case "admin":
		return 3
	case "super_admin":
		return 4
	default:
		return 0
	}
}

// RequireRole returns middleware that enforces a minimum role level.
func RequireRole(minRole string) gin.HandlerFunc {
	minLevel := RoleLevel(minRole)

	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(403, gin.H{
				"error": gin.H{"message": "관리자 권한이 필요합니다.", "code": "AUTH_003"},
			})
			return
		}

		roleStr, ok := role.(string)
		if !ok || RoleLevel(roleStr) < minLevel {
			c.AbortWithStatusJSON(403, gin.H{
				"error": gin.H{"message": "관리자 권한이 필요합니다.", "code": "AUTH_003"},
			})
			return
		}

		c.Next()
	}
}
