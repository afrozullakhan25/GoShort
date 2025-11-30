package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		hostname, _ := os.Hostname()
		c.JSON(http.StatusOK, gin.H{
			"app_name": "GoShort",
			"version":  "v1.0.0",
			"server":   hostname,
			"status":   "healthy",
			"message":  "Running on Docker & Gin",
		})
	})

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.POST("/shorten", func(c *gin.Context) {
		type RequestBody struct {
			URL string `json:"url"`
		}
		var req RequestBody
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"original":  req.URL,
			"shortened": "http://goshort.ly/xyz123", 
			"source":    "Mock DB",
		})
	})

	r.Run(":8080")
}
