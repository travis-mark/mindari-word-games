package main

import (
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

// Get channel ID from .env. Used while a single channel is used per bot.
//
// NOTE: Because this is called optimistically, it returns "" rather than an error if CHANNEL does not exist.
func getDefaultChannel() string {
	loadEnvironmentIfNeeded()
	return os.Getenv("CHANNEL")
}

// Get authorization token from .env. Used while a single token is used per bot.
//
// NOTE: Because this is called optimistically, it returns "" rather than an error if CHANNEL does not exist.
func getAuthorization() string {
	loadEnvironmentIfNeeded()
	return os.Getenv("AUTHORIZATION")
}
