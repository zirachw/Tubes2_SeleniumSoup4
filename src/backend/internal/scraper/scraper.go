package scraper

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "strconv"
    "strings"

    "github.com/PuerkitoBio/goquery"
)

/**
 *  ElementData holds recipes, image link, and tier for an element.
 */ 
type ElementData struct {
    Tier      int        `json:"tier"`
    ImageLink string     `json:"imageLink"`
    Recipes   [][2]string `json:"recipes"`
}

/**
 *  Run executes the scraper for Little Alchemy 2 with optional M&M update.
 *  If updateMAM is true, it rescans and caches the Myths & Monsters dataset;
 *  otherwise it loads from the existing cache. Returns the LA2 element map.
 */
func Run(updateMAM bool) map[string]ElementData {
    const mamCache = "data/myths_and_monsters.json"
    var mamMap map[string]ElementData

    // 1) Myths & Monsters
    if updateMAM {
        mamURL := "https://little-alchemy.fandom.com/wiki/Elements_(Myths_and_Monsters)"
        mamMap = scrapePage(mamURL, "table.list-table")
        if err := writeJSON(mamCache, mamMap); err != nil {
            log.Fatalf("failed to write M&M cache: %v", err)
        }
        fmt.Printf("Cached %d Myths & Monsters elements\n", len(mamMap))
    } else {
        file, err := os.Open(mamCache)
        if err != nil {
            log.Fatalf("failed to open M&M cache: %v", err)
        }
        defer file.Close()
        decoder := json.NewDecoder(file)
        mamMap = make(map[string]ElementData)
        if err := decoder.Decode(&mamMap); err != nil {
            log.Fatalf("failed to load M&M cache: %v", err)
        }
        fmt.Printf("Loaded %d Myths & Monsters elements from cache\n", len(mamMap))
    }

    // Build filter of M&M names
    filter := make(map[string]struct{}, len(mamMap))
    for name := range mamMap {
        filter[name] = struct{}{}
    }

    // 2) Little Alchemy 2 raw scrape
    la2URL := "https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)"
    laMap := scrapePage(la2URL, "table.list-table.col-list.icon-hover")

    // Filter out recipes containing any M&M element
    for name, data := range laMap {
        var cleaned [][2]string
        for _, pair := range data.Recipes {
            if _, skip := filter[pair[0]]; skip {
                continue
            }
            if _, skip := filter[pair[1]]; skip {
                continue
            }
            cleaned = append(cleaned, pair)
        }
        data.Recipes = cleaned
        laMap[name] = data
    }

    // Save raw LA2 output
    if err := writeJSON("data/recipes.json", laMap); err != nil {
        log.Fatalf("failed to write LA2 JSON: %v", err)
    }
    fmt.Printf("Wrote %d Little Alchemy 2 elements\n", len(laMap))

    return laMap
}

/**
 *  scrapePage fetches the HTML from the given URL and extracts elements
 */
func scrapePage(url, tableSel string) map[string]ElementData {
    resp, err := http.Get(url)
    if err != nil {
        log.Fatalf("request error %s: %v", url, err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        log.Fatalf("bad status %d for %s", resp.StatusCode, url)
    }
    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        log.Fatalf("parse error %s: %v", url, err)
    }

    result := make(map[string]ElementData)
    doc.Find("span.mw-headline").Each(func(_ int, sel *goquery.Selection) {
        id, _ := sel.Attr("id")
        var tier int
        switch {
        case id == "Starting_elements":
            tier = 0
        case strings.HasPrefix(id, "Tier_"):
            parts := strings.Split(id, "_")
            if n, err := strconv.Atoi(parts[1]); err == nil {
                tier = n
            } else {
                return
            }
        default:
            return
        }

        tbl := sel.Closest("h3").NextAll().Filter(tableSel).First()
        if tbl.Length() == 0 {
            return
        }

        tbl.Find("tbody tr").Each(func(i int, row *goquery.Selection) {
            if i == 0 {
                return
            }
            cell := row.Find("td").Eq(0)
            name := strings.TrimSpace(cell.Find("a[title]").First().Text())
            if name == "" || name == "Time" || name == "Ruins" || name == "Archeologist" {
                return
            }
            img, _ := cell.Find("a.mw-file-description.image").Attr("href")

            var recipes [][2]string
            row.Find("td").Eq(1).Find("ul li").Each(func(_ int, li *goquery.Selection) {
                var ings []string
                li.Find("a[title]").Each(func(_ int, s *goquery.Selection) {
                    txt := strings.TrimSpace(s.Text())
                    if txt != "" && txt != "Time" && txt != "Ruins" && txt != "Archeologist" {
                        ings = append(ings, txt)
                    }
                })
                if len(ings) >= 2 {
                    recipes = append(recipes, [2]string{ings[0], ings[1]})
                }
            })

            result[name] = ElementData{Tier: tier, ImageLink: img, Recipes: recipes}
        })
    })

    return result
}

/** 
 *  writeJSON writes the provided data map to the given filepath in pretty JSON.
 */ 
func writeJSON(path string, data map[string]ElementData) error {
    file, err := os.Create(path)
    if err != nil {
        return err
    }
    defer file.Close()
    enc := json.NewEncoder(file)
    enc.SetIndent("", "  ")
    return enc.Encode(data)
}
