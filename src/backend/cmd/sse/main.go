package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/zirachw/Tubes2_SeleniumSoup4/internal/dfs"
	"github.com/zirachw/Tubes2_SeleniumSoup4/internal/scraper"
)

func sseDFSHandler(recipeMap map[string]scraper.ElementData) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		ctx := r.Context()
		if ctx.Err() != nil {
			// client disconnected before we started
			log.Println("client disconnected before DFS started")
			return
		}
		fmt.Printf("client connected, starting DFS for %s\n", r.URL.Query().Get("element"))

		updates := make(chan dfs.Update)
		var tree dfs.Tree
		var counter uint64
		nextID := func() uint64 { return atomic.AddUint64(&counter, 1) }

		query := r.URL.Query()

		var count int = 0
		if query.Get("count") != "" {
			if query.Get("count") == "all" {
				// max int
				count = 2147483647
			} else {
				var err error
				count, err = strconv.Atoi(query.Get("count"))
				if err != nil {
					http.Error(w, "Invalid count parameter", http.StatusBadRequest)
					return
				}
			}
		}

		go func() {
			defer close(updates)
			dfs.DFSSearchInternal( // or DFSSearchWithUpdates
				recipeMap,
				query.Get("element"),
				count,
				&tree,
				updates,
				nextID,
				0,
			)
		}()

		// 5) Batch events
		const maxBatch = 20
		const maxDelay = 100 * time.Millisecond

		buffer := make([]dfs.Update, 0, maxBatch)
		timer := time.NewTimer(maxDelay)
		defer timer.Stop()

		sendBatch := func() error {
			if len(buffer) == 0 {
				return nil
			}
			// encode the whole slice as JSON
			b, err := json.Marshal(buffer)
			if err != nil {
				return err
			}
			// SSE frame: a single data: line with JSON payload
			fmt.Fprintf(w, "data: %s\n\n", b)
			flusher.Flush()
			buffer = buffer[:0] // reset
			timer.Reset(maxDelay)
			return nil
		}

		for {
			select {
			case <-ctx.Done():
				// client disconnected — stop processing
				log.Println("client went away, cancelling DFS")
				return

			case upd, ok := <-updates:
				if !ok {
					// DFS finished, send any remaining then close
					sendBatch()
					fmt.Printf("DFS finished, sending final update\n")
					return
				}
				buffer = append(buffer, upd)
				if len(buffer) >= maxBatch {
					if err := sendBatch(); err != nil {
						log.Println("sse write error:", err)
						return
					}
				}

			case <-timer.C:
				// send whatever we have every maxDelay
				if err := sendBatch(); err != nil {
					log.Println("sse write error:", err)
					return
				}
			}
		}
	})
}

func main() {
	// load your recipes.json
	dataBytes, err := os.ReadFile("data/recipes.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading recipes.json: %v\n", err)
		os.Exit(1)
	}
	var recipeMap map[string]scraper.ElementData
	if err := json.Unmarshal(dataBytes, &recipeMap); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing recipes.json: %v\n", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle("/stream", sseDFSHandler(recipeMap))
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
