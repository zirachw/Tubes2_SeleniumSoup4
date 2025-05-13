package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/zirachw/Tubes2_SeleniumSoup4/internal/scraper"
	"github.com/zirachw/Tubes2_SeleniumSoup4/internal/search"
)

var (
	flagElement = flag.String("e", "Airplane",
		"element to search for")
	flagPaths = flag.Int("p", 1,
		"max unique paths (if 1 uses DFSSearch, >1 uses DFSSearchParallel)")
	flagOutput = flag.String("o", "",
		"optional output JSON file name (e.g. result.json)")
)

type ResultData struct {
	Element       string      `json:"element"`
	UniquePaths   int         `json:"uniquePaths"`
	TimeTaken     string      `json:"timeTaken"`
	NodesExplored uint64      `json:"nodesExplored"`
	NodesInTree   int         `json:"nodesInTree"`
	RecipeTree    interface{} `json:"recipeTree"`
}

func main() {
	flag.Parse()

	start := time.Now()

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

	var tree search.Tree

	var counter uint64
	nextID := func() uint64 { return atomic.AddUint64(&counter, 1) }

	updates := make(chan search.Update)

	var nodesExplored uint64

	go func() {
		nodesExplored = search.DFS(
			recipeMap,
			*flagElement,
			*flagPaths,
			&tree,
			updates,
			nextID,
			0,
		)
		close(updates) // Close the channel when done
	}()

// Now receive from the channel
for evt := range updates {
		fmt.Printf(
			"  → Stage=%-15s Elem=%-10s Tier=%2d Recipe#=%2d Info=%s\n",
			evt.Stage, evt.ElementName, evt.Tier, evt.RecipeIndex, evt.Info,
		)
	}

	elapsed := time.Since(start)

	if tree == nil {
		fmt.Fprintf(os.Stderr, "Element %q not found\n", *flagElement)
		os.Exit(1)
	}

	rootEl := &search.Element{
		Name:    tree.Name,
		Tier:    tree.Tier,
		Recipes: tree.Recipes,
		ID:      tree.ID,
	}
	
	fmt.Println("\n📖 Final DFS Recipe Tree:")
	search.PrintRecipeTree(rootEl, "")
	
	nodesInTree := search.CountTreeNodes(rootEl)
	
	fmt.Printf("\nNodes explored during DFS: %d\n", nodesExplored)
	fmt.Printf("Nodes in final tree: %d\n", nodesInTree)
	fmt.Printf("Unique paths: %d\n", tree.UniquePaths)
	fmt.Printf("Time taken: %v\n", elapsed)

	if *flagOutput != "" {
		out := ResultData{
			Element:       tree.Name,
			UniquePaths:   tree.UniquePaths,
			TimeTaken:     elapsed.String(),
			NodesExplored: nodesExplored,
			NodesInTree:   nodesInTree,
			RecipeTree:    rootEl,
		}
		j, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling output JSON: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(*flagOutput, j, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", *flagOutput, err)
			os.Exit(1)
		}
		fmt.Printf("\nResults written to %s\n", *flagOutput)
	}
}