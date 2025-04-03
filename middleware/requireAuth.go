package middleware

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/nipunapamuditha/NEXO/users" // Ensure this path is correct and the User type is defined in this package
)

func RequireAuth(c *gin.Context, db *sql.DB) error {

	log.Println(c)

	errr := godotenv.Load("E:\\Personal Projects\\newsx_version_3\\.env")
	if errr != nil {
		log.Printf("Error loading .env file: %v", errr)
		log.Println("FAILED TO FETCH ENVIRONMENT VARIABLES")
		c.AbortWithStatus(http.StatusUnauthorized)
		return errr
	}

	// get ookier
	tokenstring, err := c.Cookie("Authorization")

	if err != nil {
		log.Printf("Error in getting cookie %v", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return err
	}

	tokem, err := jwt.Parse(tokenstring, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("Unexpected signing method: %v", token.Header["alg"])
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("JWT_SECRET")), nil

	})

	if claims, ok := tokem.Claims.(jwt.MapClaims); ok && tokem.Valid {

		if time.Now().Unix() > int64(claims["exp"].(float64)) {
			log.Println("Token expired")
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		// validate the  get db and set values in context

		fmt.Println(claims["sub"])

		rows, err := db.Query("SELECT * FROM users WHERE email = ?", claims["sub"])
		if err != nil {

			log.Printf("DB Query error to validate user %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)

			return err
		}

		defer rows.Close()

		// Prepare to scan each row

		var userss []users.User

		for rows.Next() {
			var user users.User
			err := rows.Scan(&user.FirstName, &user.LastName, &user.Email, &user.UniqueID)
			if err != nil {
				log.Printf("Error scanning row: %v", err)

				c.AbortWithStatus(http.StatusInternalServerError)
				return err
			}
			userss = append(userss, user) // Add user to the slice
		}

		if err = rows.Err(); err != nil {
			log.Printf("Row iteration error: %v", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return err

		}

		// Log or process the result
		if len(userss) > 0 {
			user := userss[0]
			log.Printf("User: %+v\n", user)
			c.Set("user", user)
		}

	} else {
		log.Println(err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return err
	}

	// validate

	// if valid, continue
	return nil
}
