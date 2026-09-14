package browser_test

import (
	"context"
	"fmt"
	"log"

	"github.com/ToufiqQureshi/Scraper/browser"
)

func Example() {
	ctx := context.Background()

	b, err := browser.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	page, err := b.Open(ctx, "https://example.com")
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
