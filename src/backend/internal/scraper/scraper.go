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

// ElementData holds recipes, image link, and tier for an element.
type ElementData struct {
	Tier      int         `json:"tier"`
	ImageLink string      `json:"imageLink"`
	Recipes   [][2]string `json:"recipes"`
}

// Run scrapes the Elements page, captures each element's tier, recipes, and image,
// then writes the results to data/recipes.json.
func Run() map[string]ElementData {
	const indexURL = "https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)"
	req, _ := http.NewRequest("GET", indexURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("failed to fetch index: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("bad status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalf("failed to parse index: %v", err)
	}

	all := make(map[string]ElementData)

	// Iterate each section headline (Starting_elements, Tier_X_elements)
	doc.Find("span.mw-headline").Each(func(_ int, sel *goquery.Selection) {
		id, _ := sel.Attr("id")
		// Determine tier: 0 for Starting_elements, N for Tier_N_elements
		var tier int
		switch {
		case id == "Starting_elements":
			tier = 0
		case strings.HasPrefix(id, "Tier_"):
			parts := strings.Split(id, "_")
			if len(parts) >= 2 {
				if n, err := strconv.Atoi(parts[1]); err == nil {
					tier = n
				} else {
					return
				}
			} else {
				return
			}
		default:
			return
		}

		// Find the first table after this headline
		table := sel.Closest("h3").NextAll().Filter("table.list-table.col-list.icon-hover").First()
		if table.Length() == 0 {
			return
		}

		// Iterate rows in this tier table
		table.Find("tbody tr").Each(func(i int, row *goquery.Selection) {
			if i == 0 {
				return // skip header
			}
			cell := row.Find("td").Eq(0)
			link := cell.Find("a[title]").First()
			name := strings.TrimSpace(link.Text())
			if name == "" || name == "Time" || name == "Ruins" || name == "Archeologist" {
				return
			}

			imgLink, _ := cell.Find("a.mw-file-description.image").Attr("href")

			// Collect recipe pairs, skipping any "Time"
			var recipes [][2]string
			row.Find("td").Eq(1).Find("ul li").Each(func(_ int, li *goquery.Selection) {
				var ings []string
				li.Find("a[title]").Each(func(_ int, s *goquery.Selection) {
					txt := strings.TrimSpace(s.Text())
					if txt != "" && txt != "Time" {
						ings = append(ings, txt)
					}
				})
				if len(ings) >= 2 {
					recipes = append(recipes, [2]string{ings[0], ings[1]})
				}
			})

			all[name] = ElementData{
				Tier:      tier,
				ImageLink: imgLink,
				Recipes:   recipes,
			}
			fmt.Printf("→ %s (tier %d): %d recipes\n", name, tier, len(recipes))
		})
	})

	// Write output JSON
	outFile, err := os.Create("data/recipes.json")
	if err != nil {
		log.Fatalf("create file: %v", err)
	}
	defer outFile.Close()

	encoder := json.NewEncoder(outFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(all); err != nil {
		log.Fatalf("encode JSON: %v", err)
	}

	fmt.Printf("Done: wrote %d elements to data/recipes.json\n", len(all))
	return all
}
