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
    flagPaths   = flag.Int("p", 1,     "max unique paths")
    flagUpdates = flag.Bool("u", false, "stream live updates")
    flagOutput  = flag.String("o", "",  "optional output JSON file")
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

	var counter uint64
	nextID := func() uint64 { return atomic.AddUint64(&counter, 1) }

    // 1) Run BFS or BFSParallel
    var paths []*search.Element
    if *flagPaths <= 1 {
        paths, err = search.BFS(recipeMap, *flagElement, *flagPaths)
    } else {
        paths, err = search.BFSParallel(recipeMap, *flagElement, *flagPaths)
    }
    if err != nil {
        log.Fatalf("BFS search error: %v", err)
    }

    // 2) Build full tree, streaming updates if requested
    var tree *search.Target
    if *flagUpdates {
        updates := make(chan search.Update, 100)
        go func() {
            tree = search.CreateFullTree(paths, updates, nextID)
            close(updates)
        }()
        fmt.Println("⏳ Streaming BFS events:")
        for evt := range updates {
            fmt.Printf(
                "  → Stage=%-15s Elem=%-10s Tier=%2d Recipe#=%2d Info=%s\n",
                evt.Stage, evt.ElementName, evt.Tier, evt.RecipeIndex, evt.Info,
            )
        }
    } else {
        tree = search.CreateFullTree(paths, nil, nextID)
    }

    elapsed := time.Since(start)

    // 3) If we still didn’t find anything, warn
    if tree == nil || tree.UniquePaths == 0 {
        fmt.Fprintf(os.Stderr, "Element %q not found or no paths\n", *flagElement)
        // But we continue to output JSON below
    }

    // 4) Print and count nodes
    nodeCount = 0
    fmt.Println("\n📖 Final BFS Recipe Tree:")
    rootEl := &search.Element{
        Name:    tree.Name,
        Tier:    tree.Tier,
        Recipes: tree.Recipes,
        ID:      tree.ID,
    }
    printRecipeTree(rootEl, "")

    fmt.Printf("\nTotal nodes explored: %d\n", nodeCount)
    fmt.Printf("Unique paths found: %d\n", tree.UniquePaths)
    fmt.Printf("Time taken: %v\n", elapsed)

    // 5) Emit JSON
    out := ResultData{
        Element:       *flagElement,
        UniquePaths:   tree.UniquePaths,
        TimeTaken:     elapsed.String(),
        NodesExplored: nodeCount,
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
