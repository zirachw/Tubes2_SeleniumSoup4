package search

import (
	"github.com/zirachw/Tubes2_SeleniumSoup4/internal/scraper"
)

type Child struct {
	left  Tree
	right Tree
}

type Root struct {
	value     string
	children  []Child
	up        *Root
	nSolution int
}

type Tree *Root

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
		value:     t.value,
		nSolution: t.nSolution,
		up:        parent,
	}
	newChildren := make([]Child, len(t.children))
	for i, child := range t.children {
		newChildren[i] = Child{
			left:  deepCopy(child.left, newRoot),
			right: deepCopy(child.right, newRoot),
		}
	}
	newRoot.children = newChildren

	return newRoot
}

func makeTree(element string, solved map[string]Tree) Tree {
	solutionTree, ok := solved[element]
	if ok {
		return deepCopy(solutionTree, nil)
	} else {
		return &Root{element, nil, nil, 0}
	}
}

func DFSSearch(recipe map[string]scraper.ElementData, element string,
	nSolution int, solved map[string]Tree, tree *Tree) {
	if recipe[element].Tier == 0 {
		(*tree).nSolution = 1
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
		if recipe[candidate[0]].Tier > data.Tier || recipe[candidate[1]].Tier > data.Tier {
			continue
		}

		left := makeTree(candidate[0], solved)
		left.up = *tree
		right := makeTree(candidate[1], solved)
		right.up = *tree

		(*tree).children = append((*tree).children, Child{left, right})

		if left.nSolution == 0 {
			DFSSearch(recipe, left.value, nSolution, solved, &left)
		}

		if left.nSolution == 0 {
			continue
		} else {
			solved[left.value] = left
		}

		if right.nSolution == 0 {
			DFSSearch(recipe, right.value, nSolution, solved, &right)
		}

		if right.nSolution == 0 {
			continue
		} else {
			solved[right.value] = right
		}

		if right.nSolution > 0 && left.nSolution > 0 {
			(*tree).nSolution += right.nSolution * left.nSolution
		}

		if (*tree).nSolution >= nSolution {
			return
		}
	}
}

func BFSSearch(recipe map[string]scraper.ElementData, element string,
	nSolution int, solved map[string]Tree, tree *Tree) {
}
