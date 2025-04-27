package scraper

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "strings"
	"time"

    "github.com/PuerkitoBio/goquery"
)

// ElementData holds the recipes and image link for an element
type ElementData struct {
    Recipes   [][]string `json:"recipes"`
    ImageLink string     `json:"imageLink"`
}

// Run scrapes the main Elements index and writes recipes.json.
func Run() map[string]ElementData {
    const indexURL = "https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)"
    // 1) Fetch index page
    req, _ := http.NewRequest("GET", indexURL, nil)
    req.Header.Set("User-Agent", "Mozilla/5.0")
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        log.Fatalf("fetch index: %v", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        log.Fatalf("bad status %d", resp.StatusCode)
    }

    doc, err := goquery.NewDocumentFromReader(resp.Body)
    if err != nil {
        log.Fatalf("parse index: %v", err)
    }

    table := doc.Find("table.list-table.col-list.icon-hover")
    if table.Length() == 0 {
        log.Fatal("index table not found")
    }

    // Create a map with element name as key and ElementData as value
    all := make(map[string]ElementData)
    
    // 2) Iterate rows
    table.Find("tbody tr").Each(func(i int, row *goquery.Selection) {
        if i == 0 {
            return // skip header
        }

        // 2a) Element name & URL
        cell := row.Find("td").Eq(0)
        link := cell.Find("a[title]").First()
        name := strings.TrimSpace(link.Text())
        if name == "" {
            return
        }
        
        // Skip "Time", "Ruins", and "Archeologist" elements
        if name == "Time" || name == "Ruins" || name == "Archeologist" {
            fmt.Printf("→ Skipping %s (excluded by exception rule)\n", name)
            return
        }

        // 2b) Image link from index cell
        imgHref, _ := cell.Find("a.mw-file-description.image").Attr("href")

        // 2c) Recipes from second cell
        var recipes [][]string
        row.Find("td").Eq(1).Find("ul li").Each(func(_ int, li *goquery.Selection) {
            // only select anchors that actually name an element
            var ings []string
            li.Find("a[title]").Each(func(_ int, s *goquery.Selection) {
                txt := strings.TrimSpace(s.Text())
                if txt != "" {
                    ings = append(ings, txt)
                }
            })
            // we only care about the first two real ingredients
            // Skip recipes that contain "Time" as an ingredient
            if len(ings) >= 2 && ings[0] != "Time" && ings[1] != "Time" {
                recipes = append(recipes, []string{ings[0], ings[1]})
            }
        })

        // Add the element and its data to the map
        all[name] = ElementData{
            Recipes:   recipes,
            ImageLink: imgHref,
        }
        
        // fmt.Printf("→ %s: %d recipes, image %s\n", name, len(recipes), imgHref)
    })

    // 3) Write JSON output
    out, err := os.Create("data/recipes.json")
    if err != nil {
        log.Fatalf("create file: %v", err)
    }
    defer out.Close()

    enc := json.NewEncoder(out)
    enc.SetIndent("", "  ")
    if err := enc.Encode(all); err != nil {
        log.Fatalf("encode JSON: %v", err)
    }

    fmt.Printf("Done: wrote %d elements to data/recipes.json\n", len(all))

	// Print timestamp of the scraping process
	fmt.Println("Scraping completed at:", time.Now().Format(time.RFC1123))

	return all
}