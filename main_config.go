// ==== File: main_config.go ====
package main

import (
	"crypto/rand"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv" // 1. Import the godotenv library
)

// Configuration for the basic HTTP authentication
const EnvVarAuthUser = "AUTH_USER"
const EnvVarAuthPass = "AUTH_PASS"
const EnvVarDevMode = "DEV_MODE"

var username = "admin"
var password = "1b2a"
var devModeEnabled = false

func init() {
	err := godotenv.Load() // 2. Load .env file at the beginning of init()
	if err != nil {
		log.Println("Error loading .env file, using system environment variables (if set)")
	}

	var ok bool

	envUsername, ok := os.LookupEnv(EnvVarAuthUser)
	if ok {
		username = envUsername
	}

	envPassword, ok := os.LookupEnv(EnvVarAuthPass)
	if ok {
		password = envPassword
	}

	envDevMode, ok := os.LookupEnv(EnvVarDevMode)
	if ok {
		devModeEnabled, _ = strconv.ParseBool(envDevMode)
	}

	// Optional: Add logging to verify DEV_MODE
	log.Printf("DEV_MODE environment variable: %s, parsed devModeEnabled: %v", os.Getenv("DEV_MODE"), devModeEnabled)
}

func generateRandomKey(length int) []byte {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		log.Fatal(err)
	}
	return key
}
