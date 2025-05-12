package search

import (
	"fmt"
	"runtime"
	"sort"
	"sync"

	"github.com/zirachw/Tubes2_SeleniumSoup4/internal/scraper"
)

/**
 *  BFS runs a bottom-up DP: it builds up to maxPaths example subtrees for
 *  Every element in ascending tier order, then returns the ones for targetName.
 */
func BFS(recipeMap map[string]scraper.ElementData, targetName string, maxPaths int) ([]*Element, error) {

	// 1) Group elements by tier
	byTier := make(map[int][]string, len(recipeMap))
	var tiers []int
	for name, data := range recipeMap {
		t := data.Tier
		if _, ok := byTier[t]; !ok {
			tiers = append(tiers, t)
		}
		byTier[t] = append(byTier[t], name)
	}
	sort.Ints(tiers)

	// 2) memo[name] = up to maxPaths built subtrees for that element
	memo := make(map[string][]*Element, len(recipeMap))

	// 3) Initialize base elements (tier 0)
	for _, name := range byTier[0] {
		memo[name] = []*Element{{Name: name, Tier: 0}}
	}

	// 4) DP upwards through tiers
	for _, tier := range tiers {
		if tier == 0 {
			continue
		}
		for _, name := range byTier[tier] {
			data := recipeMap[name]
			var examples []*Element

			// Try each recipe that makes `name`
			for _, pair := range data.Recipes {
				leftList, lok := memo[pair[0]]
				rightList, rok := memo[pair[1]]

				if !lok || !rok || len(leftList) == 0 || len(rightList) == 0 {
					continue
				}

				for _, L := range leftList {
					for _, R := range rightList {
						if len(examples) >= maxPaths {
							break
						}

						if data.Tier <= L.Tier || data.Tier <= R.Tier {
							continue
						}
						examples = append(examples, &Element{
							Name: name,
							Tier: data.Tier,
							Recipes: []Recipe{{
								Left:  L,
								Right: R,
							}},
						})
					}
					if len(examples) >= maxPaths {
						break
					}
				}
				if len(examples) >= maxPaths {
					break
				}
			}

			memo[name] = examples
		}
	}

	// 5) Return up to maxPaths for the target
	result, ok := memo[targetName]
	if !ok {
		return nil, nil
	}
	if len(result) > maxPaths {
		result = result[:maxPaths]
	}
	return result, nil
}

/**
 *  BFSParallel runs a bottom-up DP with per-tier concurrency,
 *  building up to maxPaths example sub-trees per element,
 *  then returns up to maxPaths for targetName.
 */
func BFSParallel(recipeMap map[string]scraper.ElementData, targetName string, maxPaths int) ([]*Element, error) {

	// 1) bucket names by tier
	byTier := make(map[int][]string, len(recipeMap))
	var tiers []int
	for name, data := range recipeMap {
		t := data.Tier
		if _, ok := byTier[t]; !ok {
			tiers = append(tiers, t)
		}
		byTier[t] = append(byTier[t], name)
	}
	sort.Ints(tiers)

	// 2) memo holds up to maxPaths examples per element
	memo := make(map[string][]*Element, len(recipeMap))

	// 3) seed tier 0
	for _, name := range byTier[0] {
		memo[name] = []*Element{{Name: name, Tier: 0}}
	}

	// 4) process higher tiers in parallel per element
	for _, tier := range tiers {
		if tier == 0 {
			continue
		}
		nextMemo := make(map[string][]*Element, len(byTier[tier]))
		var wg sync.WaitGroup
		var mu sync.Mutex
		sem := make(chan struct{}, runtime.NumCPU())

		for _, name := range byTier[tier] {
			wg.Add(1)
			sem <- struct{}{}
			go func(el string) {
				defer wg.Done()
				defer func() { <-sem }()

				data := recipeMap[el]
				var examples []*Element

				for _, pair := range data.Recipes {
					leftList, lok := memo[pair[0]]
					rightList, rok := memo[pair[1]]
					if !lok || !rok || len(leftList) == 0 || len(rightList) == 0 {
						continue
					}
					for _, L := range leftList {
						for _, R := range rightList {
							if len(examples) >= maxPaths {
								break
							}
							if data.Tier <= L.Tier || data.Tier <= R.Tier {
								continue
							}
							examples = append(examples, &Element{
								Name:    el,
								Tier:    data.Tier,
								Recipes: []Recipe{{Left: L, Right: R}},
							})
						}
						if len(examples) >= maxPaths {
							break
						}
					}
					if len(examples) >= maxPaths {
						break
					}
				}

				mu.Lock()
				nextMemo[el] = examples
				mu.Unlock()
			}(name)
		}

		wg.Wait()

		// merge this tier into memo
		for name, ex := range nextMemo {
			memo[name] = ex
		}
	}

	// 5) return up to maxPaths for target
	result, ok := memo[targetName]
	if !ok || len(result) == 0 {
		return nil, nil
	}
	if len(result) > maxPaths {
		result = result[:maxPaths]
	}
	return result, nil
}

/**
 *  CreateFullTree takes the raw []*Element paths for the same target
 *  and builds a single *Target containing all their top-level recipes.
 *  It emits Update events when ‘updates’ is non-nil.
 */
func CreateFullTree(paths []*Element, updates chan<- Update, nextID func() uint64) *Target {
    // 1) Build root Target as before
    rootID := nextID()
    tgt := &Target{Name: "", Tier: 0, Recipes: nil, UniquePaths: len(paths), ID: rootID}
    if len(paths) == 0 {
        return tgt
    }
    tgt.Name = paths[0].Name
    tgt.Tier = paths[0].Tier
    if updates != nil {
        updates <- Update{Stage:"createTarget", ElementName:tgt.Name, Tier:tgt.Tier, Info:fmt.Sprintf("target id=%d",rootID)}
    }

    // 2) Clone memo to share identical subtrees
    cloneMemo := make(map[*Element]*Element)
    var clone func(src *Element) *Element
    clone = func(src *Element) *Element {
        if c, ok := cloneMemo[src]; ok {
            return c
        }
        id := nextID()
        node := &Element{Name: src.Name, Tier: src.Tier, ID: id}
        cloneMemo[src] = node
        if updates != nil {
            updates <- Update{Stage:"createNode", ElementName:node.Name, Tier:node.Tier, Info:fmt.Sprintf("node id=%d",id)}
        }
        for idx, r := range src.Recipes {
            L := clone(r.Left)
            R := clone(r.Right)
            if updates != nil {
                updates <- Update{
                    Stage:       "createRecipe",
                    ElementName: node.Name,
                    Tier:        node.Tier,
                    RecipeIndex: idx,
                    ParentID:    node.ID,
                    LeftID:      L.ID,
                    RightID:     R.ID,
                    Info:        fmt.Sprintf("recipe %d for node %d", idx, node.ID),
                }
            }
            node.Recipes = append(node.Recipes, Recipe{Left: L, Right: R})
        }
        return node
    }

    // 3) Attach top‐level recipes **merging by parent signature**
    seen := make(map[string]int) // signature -> index in tgt.Recipes
    for pidx, rootSrc := range paths {
        clonedRoot := clone(rootSrc)
        for _, rec := range clonedRoot.Recipes {
            sig := rec.Left.Name + "|" + rec.Right.Name
            idx, exists := seen[sig]
            if !exists {
                // first time we see this parent combo: attach whole Recipe
                tgt.Recipes = append(tgt.Recipes, rec)
                idx = len(tgt.Recipes) - 1
                seen[sig] = idx
                if updates != nil {
                    updates <- Update{
                        Stage:       "attachTopRecipe",
                        ElementName: tgt.Name,
                        Tier:        tgt.Tier,
                        RecipeIndex: idx,
                        ParentID:    tgt.ID,
                        LeftID:      rec.Left.ID,
                        RightID:     rec.Right.ID,
                        Info:        fmt.Sprintf("path %d first attach of %s", pidx, sig),
                    }
                }
            } else {
                // merge this child recipe into the existing right.Recipes
                existing := &tgt.Recipes[idx]
                // append all recipes under rec.Right into existing.Right.Recipes
                for _, sub := range rec.Right.Recipes {
                    existing.Right.Recipes = append(existing.Right.Recipes, sub)
                    if updates != nil {
                        updates <- Update{
                            Stage:       "mergeRecipe",
                            ElementName: tgt.Name,
                            Tier:        tgt.Tier,
                            RecipeIndex: idx,
                            ParentID:    tgt.ID,
                            LeftID:      rec.Left.ID,
                            RightID:     rec.Right.ID,
                            Info:        fmt.Sprintf("path %d merge into %s", pidx, sig),
                        }
                    }
                }
            }
        }
    }

    return tgt
}