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
	flagUpdates = flag.Bool("u", false,
		"stream live updates (true) or run silently (false)")
	flagOutput = flag.String("o", "",
		"optional output JSON file name (e.g. result.json)")
)

var nodeCount int

func printRecipeTree(el *search.Element, indent string) int {
    if el == nil {
        return 0
    }
    fmt.Printf("%s%s (tier=%d, id=%d)\n", indent, el.Name, el.Tier, el.ID)
    nodeCount++

    if len(el.Recipes) == 0 {
        return 1
    }
    total := 0
    for i, r := range el.Recipes {
        fmt.Printf("%s  Recipe %d:\n", indent, i+1)
        fmt.Printf("%s    Left ingredient:\n", indent)
        lp := printRecipeTree(r.Left, indent+"      ")
        fmt.Printf("%s    Right ingredient:\n", indent)
        rp := printRecipeTree(r.Right, indent+"      ")
        contrib := lp * rp
        total += contrib
        fmt.Printf("%s  Recipe %d contributes %d path(s)\n", indent, i+1, contrib)
    }
    return total
}

type ResultData struct {
	Element       string      `json:"element"`
	UniquePaths   int         `json:"uniquePaths"`
	TimeTaken     string      `json:"timeTaken"`
	NodesExplored int         `json:"nodesExplored"`
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

	if *flagUpdates {
		updates := search.DFSWithUpdates(
			recipeMap,
			*flagElement,
			*flagPaths,
			&tree,
			nextID,
		)
		fmt.Println("⏳ Streaming DFS events:")
		for evt := range updates {
			fmt.Printf(
				"  → Stage=%-15s Elem=%-10s Tier=%2d Recipe#=%2d Info=%s\n",
				evt.Stage, evt.ElementName, evt.Tier, evt.RecipeIndex, evt.Info,
			)
		}

	} else {
		if *flagPaths <= 1 {
			search.DFS(recipeMap, *flagElement, *flagPaths, &tree, nextID)
		} else {
			search.DFSParallel(recipeMap, *flagElement, *flagPaths, &tree, nextID)
		}
	}

	elapsed := time.Since(start)

	if tree == nil {
		fmt.Fprintf(os.Stderr, "Element %q not found\n", *flagElement)
		os.Exit(1)
	}

	nodeCount = 0
	fmt.Println("\n📖 Final DFS Recipe Tree:")
	printRecipeTree(&search.Element{
		Name:    tree.Name,
		Tier:    tree.Tier,
		Recipes: tree.Recipes,
	}, "")
	fmt.Printf("\nTotal nodes explored: %d\n", counter)
	fmt.Printf("Unique paths found: %d\n", tree.UniquePaths)
	fmt.Printf("Time taken: %v\n", elapsed)

	if *flagOutput != "" {
		out := ResultData{
			Element:       tree.Name,
			UniquePaths:   tree.UniquePaths,
			TimeTaken:     elapsed.String(),
			NodesExplored: nodeCount,
			RecipeTree: &search.Element{
				Name:    tree.Name,
				Tier:    tree.Tier,
				Recipes: tree.Recipes,
				ID:      tree.ID,
			},
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
