package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

//func Cors(c *gin.Context) {
//	if c.Request.Method != "OPTIONS" {
//		c.Header("Access-Control-Allow-Origin", "*")
//		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
//		c.Header("Access-Control-Allow-Headers", "authorization, origin, content-type, accept")
//		c.Header("Allow", "HEAD,GET,POST,PUT,PATCH,DELETE,OPTIONS")
//		c.Header("Content-Type", "application/json")
//		c.Next()
//	} else {
//		c.Header("Access-Control-Allow-Origin", "*")
//		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
//		c.Header("Access-Control-Allow-Headers", "authorization, origin, content-type, accept")
//		c.Header("Allow", "HEAD,GET,POST,PUT,PATCH,DELETE,OPTIONS")
//		c.Header("Content-Type", "application/json")
//		c.AbortWithStatus(200)
//	}
//}

func Cors(c *gin.Context) {
	orgin := c.Request.Header.Get("Origin")
	if len(orgin) == 0 {
		return
	}

	host := c.Request.Host
	if orgin == "http://"+host || orgin == "https://"+host {
		return
	}

	if c.Request.Method == "OPTIONS" {
		generateHeaders(c)
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	generateHeaders(c)

}

func generateHeaders(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
	c.Header("Access-Control-Allow-Headers", "authorization, origin, content-type, accept, access-token")
	c.Header("Allow", "HEAD,GET,POST,PUT,PATCH,DELETE,OPTIONS")
	c.Header("Content-Type", "application/json")
}
