package search

import (
	"fmt"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/zirachw/Tubes2_SeleniumSoup4/internal/scraper"
)

const (
    MAX_RECIPES_PER_ELEMENT = 50     // Max number of recipes to store per element
    MAX_LEFT_COMBINATIONS   = 20     // Max left elements to consider for combinations
    MAX_RIGHT_COMBINATIONS  = 20     // Max right elements to consider for combinations  
    MAX_GOROUTINES          = 4      // Max concurrent goroutines
    MAX_ELEMENTS_PER_TIER   = 1000   // Max elements to process in a tier
)

func BFS(
    recipeMap map[string]scraper.ElementData, 
    targetName string, 
    maxPaths int,
) ([]*Element, uint64, error) {
    var nodeCounter uint64 = 0
    
    // Group elements by tier
    byTier := make(map[int][]string, len(recipeMap))
    var tiers []int
    for name, data := range recipeMap {
        t := data.Tier
        if _, ok := byTier[t]; !ok {
            tiers = append(tiers, t)
        }
        
        // Cap elements per tier
        if len(byTier[t]) < MAX_ELEMENTS_PER_TIER {
            byTier[t] = append(byTier[t], name)
        }
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
        sem := make(chan struct{}, MAX_GOROUTINES)  // Use constant for goroutine limit
        
        for _, name := range byTier[tier] {
            wg.Add(1)
            sem <- struct{}{}
            
            go func(el string) {
                defer wg.Done()
                defer func() { <-sem }()
                
                atomic.AddUint64(&nodeCounter, 1)
                data := recipeMap[el]
                var examples []*Element
                
                // Cap number of recipes to process
                recipesToProcess := data.Recipes
                if len(recipesToProcess) > MAX_RECIPES_PER_ELEMENT {
                    recipesToProcess = recipesToProcess[:MAX_RECIPES_PER_ELEMENT]
                }
                
                for _, pair := range recipesToProcess {
                    leftList, lok := memo[pair[0]]
                    rightList, rok := memo[pair[1]]
                    if !lok || !rok || len(leftList) == 0 || len(rightList) == 0 {
                        continue
                    }
                    
                    // Cap the number of left and right elements to consider
                    leftCapped := leftList
                    if len(leftList) > MAX_LEFT_COMBINATIONS {
                        leftCapped = leftList[:MAX_LEFT_COMBINATIONS]
                    }
                    
                    rightCapped := rightList
                    if len(rightList) > MAX_RIGHT_COMBINATIONS {
                        rightCapped = rightList[:MAX_RIGHT_COMBINATIONS]
                    }
                    
                    for _, L := range leftCapped {
                        for _, R := range rightCapped {
                            if len(examples) >= maxPaths {
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
                        
                        if len(examples) >= maxPaths {
                            break
                        }
                    }
                    
                    if len(examples) >= maxPaths {
                        break
                    }
                }
                
                // Cap the number of examples to store
                if len(examples) > maxPaths {
                    examples = examples[:maxPaths]
                }
                
                if len(examples) > 0 {
                    mu.Lock()
                    nextMemo[el] = examples
                    mu.Unlock()
                }
            }(name)
        }
        
        wg.Wait()
        
        // Merge this tier into memo
        for name, ex := range nextMemo {
            memo[name] = ex
        }
        
        // Clear memory after processing each tier
        for k := range memo {
            // Keep only elements needed for next tier and the target
            if k != targetName {
                keepElement := false
                for _, name := range byTier[tier+1] {
                    for _, pair := range recipeMap[name].Recipes {
                        if pair[0] == k || pair[1] == k {
                            keepElement = true
                            break
                        }
                    }
                    if keepElement {
                        break
                    }
                }
                
                if !keepElement {
                    delete(memo, k)
                }
            }
        }
        
        // Explicitly trigger garbage collection after each tier
        runtime.GC()
    }
    
    // Return up to maxPaths for target
    result, ok := memo[targetName]
    if !ok || len(result) == 0 {
        return nil, nodeCounter, nil
    }
    
    if len(result) > maxPaths {
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
        UniquePaths: len(paths),
        ID:          0,
    }
    if len(paths) > 0 {
        tgt.Name = paths[0].Name
        tgt.Tier = paths[0].Tier
    }

    // Clone everything with ID=0 (we'll assign real IDs in BFS)
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

    // Merge top-level recipes by signature "Left|Right"
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

    // BFS, assigning IDs in order and emitting updates
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

    for i, rec := range tgt.Recipes {
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

    for head := 0; head < len(queue); head++ {
        cur := queue[head]
        for ci, childRec := range cur.Node.Recipes {
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