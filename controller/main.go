package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Work struct {
	SleepMS int `json:"sleep_ms"`
}

var (
	workerSleep   = flag.Duration("worker_sleep", 250*time.Millisecond, "How long workers should sleep in each request")
	workerAddress = flag.String("worker_address", "https://worker-2u6il5ns4a-uc.a.run.app", "Address to send traffic") //"https://worker-lkta64fkwa-uc.a.run.app", "Address to send traffic")
	burstCount    = flag.Int("burst_count", 20, "number of concurrent requests to send to workers")
	burstDuration = flag.Duration("burst_duration", time.Minute, "Duration of traffic burst to send to worker")
)

func main() {
	log.Print("starting server...")
	flag.Parse()
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
	log.Printf("Creating a burst of %d requests for %s", *burstCount, *burstDuration)
	work := Work{
		SleepMS: int(*workerSleep / time.Millisecond),
	}

	ctx, _ := context.WithTimeout(r.Context(), *burstDuration)
	done := make(chan struct{})
	for i := 0; i < *burstCount; i++ {
		log.Printf("Starting worker %d", i)
		go func(i int) {
			sendWorkRepeatedly(ctx, http.DefaultClient, work, i)
			done <- struct{}{}
		}(i)
	}

	for i := 0; i < *burstCount; i++ {
		<-done
	}
	log.Printf("Done bursting traffic to workers")
	fmt.Fprint(w, "{}")
}

func sendWorkRepeatedly(ctx context.Context, client *http.Client, work Work, workerNo int) {
	for {
		// Quit at the end of the timer
		select {
		case <-ctx.Done():
			log.Printf("Quitting worker %d", workerNo)
			return
		default:
		}
		log.Printf("Sending a new request on goroutine %d", workerNo)

		b, err := json.Marshal(work)
		if err != nil {
			log.Printf("Failed to JSON encode work: %s", err)
			return
		}
		resp, err := client.Post(*workerAddress, "application/json", bytes.NewReader(b))
		if err != nil {
			log.Printf("Request on goroutine %d failed with err %s", workerNo, err)
		} else {
			log.Printf("Successfully sent request on goroutine %d to worker", workerNo)
			resp.Body.Close()
		}
	}
}
