package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	sampleKey := os.Getenv("SAMPLE_KEY")
	unExistingKey := os.Getenv("UN_EXISTING_KEY")

	log.Println(sampleKey, unExistingKey)

	url := os.Getenv("PROTOTYPING_TARGET_URL")
	script := os.Getenv("PROTOTYPING_JQ_SCRIPT")

	log.Println(url, script)
}
