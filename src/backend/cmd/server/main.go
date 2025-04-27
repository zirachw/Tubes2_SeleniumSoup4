package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	// "encoding/json"
	"github.com/zirachw/Tubes2_SeleniumSoup4/internal/scraper"
)

func makeRecipeHandler(data map[string]scraper.ElementData) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		elem := q.Get("element")
		algo := q.Get("algorithm")
		if algo == "" {
			algo = "bfs"
		}
		limStr := q.Get("limit")
		if limStr == "" {
			limStr = "10"
		}

		limit, err := strconv.Atoi(limStr)
		if err != nil || limit <= 0 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}

		if _, ok := data[elem]; !ok {
			http.Error(w, "unknown element", http.StatusBadRequest)
			return
		}

		// 4) run the right search
		//switch algo {
		//case "bfs":
		//    recipes = BFS(graph, elem, limit)
		//case "dfs":
		//    recipes = DFS(graph, elem, limit)
		//default:
		//    http.Error(w, "algorithm must be bfs or dfs", http.StatusBadRequest)
		//    return
		//}

		// 5) write JSON response

	}
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("../../data/recipes.json")
	if err != nil {
		http.Error(w, "File not found.", http.StatusNotFound)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/json")

	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Error while sending the file.", http.StatusInternalServerError)
	}
}

func main() {
	// Run the scraper and get the data map
	data := scraper.Run()

	// Print out the number of elements collected
	fmt.Printf("Successfully collected data for %d elements\n", len(data))

	// default mux is fine for one route:
	http.HandleFunc("/api/recipes", makeRecipeHandler(data))
	http.HandleFunc("/api/data", dataHandler)

	// start server
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

func recipeHandler(w http.ResponseWriter, r *http.Request) {
}
