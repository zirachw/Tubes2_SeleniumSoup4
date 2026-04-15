package search

import (
	"fmt"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/zirachw/Tubes2_SeleniumSoup4/internal/scraper"
)

/**
 *  BFS runs a bottom-up DP approach to find paths for the target element.
 *  It returns both the paths found and the number of nodes explored.
 */
func BFS(
	recipeMap map[string]scraper.ElementData,
	targetName string,
	maxPaths uint64,
) ([]*Element, uint64, error) {

	// Use a local node counter instead of a global variable
	var nodeCounter uint64 = 0

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

	memo := make(map[string][]*Element, len(recipeMap))

	// Initialize tier 0 elements
	for _, name := range byTier[0] {
		memo[name] = []*Element{{Name: name, Tier: 0}}
		atomic.AddUint64(&nodeCounter, 1)
	}

	// Process higher tiers
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

				atomic.AddUint64(&nodeCounter, 1)
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
                            if uint64(len(examples)) >= maxPaths {
                                break
                            }
							if data.Tier <= L.Tier || data.Tier <= R.Tier {
								continue
							}
							atomic.AddUint64(&nodeCounter, 1)
							examples = append(examples, &Element{
								Name:    el,
								Tier:    data.Tier,
								Recipes: []Recipe{{Left: L, Right: R}},
							})
						}
						if uint64(len(examples)) >= maxPaths {
							break
						}
					}
					if uint64(len(examples)) >= maxPaths {
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

	// Return up to maxPaths for target
	result, ok := memo[targetName]
	if !ok || len(result) == 0 {
		return nil, nodeCounter, nil
	}
	if uint64(len(result)) > maxPaths {
		result = result[:maxPaths]
	}
	return result, nodeCounter, nil
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
    tgt := &Target{
        Name:        name,
        Tier:        0,
        Recipes:     nil,
        UniquePaths: uint64(len(paths)),
        ID:          0,
    }
    if len(paths) > 0 {
        tgt.Name = paths[0].Name
        tgt.Tier = paths[0].Tier
    }

    // deep-clone function
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

    // helper: remove duplicated recipes by signature, merge their subtrees
    dedupeRecipes := func(recs []Recipe) []Recipe {
        uniq := make([]Recipe, 0, len(recs))
        seen := make(map[string]int, len(recs))
        for _, r := range recs {
            sig := r.Left.Name + "|" + r.Right.Name
            if idx, ok := seen[sig]; !ok {
                seen[sig] = len(uniq)
                uniq = append(uniq, r)
            } else {
                ex := &uniq[idx]
                ex.Left.Recipes  = append(ex.Left.Recipes,  r.Left.Recipes...)
                ex.Right.Recipes = append(ex.Right.Recipes, r.Right.Recipes...)
            }
        }
        return uniq
    }

    // —— merge top-level paths into tgt.Recipes (as before) ——
    seen := make(map[string]int, len(paths))
    for _, p := range paths {
        root := clone(p)
        for _, rec := range root.Recipes {
            sig := rec.Left.Name + "|" + rec.Right.Name
            if idx, ok := seen[sig]; !ok {
                seen[sig] = len(tgt.Recipes)
                tgt.Recipes = append(tgt.Recipes, rec)
            } else {
                ex := &tgt.Recipes[idx]
                // merge both sides
                ex.Left.Recipes  = append(ex.Left.Recipes,  rec.Left.Recipes...)
                ex.Right.Recipes = append(ex.Right.Recipes, rec.Right.Recipes...)
            }
        }
    }

    // assign root ID + initial update
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
    queue := make([]item, 0, len(tgt.Recipes)*2)

    // seed queue with root’s immediate children
    for i, rec := range tgt.Recipes {

        rec.Left.Recipes  = dedupeRecipes(rec.Left.Recipes)
        rec.Right.Recipes = dedupeRecipes(rec.Right.Recipes)

        rec.Left.ID  = nextID()
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
        queue = append(queue,
            item{ParentID: rec.Left.ID, RecipeIndex: i, Node: rec.Left},
            item{ParentID: rec.Right.ID, RecipeIndex: i, Node: rec.Right},
        )
    }

    // BFS out through every node
    for head := 0; head < len(queue); head++ {
        cur := queue[head]

        // —— de-dupe this node’s recipes —— 
        cur.Node.Recipes = dedupeRecipes(cur.Node.Recipes)

        for ci, childRec := range cur.Node.Recipes {
			
            // de-dupe grandchildren as well before assigning IDs
            childRec.Left.Recipes  = dedupeRecipes(childRec.Left.Recipes)
            childRec.Right.Recipes = dedupeRecipes(childRec.Right.Recipes)

            childRec.Left.ID  = nextID()
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
            queue = append(queue,
                item{ParentID: childRec.Left.ID, RecipeIndex: ci, Node: childRec.Left},
                item{ParentID: childRec.Right.ID, RecipeIndex: ci, Node: childRec.Right},
            )
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
