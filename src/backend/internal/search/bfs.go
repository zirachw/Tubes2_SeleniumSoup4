package search

import (
	"fmt"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/zirachw/Tubes2_SeleniumSoup4/internal/scraper"
)

var bfsNodeExplored uint64

func GetBFSNodeExplored() uint64 {
	return atomic.LoadUint64(&bfsNodeExplored)
}
func ResetBFSNodeExplored() {
	atomic.StoreUint64(&bfsNodeExplored, 0)
}

/**
 *  BFS runs a bottom-up DP: it builds up to maxPaths example subtrees for
 *  Every element in ascending tier order, then returns the ones for targetName.
 */
func BFS(recipeMap map[string]scraper.ElementData, targetName string, maxPaths int) ([]*Element, error) {
	ResetBFSNodeExplored()

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
		atomic.AddUint64(&bfsNodeExplored, 1)
	}

	// 4) DP upwards through tiers
	for _, tier := range tiers {
		if tier == 0 {
			continue
		}
		for _, name := range byTier[tier] {
			atomic.AddUint64(&bfsNodeExplored, 1)
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
						atomic.AddUint64(&bfsNodeExplored, 1)
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
	ResetBFSNodeExplored()

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
		atomic.AddUint64(&bfsNodeExplored, 1)
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

				atomic.AddUint64(&bfsNodeExplored, 1)
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
							atomic.AddUint64(&bfsNodeExplored, 1)
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
 *  It emits Update events when 'updates' is non-nil.
 */
func CreateFullTree(
    paths []*Element,
    updates chan<- Update,
	name string,
    nextID func() uint64,
) *Target {
    // 1) Build the root placeholder (ID=0 for now)
    tgt := &Target{
        Name:        name,
        Tier:        0,
        Recipes:     nil,
        UniquePaths: len(paths),
        ID:          0,
    }
    if len(paths) > 0 {
        tgt.Name = paths[0].Name
        tgt.Tier = paths[0].Tier
    }

    // 2) Clone everything with ID=0 (we'll assign real IDs in BFS)
    var clone func(src *Element) *Element
    clone = func(src *Element) *Element {
        node := &Element{
            Name:    src.Name,
            Tier:    src.Tier,
            Recipes: nil,
            ID:      0,
        }
        for _, r := range src.Recipes {
            L := clone(r.Left)
            R := clone(r.Right)
            node.Recipes = append(node.Recipes, Recipe{Left: L, Right: R})
        }
        return node
    }

    // 3) Merge top-level recipes by signature "Left|Right"
    seen := make(map[string]int) // sig -> index in tgt.Recipes
    for _, p := range paths {
        rootClone := clone(p)
        for _, rec := range rootClone.Recipes {
            sig := rec.Left.Name + "|" + rec.Right.Name
            if idx, ok := seen[sig]; !ok {
                seen[sig] = len(tgt.Recipes)
                tgt.Recipes = append(tgt.Recipes, rec)
            } else {
                existing := &tgt.Recipes[idx]
                existing.Right.Recipes = append(existing.Right.Recipes, rec.Right.Recipes...)
            }
        }
    }

    // 4) Now do BFS, assigning IDs in order and emitting updates
    // Assign root its ID
    tgt.ID = nextID()
    if updates != nil {
        updates <- Update{
            Stage:       "startBFS",
            ElementName: tgt.Name,
            Tier:        tgt.Tier,
            Info:        fmt.Sprintf("root id=%d, merging %d paths", tgt.ID, len(paths)),
        }
    }

    type item struct {
        ParentID    uint64
        RecipeIndex int
        Node        *Element
    }

    queue := []item{}

    // Enqueue top-level children, assigning their IDs immediately
    for i, rec := range tgt.Recipes {
        // left child
        rec.Left.ID = nextID()
        rec.Right.ID = nextID()

        if updates != nil {
            updates <- Update{
                Stage:       "startRecipe",
                ElementName: tgt.Name,
                Tier:        tgt.Tier,
                ParentID:    tgt.ID,
                RecipeIndex: i,
                LeftID:      rec.Left.ID,
                RightID:     rec.Right.ID,
                LeftLabel:   rec.Left.Name,
                RightLabel:  rec.Right.Name,
                Info:        fmt.Sprintf("recipe %d under parent %d", i, tgt.ID),
            }
        }
        queue = append(queue, item{ParentID: rec.Left.ID, RecipeIndex: i, Node: rec.Left})
        queue = append(queue, item{ParentID: rec.Right.ID, RecipeIndex: i, Node: rec.Right})
    }

    // Process the rest of the BFS
    for head := 0; head < len(queue); head++ {
        cur := queue[head]
        for ci, childRec := range cur.Node.Recipes {
            // Assign IDs just before emitting
            childRec.Left.ID = nextID()
            childRec.Right.ID = nextID()

            if updates != nil {
                updates <- Update{
                    Stage:       "startRecipe",
                    ElementName: cur.Node.Name,
                    Tier:        cur.Node.Tier,
                    ParentID:    cur.Node.ID,
                    RecipeIndex: ci,
                    LeftID:      childRec.Left.ID,
                    RightID:     childRec.Right.ID,
                    LeftLabel:   childRec.Left.Name,
                    RightLabel:  childRec.Right.Name,
                    Info:        fmt.Sprintf("recipe %d under parent %d", ci, cur.Node.ID),
                }
            }
            queue = append(queue, item{ParentID: childRec.Left.ID, RecipeIndex: ci, Node: childRec.Left})
            queue = append(queue, item{ParentID: childRec.Right.ID, RecipeIndex: ci, Node: childRec.Right})
        }
    }

    if updates != nil {
        updates <- Update{
            Stage:       "doneRecipe",
            ElementName: tgt.Name,
            Tier:        tgt.Tier,
            Info:        "completed BFS traversal",
        }
    }
    return tgt
}