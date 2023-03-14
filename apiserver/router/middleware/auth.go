package middleware

import (
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/xnile/muxwaf/pkg/token"
)

// ParseToken 处理Token
func ParseToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetBool("bypass") {
			return
		}

		var bearer = c.GetHeader("Authorization")
		var tokenStr string
		fmt.Sscanf(bearer, "Bearer %s", &tokenStr)
		claims, err := token.Decode(tokenStr)
		if err == nil {
			if uid, ok := claims["uid"].(float64); ok {
				c.Set("uid", int64(uid))
			} else {
				// TODO 日志
				c.Set("uid", 0)
			}
		} else {
			if ve, ok := err.(*jwt.ValidationError); ok {
				if ve.Errors&jwt.ValidationErrorExpired != 0 {
					c.JSON(402, gin.H{"success": 402, "message": "登陆超时，请重新登陆", "data": nil})
					c.Abort()
					return
				} else {
					// TODO
					c.Set("uid", 0)
				}
			} else {
				//TODO
				c.Set("uid", 0)
			}
		}
		c.Next()
	}
}

// ParseToken 处理Token
//func ParseToken() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		var bearer = c.GetHeader("Authorization")
//		var jwt string
//		fmt.Sscanf(bearer, "Bearer %s", &jwt)
//		if id, ok := token.Decode(jwt); ok {
//			c.Set("uid", id)
//		} else {
//			c.Set("uid", "")
//		}
//		c.Next()
//	}
//}

// AuthRequired 需要登陆
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetBool("bypass") {
			return
		}
		if id := c.GetInt64("uid"); id == 0 {
			c.JSON(403, gin.H{"success": 403, "message": "请先登录", "data": nil})
			c.Abort()
			return
		}
		c.Next()
	}
}
