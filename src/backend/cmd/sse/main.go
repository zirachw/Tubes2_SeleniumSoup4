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

	"github.com/zirachw/Tubes2_SeleniumSoup4/internal/scraper"
	"github.com/zirachw/Tubes2_SeleniumSoup4/internal/search"
)

type ResultData struct {
	Element       string      `json:"element"`
	UniquePaths   int         `json:"uniquePaths"`
	TimeTaken     string      `json:"timeTaken"`
	NodesExplored int         `json:"nodesExplored"`
	NodesInTree   int         `json:"nodesInTree"`
	RecipeTree    interface{} `json:"recipeTree"`
}

func sseHandler(recipeMap map[string]scraper.ElementData) http.Handler {
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
			log.Println("client disconnected before search started")
			return
		}
		fmt.Printf("client connected, starting search for %s\n", r.URL.Query().Get("element"))

		updates := make(chan search.Update)
		var tree search.Tree

		start := time.Now()
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

		var nodesExplored uint64 = 0;
		if query.Get("algorithm") == "DFS" {

			go func() {
				defer close(updates)
				nodesExplored = search.DFS(
					recipeMap,
					query.Get("element"),
					count,
					&tree,
					updates,
					nextID,
					0,
					&nodesExplored,
				)
			}()
		} else if query.Get("algorithm") == "BFS" {

			var err error
			var paths []*search.Element

			paths, nodesExplored, err = search.BFS(recipeMap, query.Get("element"), count)
			if err != nil {
				log.Fatalf("BFS search error: %v", err)
				return
			}

			go func() {
				tree = search.CreateFullTree(paths, updates, query.Get("element"), nextID)
				close(updates)
			}()
		}

		// 5) Batch events
		const maxBatch = 20
		const maxDelay = 100 * time.Millisecond

		buffer := make([]search.Update, 0, maxBatch)
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
				log.Println("client went away, cancelling search")
				return

			case upd, ok := <-updates:
				if !ok {
					buffer = append(buffer, search.Update{
						Stage:       "doneRecipe",
						ElementName: query.Get("element"),
					})
					sendBatch()
					fmt.Printf("search finished, sending final update\n")

					elapsed := time.Since(start)

					// 3) If we still didn't find anything, warn
					if tree == nil || tree.UniquePaths == 0 {
						fmt.Fprintf(os.Stderr, "Element %q not found or no paths\n", query.Get("element"))
					}

					// Create recipe tree for counting nodes (if needed)
					var nodesInTree int = 0
					var treeToSend interface{} = nil
					
					if tree != nil {
						// Convert the Tree to Element for counting
						rootEl := &search.Element{
							Name:    tree.Name,
							Tier:    tree.Tier,
							Recipes: tree.Recipes,
							ID:      tree.ID,
						}
						
						// Count nodes in the tree
						nodesInTree = search.CountTreeNodes(rootEl)
						
						// Decide what to send as recipe tree
						if query.Get("liveUpdate") != "true" {
							treeToSend = tree
						}
					}

					fmt.Printf("\nTotal nodes explored: %d\n", nodesExplored)
					fmt.Printf("Nodes in final tree: %d\n", nodesInTree)
					fmt.Printf("Unique paths found: %d\n", tree.UniquePaths)
					fmt.Printf("Time taken: %v\n", elapsed)

					// 5) Emit JSON
					out := ResultData{
						Element:       query.Get("element"),
						UniquePaths:   tree.UniquePaths,
						TimeTaken:     elapsed.String(),
						NodesExplored: int(nodesExplored),
						NodesInTree:   nodesInTree,
						RecipeTree:    treeToSend,
					}
					buf, err := json.Marshal(out)
					if err != nil {
						log.Fatalf("Error marshaling JSON: %v", err)
					}
					// write to user client
					fmt.Fprintf(w, "data: %s\n\n", buf)
					flusher.Flush()
					return
				}
				buffer = append(buffer, upd)
				// print update to stdout
				/*
					fmt.Printf(
						"  → Stage=%-15s Elem=%-10s Tier=%2d Recipe#=%2d Info=%s\n parentID=%d\n, leftID=%d\n, rightID=%d\n",
						upd.Stage, upd.ElementName, upd.Tier, upd.RecipeIndex, upd.Info, upd.ParentID, upd.LeftID, upd.RightID,
					)
				*/
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
	mux.Handle("/stream", sseHandler(recipeMap))
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}