package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Welcome(c *gin.Context) {
	userID, _ := c.Get("userID")
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Welcome, user %v!", userID)})
}