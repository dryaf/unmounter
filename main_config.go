package main

import (
	"crypto/rand"
	"log"
	"os"
)

// Configuration for the basic HTTP authentication
const EnvVarAuthUser = "AUTH_USER"
const EnvVarAuthPass = "AUTH_PASS"

var username = "admin"
var password = "sAcure1passw"

func init() {
	var ok bool

	envUsername, ok := os.LookupEnv(EnvVarAuthUser)
	if ok {
		username = envUsername
	}

	envPassword, ok := os.LookupEnv(EnvVarAuthPass)
	if ok {
		password = envPassword
	}
}

func generateRandomKey(length int) []byte {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		log.Fatal(err)
	}
	return key
}
