package main

import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "os"
    "time"
	"sync/atomic"

    "github.com/zirachw/Tubes2_SeleniumSoup4/internal/scraper"
    "github.com/zirachw/Tubes2_SeleniumSoup4/internal/search"
)

var (
    flagElement = flag.String("e", "",  "element to search for (case‐insensitive)")
    flagPaths   = flag.Uint64("p", 1,     "max unique paths")
    flagUpdates = flag.Bool("u", false, "stream live updates")
    flagOutput  = flag.String("o", "",  "optional output JSON file")
)

type ResultData struct {
    Element       string      `json:"element"`
    UniquePaths   uint64      `json:"uniquePaths"`
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

	var counter uint64
	nextID := func() uint64 { return atomic.AddUint64(&counter, 1) }

    var paths []*search.Element
    var nodesExplored uint64
    
    // Call the updated BFS function that returns the node count
    paths, nodesExplored, err = search.BFS(recipeMap, *flagElement, *flagPaths)

    if err != nil {
        log.Fatalf("BFS search error: %v", err)
    }

    var tree *search.Target
    if *flagUpdates {
        updates := make(chan search.Update, 100)
        go func() {
            tree = search.CreateFullTree(paths, updates, *flagElement, nextID)
            close(updates)
        }()
        fmt.Println("⏳ Streaming BFS events:")
        for evt := range updates {
            fmt.Printf(
                "  → Stage=%-15s Info=%s\n",
                evt.Stage, evt.Info,
            )
        }
    } else {
        tree = search.CreateFullTree(paths, nil, *flagElement, nextID)
    }

    elapsed := time.Since(start)

    if tree == nil || tree.UniquePaths == 0 {
        fmt.Fprintf(os.Stderr, "Element %q not found or no paths\n", *flagElement)
    }

    rootEl := &search.Element{
        Name:    tree.Name,
        Tier:    tree.Tier,
        Recipes: tree.Recipes,
        ID:      tree.ID,
    }

    fmt.Println("\n📖 Final BFS Recipe Tree:")
    search.PrintRecipeTree(rootEl, "")

    nodesInTree := search.CountTreeNodes(rootEl)

    fmt.Printf("\nNodes explored during BFS: %d\n", nodesExplored)
    fmt.Printf("Nodes in final tree: %d\n", nodesInTree)
    fmt.Printf("Unique paths found: %d\n", tree.UniquePaths)
    fmt.Printf("Time taken: %v\n", elapsed)

    out := ResultData{
        Element:       *flagElement,
        UniquePaths:   tree.UniquePaths,
        TimeTaken:     elapsed.String(),
        NodesExplored: nodesExplored,
        NodesInTree:   nodesInTree,
        RecipeTree:    tree,
    }
    buf, err := json.MarshalIndent(out, "", "  ")
    if err != nil {
        log.Fatalf("Error marshaling JSON: %v", err)
    }
    if *flagOutput != "" {
        if err := os.WriteFile(*flagOutput, buf, 0644); err != nil {
            log.Fatalf("Error writing %s: %v", *flagOutput, err)
        }
        fmt.Printf("\nResults written to %s\n", *flagOutput)
    } else {
        fmt.Println(string(buf))
    }
}