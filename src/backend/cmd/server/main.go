
package main

import (
	"fmt"
	"github.com/zirachw/Tubes2_SeleniumSoup4/internal/scraper"
)

func main() {
	// Run the scraper and get the data map
	data := scraper.Run()
	
	// Print out the number of elements collected
	fmt.Printf("Successfully collected data for %d elements\n", len(data))
}