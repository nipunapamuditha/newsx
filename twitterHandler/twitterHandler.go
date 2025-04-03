package twitterhandler

import (
	"database/sql"

	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TwitterRequest struct {
	TwitterUsername string `json:"tw_user_name"`
}

type TwitterUserResponse struct {
	Data struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Username string `json:"username"`
	} `json:"data"`
}

func GetTwitterUsernames(context *gin.Context, db *sql.DB) {

	var request TwitterRequest

	if err := context.ShouldBindJSON(&request); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if request.TwitterUsername == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Twitter username is required"})
		return
	}

	tuname := request.TwitterUsername
	log.Println("Twitter username: ", tuname)

	context.JSON(http.StatusOK, gin.H{
		"message":  "User data fetched successfully",
		"name":     tuname,
		"username": tuname,
	})

	// store it in db or show no username found error to user

	//

	return

}
