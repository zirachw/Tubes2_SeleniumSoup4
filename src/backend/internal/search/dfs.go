package search

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/zirachw/Tubes2_SeleniumSoup4/internal/scraper"
)

/**
 *  A memoized recursive DFS returning how many full root → leaf chains an element can produce.
 */
func countPaths(recipeData map[string]scraper.ElementData, name string) int {
	memo := map[string]int{}
	var dfs func(string) int
	dfs = func(el string) int {
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
			lName, rName := rec[0], rec[1]
			lData, lok := recipeData[lName]
			rData, rok := recipeData[rName]
			if !lok || !rok || lData.Tier >= data.Tier || rData.Tier >= data.Tier {
				continue
			}
			sum += dfs(lName) * dfs(rName)
		}
		memo[el] = sum
		return sum
	}
	return dfs(name)
}

/**
 *  DFS builds up to maxUniquePaths full paths.
 *  It only follows recipes whose ingredient tiers are strictly lower.
 */
func DFS(
	recipeMap map[string]scraper.ElementData,
	rootName string,
	maxUniquePaths int,
	outPtr *Tree,
	nextID func() uint64,
) {

	data, ok := recipeMap[rootName]
	if !ok {
		fmt.Printf("element %q not found\n", rootName)
		*outPtr = nil
		return
	}
	root := &Target{
		Name:        rootName,
		Tier:        data.Tier,
		Recipes:     nil,
		UniquePaths: 0,
		ID:          nextID(),
	}
	*outPtr = root

	if data.Tier == 0 {
		root.UniquePaths = 1
		return
	}

	for _, rec := range data.Recipes {
		if root.UniquePaths >= maxUniquePaths {
			break
		}

		ln, rn := rec[0], rec[1]
		ld, lok := recipeMap[ln]
		rd, rok := recipeMap[rn]
		if !lok || !rok || ld.Tier >= data.Tier || rd.Tier >= data.Tier {
			continue
		}

		// how many full paths each side can yield?
		leftCount := countPaths(recipeMap, ln)
		rightCount := countPaths(recipeMap, rn)
		totalPaths := leftCount * rightCount

		remain := maxUniquePaths - root.UniquePaths

		var leftElem, rightElem *Element
		var used int

		if totalPaths <= remain {
			leftElem, _ = buildSubtree(recipeMap, ln, leftCount, nextID)
			rightElem, _ = buildSubtree(recipeMap, rn, rightCount, nextID)
			used = totalPaths

		} else {

			// partial: only take exactly `remain` paths out of this recipe
			// ceil(remain / rightCount) distinct left sub‐paths
			leftNeeded := (remain + rightCount - 1) / rightCount
			if leftNeeded > leftCount {
				leftNeeded = leftCount
			}
			leftElem, _ = buildSubtree(recipeMap, ln, leftNeeded, nextID)

			// for right, only actually need `remain` leaves in total,
			// but all sit under the *first* leftNeeded leaf-branch we built,
			// so only need at most `remain` right subpaths:
			rightNeeded := remain
			if rightNeeded > rightCount {
				rightNeeded = rightCount
			}
			rightElem, _ = buildSubtree(recipeMap, rn, rightNeeded, nextID)

			used = remain
		}

		// attach this (possibly‐pruned) recipe
		root.Recipes = append(root.Recipes, Recipe{
			Left:  leftElem,
			Right: rightElem,
		})
		root.UniquePaths += used
	}
}

/**
 *  DFSParallel is a parallelized version of DFS.
 *  It uses a worker pool to count paths and build subtrees.
 *  Returns the built tree and the number of unique paths.
 */
func DFSParallel(
	recipeData map[string]scraper.ElementData,
	targetName string,
	maxUniquePaths int,
	outTree *Tree,
	nextID func() uint64,
) {
	rootData, exists := recipeData[targetName]
	if !exists {
		fmt.Printf("element %q not found\n", targetName)
		*outTree = nil
		return
	}

	root := &Target{Name: targetName, Tier: rootData.Tier, ID: nextID()}
	*outTree = root

	if rootData.Tier == 0 {
		root.UniquePaths = 1
		return
	}

	numWorkers := runtime.NumCPU()

	// -- Phase 1: parallel countPaths via worker pool --

	type countJob struct {
		index               int
		leftName, rightName string
	}
	type countResult struct {
		index                         int
		leftPathCount, rightPathCount int
	}

	jobsCount := make(chan countJob, len(rootData.Recipes))
	resultsCount := make(chan countResult, len(rootData.Recipes))

	var wgCount sync.WaitGroup

	// launch numWorkers count‐workers
	wgCount.Add(numWorkers)
	for w := 0; w < numWorkers; w++ {
		go func() {
			defer wgCount.Done()
			for job := range jobsCount {
				lc := countPaths(recipeData, job.leftName)
				rc := countPaths(recipeData, job.rightName)
				resultsCount <- countResult{
					index:          job.index,
					leftPathCount:  lc,
					rightPathCount: rc,
				}
			}
		}()
	}

	for idx, rec := range rootData.Recipes {
		jobsCount <- countJob{idx, rec[0], rec[1]}
	}
	close(jobsCount)
	wgCount.Wait()
	close(resultsCount)

	type recipeInfo struct {
		leftName, rightName           string
		leftPathCount, rightPathCount int
	}
	infos := make([]recipeInfo, len(rootData.Recipes))
	for res := range resultsCount {
		rec := rootData.Recipes[res.index]
		infos[res.index] = recipeInfo{
			leftName:       rec[0],
			rightName:      rec[1],
			leftPathCount:  res.leftPathCount,
			rightPathCount: res.rightPathCount,
		}
	}

	// -- Phase 2: parallel subtree‐build via worker pool --

	type buildJob struct {
		index                 int
		leftName, rightName   string
		leftCount, rightCount int
		take                  int
	}
	type buildResult struct {
		index                 int
		builtLeft, builtRight *Element
		usedPaths             int
	}

	jobsBuild := make(chan buildJob, len(infos))
	resultsBuild := make(chan buildResult, len(infos))

	var wgBuild sync.WaitGroup
	wgBuild.Add(numWorkers)
	for w := 0; w < numWorkers; w++ {
		go func() {
			defer wgBuild.Done()
			for job := range jobsBuild {
				var leftNeed, rightNeed int
				if job.take == job.leftCount*job.rightCount {
					leftNeed, rightNeed = job.leftCount, job.rightCount
				} else {
					leftNeed = (job.take + job.rightCount - 1) / job.rightCount
					if leftNeed > job.leftCount {
						leftNeed = job.leftCount
					}
					rightNeed = job.take
				}

				leftTree, _ := buildSubtree(recipeData, job.leftName, leftNeed, nextID)
				rightTree, _ := buildSubtree(recipeData, job.rightName, rightNeed, nextID)

				resultsBuild <- buildResult{
					index:      job.index,
					builtLeft:  leftTree,
					builtRight: rightTree,
					usedPaths:  job.take,
				}
			}
		}()
	}

	// schedule build jobs in order, respecting remaining quota
	remaining := maxUniquePaths
	for idx, info := range infos {
		if remaining <= 0 {
			break
		}
		ld, lok := recipeData[info.leftName]
		rd, rok := recipeData[info.rightName]
		if !lok || !rok || ld.Tier >= rootData.Tier || rd.Tier >= rootData.Tier {
			continue
		}
		totalPaths := info.leftPathCount * info.rightPathCount
		take := totalPaths
		if take > remaining {
			take = remaining
		}
		remaining -= take

		jobsBuild <- buildJob{
			index:      idx,
			leftName:   info.leftName,
			rightName:  info.rightName,
			leftCount:  info.leftPathCount,
			rightCount: info.rightPathCount,
			take:       take,
		}
	}
	close(jobsBuild)
	wgBuild.Wait()
	close(resultsBuild)

	buildMap := make(map[int]buildResult, len(infos))
	for br := range resultsBuild {
		buildMap[br.index] = br
	}
	for i := 0; i < len(infos); i++ {
		if br, ok := buildMap[i]; ok {
			root.Recipes = append(root.Recipes,
				Recipe{Left: br.builtLeft, Right: br.builtRight},
			)
			root.UniquePaths += br.usedPaths
		}
	}
}

/**
 *  DFSWithUpdates kicks off the DFS in a goroutine,
 *  returns a channel you can range over for live updates.
 */
func DFSWithUpdates(
	recipeData map[string]scraper.ElementData,
	rootName string,
	maxUniquePaths int,
	outPtr *Tree,
	nextID func() uint64,
) <-chan Update {
	updates := make(chan Update, 100)
	go func() {
		DFSInternal(recipeData, rootName, maxUniquePaths, outPtr, updates, nextID, 0)
		if *outPtr != nil {
			updates <- Update{
				Stage:       "completeDFS",
				ElementName: rootName,
				Tier:        (*outPtr).Tier,
				RecipeIndex: 0,
				Info:        fmt.Sprintf("UniquePaths=%d", (*outPtr).UniquePaths),
			}
		}
		close(updates)
	}()
	return updates
}

/**
 *  DFSInternal is the same worker-pool DFS as before,
 *  But peppered with updateCh<-Update calls at key moments.
 */
func DFSInternal(
	recipeData map[string]scraper.ElementData,
	targetName string,
	maxUniquePaths int,
	outPtr *Tree,
	updateCh chan<- Update,
	nextID func() uint64,
	forcedID uint64,
) {
	data, ok := recipeData[targetName]
	if !ok {
		updateCh <- Update{Stage: "error", ElementName: targetName, Tier: 0, Info: "not found"}
		*outPtr = nil
		return
	}

	id := forcedID
	if id == 0 {
		id = nextID()
	}
	root := &Target{Name: targetName, Tier: data.Tier, ID: id}
	*outPtr = root

	updateCh <- Update{
		Stage:       "startDFS",
		ElementName: targetName,
		Tier:        data.Tier,
	}

	if data.Tier == 0 {
		root.UniquePaths = 1
		updateCh <- Update{
			Stage:       "foundBase",
			ElementName: targetName,
			Tier:        data.Tier,
		}
		return
	}

	numWorkers := runtime.NumCPU()
	type countJob struct {
		idx         int
		left, right string
	}
	type countRes struct{ idx, leftCnt, rightCnt int }

	jobsCount := make(chan countJob, len(data.Recipes))
	resultsCount := make(chan countRes, len(data.Recipes))

	var wgCount sync.WaitGroup
	wgCount.Add(numWorkers)
	for w := 0; w < numWorkers; w++ {
		go func() {
			defer wgCount.Done()
			for job := range jobsCount {
				lc := countPaths(recipeData, job.left)
				rc := countPaths(recipeData, job.right)
				resultsCount <- countRes{job.idx, lc, rc}
			}
		}()
	}
	for i, rec := range data.Recipes {
		jobsCount <- countJob{i, rec[0], rec[1]}
	}
	close(jobsCount)
	wgCount.Wait()
	close(resultsCount)

	type recInfo struct {
		left, right           string
		leftCount, rightCount int
	}
	infos := make([]recInfo, len(data.Recipes))
	for res := range resultsCount {
		rec := data.Recipes[res.idx]
		infos[res.idx] = recInfo{rec[0], rec[1], res.leftCnt, res.rightCnt}
	}

	updateCh <- Update{
		Stage:       "startPhase2",
		ElementName: targetName,
		Tier:        data.Tier,
		Info:        "building recipes",
	}

	type buildJob struct {
		idx                   int
		leftName, rightName   string
		leftCount, rightCount int
		take                  int
	}
	type buildRes struct {
		idx                 int
		leftElem, rightElem *Element
		usedPaths           int
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

				var leftNeed, rightNeed int
				full := job.leftCount * job.rightCount
				if job.take == full {
					leftNeed, rightNeed = job.leftCount, job.rightCount
				} else {
					leftNeed = (job.take + job.rightCount - 1) / job.rightCount
					if leftNeed > job.leftCount {
						leftNeed = job.leftCount
					}
					rightNeed = job.take
				}

				updateCh <- Update{
					Stage:       "startBuildLeft",
					ElementName: job.leftName,
					Tier:        recipeData[job.leftName].Tier,
					RecipeIndex: job.idx,
				}
				leftElem, _ := buildSubtreeInternal(recipeData, job.leftName, leftNeed, updateCh, nextID, leftID)
				updateCh <- Update{
					Stage:       "endBuildLeft",
					ElementName: job.leftName,
					Tier:        recipeData[job.leftName].Tier,
					RecipeIndex: job.idx,
				}

				updateCh <- Update{
					Stage:       "startBuildRight",
					ElementName: job.rightName,
					Tier:        recipeData[job.rightName].Tier,
					RecipeIndex: job.idx,
				}
				rightElem, _ := buildSubtreeInternal(recipeData, job.rightName, rightNeed, updateCh, nextID, rightID)
				updateCh <- Update{
					Stage:       "endBuildRight",
					ElementName: job.rightName,
					Tier:        recipeData[job.rightName].Tier,
					RecipeIndex: job.idx,
				}

				updateCh <- Update{
					Stage:       "doneRecipe",
					ElementName: targetName,
					Tier:        data.Tier,
					RecipeIndex: job.idx,
					Info:        fmt.Sprintf("used %d paths", job.take),
				}

				resultsBuild <- buildRes{job.idx, leftElem, rightElem, job.take}
			}
		}()
	}

	remaining := maxUniquePaths
	for i, info := range infos {
		if remaining <= 0 {
			break
		}
		ld, lok := recipeData[info.left]
		rd, rok := recipeData[info.right]
		if !lok || !rok || ld.Tier >= data.Tier || rd.Tier >= data.Tier {
			continue
		}
		total := info.leftCount * info.rightCount
		take := total
		if take > remaining {
			take = remaining
		}
		remaining -= take
		jobsBuild <- buildJob{i, info.left, info.right, info.leftCount, info.rightCount, take}
	}
	close(jobsBuild)
	wgBuild.Wait()
	close(resultsBuild)

	for res := range resultsBuild {
		root.Recipes = append(root.Recipes, Recipe{res.leftElem, res.rightElem})
		root.UniquePaths += res.usedPaths
	}
}

/**
 *  BuildSubtree invokes DFSSearch on a sub‐element to
 *  Get exactly up to maxPaths, then returns the *Element
 *  Tree plus how many it actually yielded.
 */
func buildSubtree(
	recipeMap map[string]scraper.ElementData,
	name string,
	maxPaths int,
	nextID func() uint64,
) (*Element, int) {
	var tgt Tree
	DFS(recipeMap, name, maxPaths, &tgt, nextID)
	if tgt == nil {
		return nil, 0
	}

	e := &Element{Name: tgt.Name, Tier: tgt.Tier, ID: tgt.ID}

	e.Recipes = tgt.Recipes
	return e, tgt.UniquePaths
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
) (*Element, int) {
	var subtree Tree
	DFSInternal(recipeData, elementName, maxPaths, &subtree, updateCh, nextID, forcedID)
	e := &Element{Name: subtree.Name, Tier: subtree.Tier, Recipes: subtree.Recipes, ID: subtree.ID}
	return e, subtree.UniquePaths
}
