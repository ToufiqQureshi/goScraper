package scraper_test

import (
	"fmt"
	"log"

	"github.com/ToufiqQureshi/Scraper"
)

func Example() {
	browser, err := scraper.New()
	if err != nil {
		log.Fatal(err)
	}
	defer browser.Close()

	page, err := browser.Open("https://example.com")
	if err != nil {
		log.Fatal(err)
	}
	defer page.Close()

	text, err := page.Text("h1")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(text)
	// Output:
	// Example Domain
}
