package main

import (
	"github.com/JimmyZhangJW/videoPlatformBackendDemo/controllers"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// ignore CORS rules for local testing
func SetCORSHeaderMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
		} else {
			c.Next()
		}
	}
}

func main() {
	r := gin.Default()
	r.Use(SetCORSHeaderMiddleware())

	// Get all public video's meta
	r.GET("/videoMetas", controllers.VideoC.GetPublicVideoMeta)

	// Upload video's metadata
	r.POST("/videoMetas", controllers.VideoC.PostVideoMetaData)

	// Upload video chunk content
	r.POST("/videoChunks", controllers.VideoC.PostVideoChunk)

	// Merge video chunks if all chunks are received, otherwise return the hashes of missing contents
	r.POST("/merge", controllers.VideoC.Merge)

	// Serving static storage file
	r.Static("/storage", "storage")

	log.Fatalln(r.Run()) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
