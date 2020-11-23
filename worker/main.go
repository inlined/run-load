package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Work struct {
	SleepMS int `json:"sleep_ms"`
}

func main() {
	log.Print("starting server...")
	http.HandleFunc("/", handler)

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	// Start HTTP server.
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	var work Work
	err := json.NewDecoder(r.Body).Decode(&work)
	if err != nil {
		log.Printf("Couldn't parse body. Using default")
	}
	sleepMS := work.SleepMS
	if sleepMS == 0 {
		sleepMS = 250
	}
	sleep := time.Duration(sleepMS) * time.Millisecond
	log.Println("Sleeping for", sleep)
	time.Sleep(sleep)
	w.Header().Add("Content-Type", "application/json")
	fmt.Fprint(w, "{}")
	r.Body.Close()
}
