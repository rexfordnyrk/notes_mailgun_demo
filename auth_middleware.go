package main

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

func authMiddleware(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("userID")
	if userID == nil {
		c.Redirect(http.StatusSeeOther, "/login")
		c.Abort()
		return
	}
	c.Set("userID", userID)
	c.Next()
}

func setSession(c *gin.Context, userID uint) {
	session := sessions.Default(c)
	session.Set("userID", userID)
	session.Save()
}

func getSessionUserID(c *gin.Context) (uint, bool) {
	session := sessions.Default(c)
	userID := session.Get("userID")
	if userID == nil {
		return 0, false
	}
	uid, ok := userID.(uint)
	if !ok {
		// If stored as int, try conversion
		if i, ok := userID.(int); ok {
			return uint(i), true
		}
		return 0, false
	}
	return uid, true
}

func clearSession(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("userID")
	session.Save()
}
