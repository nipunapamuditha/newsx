package main

import (
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nipunapamuditha/NEXO/db"
	"github.com/nipunapamuditha/NEXO/logging"
	"github.com/nipunapamuditha/NEXO/middleware"
	twitterhandler "github.com/nipunapamuditha/NEXO/twitterHandler"

	//	twitterhandler "github.com/nipunapamuditha/NEXO/twitterHandler"
	"github.com/nipunapamuditha/NEXO/users"
)

func main() {

	// initialize logging
	logging.Initialize_logging_to_file()

	// initialize database and test connection
	database := db.Initialize_database()

	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "https://demo.newsloop.xyz")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Handle preflight
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	v1 := router.Group("/v1")

	{

		// user signup and login handled b this endpoint

		v1.POST("/signuporlogin", func(c *gin.Context) {

			status, err, jwrtoken, new_status := users.UserSignUp(c, database)
			if err != nil {
				c.JSON(status, gin.H{"error": err.Error()})
				return
			}
			c.SetSameSite(http.SameSiteNoneMode)                                                   // Required for cross-origin cookies
			c.SetCookie("Authorization", jwrtoken, 3600, "/", "newsxapi.newsloop.xyz", true, true) // Secure=true for HTTPS
			c.SetSameSite(http.SameSiteNoneMode)                                                   // Required for cross-origin cookies
			c.SetCookie("Authorization", jwrtoken, 3600, "/", "newsxapi.newsloop.xyz", true, true) // Secure=true for HTTPS
			c.JSON(status, gin.H{
				"message":       "User signed up successfully",
				"existing_user": new_status,
			})
		})

		v1.POST("/getTwUsernames", func(c *gin.Context) {
			if err := middleware.RequireAuth(c, database); err != nil {
				log.Println(err)
				return

			}
			twitterhandler.GetTwitterUsernames(c, database)
		})

		v1.GET("/getuser", func(c *gin.Context) {
			if err := middleware.RequireAuth(c, database); err != nil {
				log.Println(err)
				return
			}
			users.Validate(c)
		})

		v1.POST("/update-preferences", func(c *gin.Context) {
			if err := middleware.RequireAuth(c, database); err != nil {
				log.Println(err)
				return
			}
			users.UpdateTwPref(c, database)
		})

		v1.POST("/publish-preferances", func(c *gin.Context) {
			if err := middleware.RequireAuth(c, database); err != nil {
				log.Println(err)
				return
			}
			users.PublishTwPref(c, database)
			// get user  preferances
		})

		v1.GET("/Get_Preferances", func(c *gin.Context) {
			if err := middleware.RequireAuth(c, database); err != nil {
				log.Println(err)
				return
			}
			users.GETTwPref(c, database)
		})

		v1.GET("/Generate_now", func(c *gin.Context) {
			if err := middleware.RequireAuth(c, database); err != nil {
				log.Println(err)
				return
			}
			c.Writer.Header().Set("Content-Type", "text/event-stream")
			c.Writer.Header().Set("Cache-Control", "no-cache")
			c.Writer.Header().Set("Connection", "keep-alive")
			c.Writer.Header().Set("Transfer-Encoding", "chunked")

			// Flush headers immediately
			c.Writer.Flush()

			// Create a channel for streaming updates
			updates := make(chan string)
			done := make(chan bool)

			// Start a goroutine to send updates to the channel
			go users.GenerateWithStream(c, database, updates, done)

			// Handle client disconnection
			c.Stream(func(w io.Writer) bool {
				select {
				case update := <-updates:
					// Send an SSE formatted message
					c.SSEvent("message", update)
					return true
				case <-done:
					// Generation complete
					return false
				case <-c.Request.Context().Done():
					// Client disconnected
					return false
				}
			})

			// create autio genertaion flow here
			// custom deepsek model for voiceover text
			// custom tts for voicever audio file generateion
		})

		v1.GET("/getaudio_files", func(c *gin.Context) {
			if err := middleware.RequireAuth(c, database); err != nil {
				log.Println(err)
				return
			}
			users.FetchAudioFiles(c, database)
			// show liks to audio files
		})

	}
	router.Run("0.0.0.0:9090")

}
