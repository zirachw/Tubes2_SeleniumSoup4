package search

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/zirachw/Tubes2_SeleniumSoup4/internal/scraper"
)

/**
 *  A memoized recursive DFS returning how many full root → leaf chains an element can produce.
 *  It also updates the provided nodeCounter for tracking node exploration.
 */
func countPaths(recipeData map[string]scraper.ElementData, name string, nodeCounter *uint64) int {
	atomic.AddUint64(nodeCounter, 1)
	memo := make(map[string]int)
	var dfs func(string) int
	dfs = func(el string) int {
		atomic.AddUint64(nodeCounter, 1)
		if v, ok := memo[el]; ok {
			return v
		}
		data := recipeData[el]
		if data.Tier == 0 {
			memo[el] = 1
			return 1
		}
		sum := 0
		for _, rec := range data.Recipes {
			ln, rn := rec[0], rec[1]
			ld, lok := recipeData[ln]
			rd, rok := recipeData[rn]

			// skip if missing, tier-too-high, or uncraftable side
			if !lok || !rok ||
				ld.Tier >= data.Tier || rd.Tier >= data.Tier ||
				(lok && ld.Tier != 0 && len(ld.Recipes) == 0) ||
				(rok && rd.Tier != 0 && len(rd.Recipes) == 0) {
				continue
			}

			sum += dfs(ln) * dfs(rn)
		}
		memo[el] = sum
		return sum
	}
	return dfs(name)
}

/**
 *  BuildSubtreeInternal invokes DFSSearchInternal on one element up to maxPaths,
 *  Streaming its own nested events, and returns the built *Element + how many paths.
 */
func buildSubtreeInternal(
	recipeData map[string]scraper.ElementData,
	elementName string,
	maxPaths int,
	updateCh chan<- Update,
	nextID func() uint64,
	forcedID uint64,
	nodeCounter *uint64,
) (*Element, int) {
	var subtree Tree
	DFS(recipeData, elementName, maxPaths, &subtree, updateCh, nextID, forcedID, nodeCounter)
	e := &Element{Name: subtree.Name, Tier: subtree.Tier, Recipes: subtree.Recipes, ID: subtree.ID}
	return e, subtree.UniquePaths
}

/**
 *  DFS is the main search function that builds a recipe tree for the target element.
 *  It returns the number of nodes explored during the search.
 */
func DFS(
	recipeData map[string]scraper.ElementData,
	targetName string,
	maxUniquePaths int,
	outPtr *Tree,
	updateCh chan<- Update,
	nextID func() uint64,
	forcedID uint64,
	nodeCounter *uint64,
) uint64 {
	data, ok := recipeData[targetName]
	if !ok {
		if updateCh != nil {
			updateCh <- Update{Stage: "error", ElementName: targetName, Tier: 0, Info: "not found"}
		}
		*outPtr = nil
		return *nodeCounter
	}
	id := forcedID
	if id == 0 {
		id = nextID()
	}
	root := &Target{Name: targetName, Tier: data.Tier, ID: id}
	*outPtr = root

	atomic.AddUint64(nodeCounter, 1)
	if updateCh != nil {
		updateCh <- Update{Stage: "startDFS", ElementName: targetName, Tier: data.Tier}
	}

	// —— Phase 1: parallel countPaths with pre-filter ——
	numWorkers := runtime.NumCPU()
	type countJob struct {
		idx         int
		left, right string
	}
	type countRes struct {
		idx, leftCnt, rightCnt int
	}

	jobsCount := make(chan countJob, len(data.Recipes))
	resultsCount := make(chan countRes, len(data.Recipes))
	var wgCount sync.WaitGroup
	wgCount.Add(numWorkers)
	for w := 0; w < numWorkers; w++ {
		go func() {
			defer wgCount.Done()
			for job := range jobsCount {
				lc := countPaths(recipeData, job.left, nodeCounter)
				rc := countPaths(recipeData, job.right, nodeCounter)
				resultsCount <- countRes{job.idx, lc, rc}
			}
		}()
	}

	// enqueue only craftable‐direct recipes
	for i, rec := range data.Recipes {
		ln, rn := rec[0], rec[1]
		ld, lok := recipeData[ln]
		rd, rok := recipeData[rn]
		if !lok || !rok ||
			ld.Tier >= data.Tier || rd.Tier >= data.Tier ||
			(lok && ld.Tier != 0 && len(ld.Recipes) == 0) ||
			(rok && rd.Tier != 0 && len(rd.Recipes) == 0) {
			continue
		}
		jobsCount <- countJob{i, ln, rn}
	}
	close(jobsCount)
	wgCount.Wait()
	close(resultsCount)

	type recInfo struct {
		left, right           string
		leftCount, rightCount int
	}
	infos := make([]recInfo, 0, len(data.Recipes))
	for res := range resultsCount {
		rec := data.Recipes[res.idx]
		infos = append(infos, recInfo{rec[0], rec[1], res.leftCnt, res.rightCnt})
	}

	// —— Phase 2: parallel buildSubtreeInternal with shallow-leaf ——
	type buildJob struct {
		idx                   int
		leftName, rightName   string
		leftCount, rightCount int
		take                  int
	}
	type buildRes struct {
		idx                   int
		builtLeft, builtRight *Element
		usedPaths             int
	}

	jobsBuild := make(chan buildJob, len(infos))
	resultsBuild := make(chan buildRes, len(infos))
	var wgBuild sync.WaitGroup
	wgBuild.Add(numWorkers)
	for w := 0; w < numWorkers; w++ {
		go func() {
			defer wgBuild.Done()
			for job := range jobsBuild {
				parentID := root.ID
				leftID := nextID()
				rightID := nextID()

				if updateCh != nil {
					updateCh <- Update{
						Stage:       "startRecipe",
						ElementName: targetName,
						Tier:        data.Tier,
						RecipeIndex: job.idx,
						Info:        fmt.Sprintf("%s + %s → take %d", job.leftName, job.rightName, job.take),
						ParentID:    parentID,
						LeftID:      leftID,
						RightID:     rightID,
						LeftLabel:   job.leftName,
						RightLabel:  job.rightName,
					}
				}

				var leftElem, rightElem *Element
				var used int

				if job.leftCount == 0 || job.rightCount == 0 {
					ld := recipeData[job.leftName]
					rd := recipeData[job.rightName]
					leftElem = &Element{Name: job.leftName, Tier: ld.Tier, ID: leftID}
					rightElem = &Element{Name: job.rightName, Tier: rd.Tier, ID: rightID}
					used = 0

				} else {
					full := job.leftCount * job.rightCount
					var leftNeed, rightNeed int
					if job.take == full {
						leftNeed, rightNeed = job.leftCount, job.rightCount
					} else {
						leftNeed = (job.take + job.rightCount - 1) / job.rightCount
						if leftNeed > job.leftCount {
							leftNeed = job.leftCount
						}
						rightNeed = job.take
					}

					leftElem, _ = buildSubtreeInternal(recipeData, job.leftName, leftNeed, updateCh, nextID, leftID, nodeCounter)
					rightElem, _ = buildSubtreeInternal(recipeData, job.rightName, rightNeed, updateCh, nextID, rightID, nodeCounter)

					used = job.take
				}

				resultsBuild <- buildRes{job.idx, leftElem, rightElem, used}
			}
		}()
	}

	// Schedule build jobs (always include the recipe, but prune to remaining)
	remaining := maxUniquePaths
	for idx, info := range infos {
		if remaining <= 0 {
			break
		}
		total := info.leftCount * info.rightCount
		take := total
		if take > remaining {
			take = remaining
		}
		remaining -= take
		jobsBuild <- buildJob{idx, info.left, info.right, info.leftCount, info.rightCount, take}
	}
	close(jobsBuild)
	wgBuild.Wait()
	close(resultsBuild)

	// Collect & attach
	for br := range resultsBuild {
		root.Recipes = append(root.Recipes, Recipe{Left: br.builtLeft, Right: br.builtRight})
		root.UniquePaths += br.usedPaths
	}

	return *nodeCounter
}
