package scraper_test

import (
	"context"
	"fmt"
	"log"

	"github.com/ToufiqQureshi/Scraper"
)

func Example() {
	ctx := context.Background()

	browser, err := scraper.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer browser.Close()

	page, err := browser.Open(ctx, "https://example.com")
	if err != nil {
		log.Fatal(err)
	}
	defer page.Close()

	text, err := page.Text(ctx, "h1")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(text)
}
