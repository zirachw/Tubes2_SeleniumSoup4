package main

import (
	"github.com/zirachw/Tubes2_SeleniumSoup4/internal/scraper"
)

func main() {
	scraper.Hello()

	result, err := scraper.TestRequest("https://www.google.com")
	if err != nil {
		panic(err)
	}
	println(result)

}
