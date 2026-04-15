package search

import (
	"fmt"
)

/**
 *  Target element we’re searching for.
 */
type Target struct {
	Name        string   `json:"name"`
	Tier        int      `json:"tier"`
	Recipes     []Recipe `json:"recipes"`
	UniquePaths uint64    `json:"uniquePaths"`
	ID          uint64   `json:"id"`
}

/**
 *  Target is the final JSON tree.
 *  It contains the full recipe tree for a given element.
 */
type Tree = *Target

/**
 *  Element is a node in the final JSON tree.
 */
type Element struct {
	Name    string   `json:"name"`
	Tier    int      `json:"tier"`
	Recipes []Recipe `json:"recipes,omitempty"`
	ID      uint64   `json:"id"`
}

/**
 *  Recipe pairs two child Elements.
 */
type Recipe struct {
	Left  *Element `json:"left"`
	Right *Element `json:"right"`
}

/**
 *  Update events get streamed as we build.
 */
type Update struct {
	Stage       string // e.g. "startDFS", "startRecipe", "startBuildLeft", ...
	ElementName string
	Tier        int    // actual tier of ElementName
	RecipeIndex int    // which recipe of the *parent* we’re on
	Info        string // any extra detail

	ParentID uint64
	LeftID   uint64
	RightID  uint64

	LeftLabel  string
	RightLabel string
}

// Counts nodes in a recipe tree
func CountTreeNodes(el *Element) int {
    if el == nil {
        return 0
    }

    count := 1 // Count this node
    for _, r := range el.Recipes {
        count += CountTreeNodes(r.Left)
        count += CountTreeNodes(r.Right)
    }
    return count
}

// Prints a recipe tree structure
func PrintRecipeTree(el *Element, indent string) int {
    if el == nil {
        return 0
    }
    fmt.Printf("%s%s (tier=%d, id=%d)\n", indent, el.Name, el.Tier, el.ID)

    if len(el.Recipes) == 0 {
        return 1
    }
    total := 0
    for i, r := range el.Recipes {
        fmt.Printf("%s  Recipe %d:\n", indent, i+1)
        fmt.Printf("%s    Left ingredient:\n", indent)
        lp := PrintRecipeTree(r.Left, indent+"      ")
        fmt.Printf("%s    Right ingredient:\n", indent)
        rp := PrintRecipeTree(r.Right, indent+"      ")
        contrib := lp * rp
        total += contrib
        fmt.Printf("%s  Recipe %d contributes %d path(s)\n", indent, i+1, contrib)
    }
    return total
}