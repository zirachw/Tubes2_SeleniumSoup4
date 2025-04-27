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

type ElementData struct {
	Recipes   [][]string `json:"recipes"`
	ImageLink string     `json:"imageLink"`
}

func Run() map[string]ElementData {
	const indexURL = "https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)"
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
	all := make(map[string]ElementData)
	table.Find("tbody tr").Each(func(i int, row *goquery.Selection) {
		if i == 0 {
			return
		}
		cell := row.Find("td").Eq(0)
		link := cell.Find("a[title]").First()
		name := strings.TrimSpace(link.Text())
		if name == "" {
			return
		}
		if name == "Time" || name == "Ruins" || name == "Archeologist" {
			fmt.Printf("→ Skipping %s (excluded by exception rule)\n", name)
			return
		}
		imgHref, _ := cell.Find("a.mw-file-description.image").Attr("href")
		var recipes [][]string
		row.Find("td").Eq(1).Find("ul li").Each(func(_ int, li *goquery.Selection) {
			var ings []string
			li.Find("a[title]").Each(func(_ int, s *goquery.Selection) {
				txt := strings.TrimSpace(s.Text())
				if txt != "" {
					ings = append(ings, txt)
				}
			})
			if len(ings) >= 2 && ings[0] != "Time" && ings[1] != "Time" {
				recipes = append(recipes, []string{ings[0], ings[1]})
			}
		})
		all[name] = ElementData{
			Recipes:   recipes,
			ImageLink: imgHref,
		}
		// fmt.Printf("→ %s: %d recipes, image %s\n", name, len(recipes), imgHref)
	})
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
	fmt.Println("Scraping completed at:", time.Now().Format(time.RFC1123))

	return all
}
