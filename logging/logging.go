package logging

import (
	"log"
	"os"
)

func Initialize_logging_to_file() {

	file, err := os.OpenFile("newsx.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(file)
	log.SetFlags(log.LstdFlags | log.Lshortfile) // Add this line to include date and time in the log

	log.Println("Application Starting")

}
