package users

import (
	"context"
	"database/sql"
	"strings"

	"fmt"

	"time"

	"log"
	"net/http"
	"net/url"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"

	"github.com/nipunapamuditha/NEXO/audio_generation_azure"

	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nipunapamuditha/NEXO/utils"
)

type User struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	UniqueID  string `json:"unique_id"`
}

var requestBody struct {
	Token string `json:"token"`
}

type TwitterPublish struct {
	Usernames []string `json:"usernames"`
}

func UserSignUp(context *gin.Context, db *sql.DB) (int, error, string, int) {

	var signup_user User

	if err := context.BindJSON(&requestBody); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return http.StatusBadRequest, err, "", 0
	}

	token := requestBody.Token // google auth token

	G_token, err := utils.GetEnvVariable("GOOGLE_CLIENT_ID")
	if err != nil {
		log.Printf("Error in getting google auth token %v", err)
		return http.StatusInternalServerError, err, "", 0
	}

	// Validate the token
	payload, err := idtoken.Validate(context, token, G_token)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return http.StatusUnauthorized, err, "", 0
	}

	// Assign the values to user object
	signup_user.Email = payload.Claims["email"].(string)
	signup_user.FirstName = payload.Claims["given_name"].(string)
	signup_user.LastName = payload.Claims["family_name"].(string)
	signup_user.UniqueID = payload.Claims["jti"].(string)

	log.Printf("User: %+v", signup_user)

	// get ggole auth token and validate the user

	// asign the values to user object

	rows, err := db.Query("SELECT * FROM users WHERE email = ?", signup_user.Email)
	if err != nil {

		log.Printf("DB Query error %v", err)
		return http.StatusInternalServerError, err, "", 0

	}

	defer rows.Close()

	if rows.Next() {

		log.Printf("User with email %s already exists", signup_user.Email)

		jwttoken, err := GenerateJWTToken(signup_user.Email)
		if err != nil {
			log.Printf("Error in generating JWT token %v", err)
			return http.StatusInternalServerError, err, "", 1
		}

		return http.StatusOK, nil, jwttoken, 1

	} else {

		log.Printf("inserting user into the database")

		// Insert the user into the database
		_, err := db.Exec("INSERT INTO users (first_name, last_name, email, unique_id) VALUES (?, ?, ?, ?)", signup_user.FirstName, signup_user.LastName, signup_user.Email, signup_user.UniqueID)
		if err != nil {
			log.Printf("DB Insert error %v", err)
			return http.StatusInternalServerError, err, "", 0
		}

		// Generate a JWT token for the user

		tokenString, err := GenerateJWTToken(signup_user.Email)
		if err != nil {
			log.Printf("Error in generating JWT token %v", err)
			return http.StatusInternalServerError, err, "", 0
		}

		// checkeing weather the user is new or not

		return http.StatusOK, nil, tokenString, 0
	}

}

func GetPreferances(context *gin.Context, db *sql.DB) (int, error) {

	// Get the user from the context
	dbuser, exists := context.Get("user")
	if !exists {
		log.Printf("User not found in context")
		return http.StatusUnauthorized, fmt.Errorf("User not found in context")
	}

	newuser, ok := dbuser.(User)
	if !ok {
		log.Printf("Error asserting user type")
		return http.StatusInternalServerError, fmt.Errorf("error asserting user type")
	}

	// Query the database for the user's preferances
	rows, err := db.Query("SELECT * FROM user_pref WHERE user_email = ?", newuser.Email)
	if err != nil {
		log.Printf("DB Query error %v", err)
		return http.StatusInternalServerError, err
	}

	defer rows.Close()

	// get users preferances in to an arry  of usernames

	return 0, nil
}

func GenerateJWTToken(useremail string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": useremail,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})

	jwtSecret, err := utils.GetEnvVariable("JWT_SECRET")
	if err != nil {
		log.Printf("Error in getting JWT secret: %v", err)

		return "", err
	}

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Printf("Error in generating JWT token: %v", err)
		return "", err
	}

	log.Printf("Token generated successfully: %v", tokenString)
	return tokenString, nil
}

// experimental test cookie function

func Validate(c *gin.Context) {

	user, exists := c.Get("user")

	if !exists {
		log.Printf("User not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	log.Printf("Type of user: %T", user)
	log.Printf("Value of user: %+v", user)

	// Type assert the user to your User struct
	userStruct, ok := user.(User)
	if !ok {
		log.Printf("User type assertion failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User type assertion failed"})
		return
	}

	// Now you can access the fields of userStruct
	fmt.Printf("FirstName: %s, LastName: %s, Email: %s, UniqueID: %s\n", userStruct.FirstName, userStruct.LastName, userStruct.Email, userStruct.UniqueID)
	c.JSON(http.StatusOK, gin.H{
		"user": userStruct,
	})
}

func UpdateTwPref(c *gin.Context, db *sql.DB) {

	user, exists := c.Get("user")

	if !exists {
		log.Printf("User not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	log.Printf("Type of user: %T", user)
	log.Printf("Value of user: %+v", user)

	// Type assert the user to your User struct
	userStruct, ok := user.(User)
	if !ok {
		log.Printf("User type assertion failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User type assertion failed"})
		return
	}

	type userDataRequest struct {
		TwitterUsernames []string `json:"tw_user_name"`
	}

	var request userDataRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if len(request.TwitterUsernames) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Twitter username is required"})
		return
	}

	log.Printf("Twitter usernames: %v", request.TwitterUsernames)

	// Now you can access the fields of userStruct
	fmt.Printf("FirstName: %s, LastName: %s, Email: %s, UniqueID: %s\n", userStruct.FirstName, userStruct.LastName, userStruct.Email, userStruct.UniqueID)
	c.JSON(http.StatusOK, gin.H{
		"user": userStruct,
	})

}

func PublishTwPref(c *gin.Context, db *sql.DB) {

	// get username list slice from the context

	var TwitterPublish TwitterPublish

	if error := c.BindJSON(&TwitterPublish); error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if len(TwitterPublish.Usernames) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Twitter username is required"})
		return
	}

	// get user from the context

	user, exists := c.Get("user")

	if !exists {
		log.Printf("User not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	// get email from the user

	userStruct, ok := user.(User)
	if !ok {
		log.Printf("User type assertion failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User type assertion failed"})
		return
	}

	// get the user email from the user struct

	user_email := userStruct.Email

	// update the user preferances in the database user_pref table

	// Delete existing preferences for this user
	_, err := db.Exec("DELETE FROM user_pref WHERE user_email = ?", user_email)
	if err != nil {
		log.Printf("Error deleting existing preferences: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB Delete error"})
		return
	}

	for _, username := range TwitterPublish.Usernames {
		_, err := db.Exec("INSERT INTO user_pref (user_email, profile_username) VALUES (?, ?)", user_email, username)
		if err != nil {
			log.Printf("Error inserting username %s: %v", username, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB Insert error"})
			return
		}

	}

	c.JSON(http.StatusOK, gin.H{"message": "Twitter usernames inserted successfully"})

	// update databse relaetd to user

	// update it on mysql db

}

func GETTwPref(c *gin.Context, db *sql.DB) {

	user, exists := c.Get("user")

	if !exists {
		log.Printf("User not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userStruct, ok := user.(User)
	if !ok {
		log.Printf("User type assertion failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User type assertion failed"})
		return
	}

	// check db and get user rpeferances fro email

	// get the user email from the user struct

	user_email := userStruct.Email

	// Query the database for the user's preferances
	rows, err := db.Query("SELECT * FROM user_pref WHERE user_email = ?", user_email)
	if err != nil {
		log.Printf("DB Query error %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB Query error"})
		return
	}

	// get users preferances in to an arry  of usernames

	var usernames []string

	for rows.Next() {
		var profile_username string
		var user_email string
		err := rows.Scan(&user_email, &profile_username)
		if err != nil {
			log.Printf("Error scanning username: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning username"})
			return
		}

		usernames = append(usernames, profile_username)
	}

	defer rows.Close()

	c.JSON(http.StatusOK, gin.H{"usernames": usernames})

}

func GenerateWithStream(c *gin.Context, db *sql.DB, updates chan<- string, done chan<- bool) {

	defer close(done)

	user, exists := c.Get("user")

	if !exists {
		log.Printf("User not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userStruct, ok := user.(User)
	if !ok {
		log.Printf("User type assertion failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User type assertion failed"})
		return
	}

	// check db and get user rpeferances fro email

	// get the user email from the user struct

	user_email := userStruct.Email

	// Query the database for the user's preferances

	updates <- "Featching user preferances"

	rows, err := db.Query("SELECT * FROM user_pref WHERE user_email = ?", user_email)
	if err != nil {
		log.Printf("DB Query error %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB Query error"})
		return
	}

	// get users preferances in to an arry  of usernames

	var usernames []string

	for rows.Next() {
		var profile_username string
		var user_email string
		err := rows.Scan(&user_email, &profile_username)
		if err != nil {
			log.Printf("Error scanning username: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning username"})
			return
		}
		// list of usernames are below
		usernames = append(usernames, profile_username)
	}

	defer rows.Close()

	updates <- "Featching news articles"

	articles, err := audio_generation_azure.Fetach_substack_rss(usernames)
	if err != nil {
		log.Printf("Error fetching articles: %v", err)
		updates <- "ERROR: " + err.Error()
		return
	}

	updates <- "Articals featched successfully"

	time.Sleep(2 * time.Second)

	updates <- "Generating script"

	news_script, err := audio_generation_azure.Generate_script_azure(articles)
	if err != nil {
		log.Printf("Error generating script: %v", err)
		updates <- "ERROR: " + err.Error()
		return
	}
	// test
	time.Sleep(2 * time.Second)

	updates <- "Generating Voiceover"

	status, err := audio_generation_azure.Generate_audio_file_azure(news_script, user_email)
	log.Printf("Audio generation status: %v, error: %v", status, err)
	if err != nil {
		log.Printf("Error generating audio: %v", err)
		updates <- "ERROR: " + err.Error()
		return
	}

	// Add explicit success logging
	log.Printf("Audio generation completed successfully")
	updates <- "Audio file generated successfully"
	updates <- "SUCCESS: true"

}

// fetch audio files

func FetchAudioFiles(c *gin.Context, db *sql.DB) {
	// Get user from the context
	user, exists := c.Get("user")
	if !exists {
		log.Printf("User not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userStruct, ok := user.(User)
	if !ok {
		log.Printf("User type assertion failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User type assertion failed"})
		return
	}

	// Get user email
	userEmail := userStruct.Email

	// Configure MinIO client
	minioEndpoint := "minioapi.newsloop.xyz"
	bucketName := "newsx"
	userPrefix := userEmail + "/"

	// Get MinIO credentials from environment variables
	accessKeyID, err := utils.GetEnvVariable("MINIO_ACCESS_KEY")
	if err != nil {
		log.Printf("Error getting MinIO access key: %v", err)
		return
	}

	secretAccessKey, err := utils.GetEnvVariable("MINIO_SECRET_KEY")
	if err != nil {
		log.Printf("Error getting MinIO secret key: %v", err)
		return
	}

	// Initialize MinIO client
	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		log.Printf("Error creating MinIO client: %v", err)
		return
	}

	// List all objects in the user's folder
	var audioFiles []string
	ctx := context.Background()

	objectCh := minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    userPrefix,
		Recursive: true,
	})

	// Generate presigned URLs for each object (valid for 24 hours)
	for object := range objectCh {
		if object.Err != nil {
			log.Printf("Error listing objects: %v", object.Err)
			continue
		}

		// Generate a presigned URL
		presignedURL, err := minioClient.PresignedGetObject(ctx, bucketName, object.Key, time.Hour*24, nil)
		if err != nil {
			log.Printf("Error generating URL for %s: %v", object.Key, err)
			continue
		}

		// Add to our list of files
		audioFiles = append(audioFiles, presignedURL.String())
	}

	if len(audioFiles) == 0 {
		log.Printf("No audio files found for user %s", userEmail)
		c.JSON(http.StatusOK, gin.H{"audio_files": []string{}}) // Return empty slice, not an error
		return                                                  // Return empty slice, not an error
	}

	c.JSON(http.StatusOK, gin.H{
		"audio_files": audioFiles})
}

func DeleteAudioFile(c *gin.Context, db *sql.DB) {
	// Get user from the context
	user, exists := c.Get("user")
	if !exists {
		log.Printf("User not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userStruct, ok := user.(User)
	if !ok {
		log.Printf("User type assertion failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User type assertion failed"})
		return
	}

	// Get user email
	userEmail := userStruct.Email

	// Parse request body
	var requestBody struct {
		ObjectName string `json:"object_name"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	log.Printf("Request body: %+v", requestBody)
	log.Printf("User email: %s", requestBody.ObjectName)

	log.Printf("Request body: %+v", requestBody)

	// URL decode the object name
	decodedObjectName, err := url.QueryUnescape(requestBody.ObjectName)
	if err != nil {
		log.Printf("Error decoding object name: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid object name format"})
		return
	}

	log.Printf("Decoded object name: %s", decodedObjectName)

	if decodedObjectName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Object name is required"})
		return
	}

	// Configure MinIO client
	minioEndpoint := "minioapi.newsloop.xyz"
	bucketName := "newsx"

	// Get MinIO credentials from environment variables
	accessKeyID, err := utils.GetEnvVariable("MINIO_ACCESS_KEY")
	if err != nil {
		log.Printf("Error getting MinIO access key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get MinIO credentials"})
		return
	}

	secretAccessKey, err := utils.GetEnvVariable("MINIO_SECRET_KEY")
	if err != nil {
		log.Printf("Error getting MinIO secret key: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get MinIO credentials"})
		return
	}

	// Initialize MinIO client
	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		log.Printf("Error creating MinIO client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to storage"})
		return
	}

	// The object name already includes the userPrefix in your JSON request
	objectKey := decodedObjectName

	// Optional: Validate that the object belongs to this user
	if !strings.HasPrefix(objectKey, userEmail+"/") {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this file"})
		return
	}

	err = minioClient.RemoveObject(context.Background(), bucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		log.Printf("Error deleting object %s: %v", objectKey, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete audio file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Audio file deleted successfully"})
}

// locally avalible functions for generatenow #########################
