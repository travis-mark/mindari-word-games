package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var envLoaded bool

// Load .env once to get secrets.
//
// Guarded by a simple lock - not threadsafe.
func loadEnvironmentIfNeeded() {
	if envLoaded {
		return
	}
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	envLoaded = true
}

// Get authorization token (BEARER or BOT) from .env.
func getAuthorization() (string, error) {
	loadEnvironmentIfNeeded()
	bearer := os.Getenv("BEARER")
	if bearer != "" {
		return "Bearer " + bearer, nil
	}
	bot := os.Getenv("BOT")
	if bot != "" {
		return "Bot " + bot, nil
	}
	return "", fmt.Errorf("No authorization found in BEARER or BOT. Check your environment or .env file.")
}
