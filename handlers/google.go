package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GoogleLogin(c *gin.Context) {
	// Frontend handles OAuth popup
	// Backend only verifies token / creates user (later)
	// TODO: Verify Google ID token

	c.JSON(http.StatusOK, gin.H{
		"message": "Google OAuth success (stub)",
	})
}
