package search

import (
	// "encoding/json"
	// "sync"

	"github.com/zirachw/Tubes2_SeleniumSoup4/internal/scraper"
)

func UpdateSolutions(tree *Tree) {
	(*tree).NSolution = 0

	for _, child := range (*tree).Children {
		if child.Left.NSolution > 0 && child.Right.NSolution > 0 {
			(*tree).NSolution += child.Left.NSolution * child.Right.NSolution
		}
	}
}

func BFSSearch(recipe map[string]scraper.ElementData, element string,
	nSolution int, solved map[string]Tree, tree *Tree) {
	if recipe[element].Tier == 0 {
		(*tree).NSolution = 1
		return
	}

	_, ok := recipe[element]
	if !ok {
		*tree = nil
		print("gaada\n")
		return
	}

	type QueueItem struct {
		TreeNode *Tree
		Element  string
	}

	queue := []QueueItem{}

	queue = append(queue, QueueItem{tree, element})

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		currentTree := current.TreeNode
		currentElement := current.Element

		if currentTree.NSolution > 0 {
			continue
		}

		currentData, ok := recipe[currentElement]
		if !ok {
			continue
		}

		candidates := currentData.Recipes

		// Try each candidate recipe
		for _, candidate := range candidates {
			(*currentTree).NVisited += 1

			// Skip if ingredients have higher tier than the current element
			if recipe[candidate[0]].Tier > currentData.Tier || recipe[candidate[1]].Tier > currentData.Tier {
				continue
			}

			// Create trees for left and right children
			left := makeTree(candidate[0], solved)
			left.Up = *currentTree
			right := makeTree(candidate[1], solved)
			right.Up = *currentTree

			// Add them to the current tree's children
			(*currentTree).Children = append((*currentTree).Children, Child{left, right})

			// Process left child
			if left.NSolution == 0 {
				// Check if it's already in solved map
				if solvedLeft, exists := solved[left.Value]; exists && solvedLeft.NSolution > 0 {
					left.NSolution = solvedLeft.NSolution
					left.NVisited = solvedLeft.NVisited
				} else {
					// Add to queue for BFS processing
					queue = append(queue, QueueItem{&left, left.Value})
				}
			}
			(*currentTree).NVisited += left.NVisited

			// Process right child
			if right.NSolution == 0 {
				// Check if it's already in solved map
				if solvedRight, exists := solved[right.Value]; exists && solvedRight.NSolution > 0 {
					right.NSolution = solvedRight.NSolution
					right.NVisited = solvedRight.NVisited
				} else {
					// Add to queue for BFS processing
					queue = append(queue, QueueItem{&right, right.Value})
				}
			}
			(*currentTree).NVisited += right.NVisited

			// Update solutions for current node if both children have solutions
			if right.NSolution > 0 && left.NSolution > 0 {
				(*currentTree).NSolution += right.NSolution * left.NSolution

				// Add to solved map
				solved[left.Value] = left
				solved[right.Value] = right
			}

			// Early termination if we found enough solutions
			if (*currentTree).NSolution >= nSolution {
				return
			}
		}
	}
}
