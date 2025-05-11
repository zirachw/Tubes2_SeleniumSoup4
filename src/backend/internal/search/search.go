package search

import (
	"encoding/json"
	"sync"

	"github.com/zirachw/Tubes2_SeleniumSoup4/internal/scraper"
)

type Child struct {
	Left  Tree
	Right Tree
}

type Root struct {
	Value     string
	Children  []Child
	Up        *Root
	NSolution int
	NVisited  int
}

type Tree *Root

type SRoot struct {
	Value     string   `json:"value"`
	Children  []SChild `json:"children,omitempty"`
	NSolution int      `json:"nSolution"`
}

type SChild struct {
	Left  *SRoot `json:"left,omitempty"`
	Right *SRoot `json:"right,omitempty"`
}

func SerializeTree(t Tree) *SRoot {
	if t == nil {
		return nil
	}

	children := make([]SChild, len(t.Children))
	for i, child := range t.Children {
		children[i] = SChild{
			Left:  SerializeTree(child.Left),
			Right: SerializeTree(child.Right),
		}
	}

	return &SRoot{
		Value:     t.Value,
		NSolution: t.NSolution,
		Children:  children,
	}
}

func TreeToJSON(t Tree) ([]byte, error) {
	serializable := SerializeTree(t)
	bytes, err := json.MarshalIndent(serializable, "", "  ")
	if err != nil {
		return make([]byte, 0), err
	}
	return bytes, nil
}

// func Contains(a []string, x string) bool {
// 	for _, n := range a {
// 		if x == n {
// 			return true
// 		}
// 	}
// 	return false
// }

func deepCopy(t Tree, parent Tree) Tree {
	if t == nil {
		return nil
	}
	newRoot := &Root{
		Value:     t.Value,
		NSolution: t.NSolution,
		Up:        parent,
	}
	newChildren := make([]Child, len(t.Children))
	for i, child := range t.Children {
		newChildren[i] = Child{
			Left:  deepCopy(child.Left, newRoot),
			Right: deepCopy(child.Right, newRoot),
		}
	}
	newRoot.Children = newChildren

	return newRoot
}

func makeTree(element string, solved map[string]Tree) Tree {
	solutionTree, ok := solved[element]
	if ok {
		return deepCopy(solutionTree, nil)
	} else {
		return &Root{element, nil, nil, 0, 0}
	}
}

func DFSSearch(recipe map[string]scraper.ElementData, element string,
	nSolution int, solved map[string]Tree, tree *Tree) {
	if recipe[element].Tier == 0 {
		(*tree).NSolution = 1
		return
	}

	data, ok := recipe[element]
	if !ok {
		*tree = nil
		print("gaada\n")
		return
	}

	candidates := data.Recipes

	for _, candidate := range candidates {
		(*tree).NVisited += 1
		if recipe[candidate[0]].Tier > data.Tier || recipe[candidate[1]].Tier > data.Tier {
			continue
		}

		left := makeTree(candidate[0], solved)
		left.Up = *tree
		right := makeTree(candidate[1], solved)
		right.Up = *tree

		(*tree).Children = append((*tree).Children, Child{left, right})

		if left.NSolution == 0 {
			DFSSearch(recipe, left.Value, nSolution, solved, &left)
		}
		(*tree).NVisited += left.NVisited

		if left.NSolution == 0 {
			continue
		} else {
			solved[left.Value] = left
		}

		if right.NSolution == 0 {
			DFSSearch(recipe, right.Value, nSolution, solved, &right)
		}
		(*tree).NVisited += right.NVisited

		if right.NSolution == 0 {
			continue
		} else {
			solved[right.Value] = right
		}

		if right.NSolution > 0 && left.NSolution > 0 {
			(*tree).NSolution += right.NSolution * left.NSolution
		}

		if (*tree).NSolution >= nSolution {
			return
		}
	}
}

func BFSSearch(recipe map[string]scraper.ElementData, element string,
	nSolution int, solved map[string]Tree, tree *Tree) {
}

func ParallelDFSSearch(recipe map[string]scraper.ElementData, element string,
	nSolution int, solved map[string]Tree, tree *Tree) {

	// If nSolution is 1 or less, use the regular DFS search
	if nSolution <= 1 {
		DFSSearch(recipe, element, nSolution, solved, tree)
		return
	}

	// Base case for tier 0 elements
	if recipe[element].Tier == 0 {
		(*tree).NSolution = 1
		return
	}

	data, ok := recipe[element]
	if !ok {
		*tree = nil
		print("gaada\n")
		return
	}

	candidates := data.Recipes

	// Use a mutex to protect the shared solved map and tree.nSolution
	var solvedMutex sync.Mutex
	var treeMutex sync.Mutex
	var wg sync.WaitGroup

	// Process candidates in parallel
	for _, candidate := range candidates {
		if recipe[candidate[0]].Tier > data.Tier || recipe[candidate[1]].Tier > data.Tier {
			continue
		}

		// Create a local copy of candidate for the goroutine
		currentCandidate := candidate

		// Add to children first, under tree mutex protection
		solvedMutex.Lock()
		left := makeTree(currentCandidate[0], solved)
		left.Up = *tree
		right := makeTree(currentCandidate[1], solved)
		right.Up = *tree

		treeMutex.Lock()
		(*tree).Children = append((*tree).Children, Child{left, right})
		treeMutex.Unlock()
		solvedMutex.Unlock()

		// Start a goroutine for this candidate path
		wg.Add(1)
		go func(left Tree, right Tree, candidate [2]string) {
			defer wg.Done()

			// Local copied variables for thread safety
			localLeft := left
			localRight := right

			// Process left child
			if localLeft.NSolution == 0 {
				// Create a thread-local copy of the solved map
				solvedMutex.Lock()
				localSolved := make(map[string]Tree)
				for k, v := range solved {
					localSolved[k] = v
				}
				solvedMutex.Unlock()

				// Process left branch
				DFSSearch(recipe, localLeft.Value, nSolution, localSolved, &localLeft)

				// Update the shared solved map with results from this thread
				if localLeft.NSolution > 0 {
					solvedMutex.Lock()
					solved[localLeft.Value] = localLeft
					solvedMutex.Unlock()
				}
			}

			if localLeft.NSolution == 0 {
				return
			}

			// Process right child
			if localRight.NSolution == 0 {
				// Create a thread-local copy of the solved map
				solvedMutex.Lock()
				localSolved := make(map[string]Tree)
				for k, v := range solved {
					localSolved[k] = v
				}
				solvedMutex.Unlock()

				// Process right branch
				DFSSearch(recipe, localRight.Value, nSolution, localSolved, &localRight)

				// Update the shared solved map with results from this thread
				if localRight.NSolution > 0 {
					solvedMutex.Lock()
					solved[localRight.Value] = localRight
					solvedMutex.Unlock()
				}
			}

			if localRight.NSolution == 0 {
				return
			}

			// Calculate solutions and update the tree
			if localRight.NSolution > 0 && localLeft.NSolution > 0 {
				treeMutex.Lock()
				(*tree).NSolution += localRight.NSolution * localLeft.NSolution
				// _bool := (*tree).nSolution >= nSolution
				treeMutex.Unlock()

				// If we have enough solutions, we can mark this to exit early
				// But we still need to let all goroutines complete
			}

		}(left, right, candidate)

		// Check if we already have enough solutions before starting more goroutines
		treeMutex.Lock()
		if (*tree).NSolution >= nSolution {
			treeMutex.Unlock()
			break
		}
		treeMutex.Unlock()
	}

	// Wait for all goroutines to complete
	wg.Wait()
}

// DFSSearchWrapper decides which implementation to use based on nSolution
func DFSSearchWrapper(recipe map[string]scraper.ElementData, element string,
	nSolution int, solved map[string]Tree, tree *Tree) {
	if nSolution > 1 {
		ParallelDFSSearch(recipe, element, nSolution, solved, tree)
	} else {
		DFSSearch(recipe, element, nSolution, solved, tree)
	}
}
