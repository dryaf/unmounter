package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"github.com/google/safehtml/template"
)

func basicAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != username || pass != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		handler(w, r)
	}
}

func listMounts(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") == "application/json" {
		handleJSONRequest(w, r)
	} else {
		handleTemplateRequest(w, r)
	}
}

func handleJSONRequest(w http.ResponseWriter, r *http.Request) {
	var unmountErr error
	if r.Method == "POST" {
		var requestData struct {
			Device string `json:"device"`
		}
		if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
			http.Error(w, "JSON decode error", http.StatusBadRequest)
			return
		}
		unmountErr = unmountDevice(requestData.Device)
	}

	response := getSystemStatus()

	if unmountErr != nil {
		response.Error = unmountErr.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleTemplateRequest(w http.ResponseWriter, r *http.Request) {
	var unmountErr error
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "ParseForm() error", http.StatusInternalServerError)
			return
		}
		device := r.FormValue("device")
		unmountErr = unmountDevice(device)
		if unmountErr == nil {
			// Redirect after successful POST operation
			randomNumber := rand.Float64()
			http.Redirect(w, r, fmt.Sprint("/?r=", randomNumber), http.StatusSeeOther)
			return
		}
	}

	response := getSystemStatus()

	if unmountErr != nil {
		response.Error = unmountErr.Error()
	}

	t := template.Must(template.New("index").Parse(htmlTemplate))
	err := t.Execute(w, response)
	if err != nil {
		log.Println(err)
	}
}
