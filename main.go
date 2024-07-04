package main

import (
	"fmt"
	"log"
	"webautomation/browser"

	"github.com/playwright-community/playwright-go"
)

func getFirstH1(url string) (string, error) {
	b, err := browser.GetBrowser(false)
	if err != nil {
		fmt.Printf("count not open the browser: %v", err)
		return "", err
	}

	defer b.Close()

	// Open a new page
	page, err := b.NewPage()
	if err != nil {
		return "", fmt.Errorf("could not create new page: %w", err)
	}

	// Navigate to the URL
	resp, err := page.Goto(url)
	if err != nil {
		return "", fmt.Errorf("could not go to URL: %w", err)
	}
	page.WaitForLoadState()
	fmt.Printf("Page loaded with %d status code.\n", resp.Status())

	// Pause to see the playwrite inspector
	page.Pause()

	page.GetByRole("link", playwright.PageGetByRoleOptions{Name: "Get started"}).Click()
	page.GetByRole("link", playwright.PageGetByRoleOptions{Name: "How to install Playwright"}).Click()
	title := page.GetByRole("heading", playwright.PageGetByRoleOptions{Name: "Installation"})

	h1Text, err := title.TextContent()
	if err != nil {
		return "", fmt.Errorf("could not get text content of <h1> element: %w", err)
	}

	return h1Text, nil
}

func main() {
	url := "https://playwright.dev/"
	h1Text, err := getFirstH1(url)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("First H1 text: %s\n", h1Text)
}
