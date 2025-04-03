package db

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type Databse_config struct {
	PublicHost string
	Port       string
	DBUser     string
	DBPassword string
	DBName     string
}

func Initialize_database() *sql.DB {

	err := godotenv.Load("E:\\Personal Projects\\newsx_version_3\\.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	main_Databse_config := Databse_config{
		PublicHost: os.Getenv("DB_HOST"),
		Port:       os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
	}

	db, err := sql.Open("mysql", main_Databse_config.DBUser+":"+main_Databse_config.DBPassword+"@tcp("+main_Databse_config.PublicHost+":"+main_Databse_config.Port+")/"+main_Databse_config.DBName)

	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Database Drivier initialized successfully")
	}

	// Attempt to establish a connection to the database
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Database connection established successfully")
	}

	return db

	// initialize database and test connection
}
