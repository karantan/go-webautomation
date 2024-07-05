package main

import (
	"encoding/csv"
	"fmt"
	"net/url"
	"os"
	"strings"
	"webautomation/browser"

	"github.com/playwright-community/playwright-go"
)

const linkedInLogin = "https://www.linkedin.com/login"

var (
	username = "foo"
	password = "secret"
)

func linkedinLogin(page playwright.Page) {
	page.Goto(linkedInLogin)
	page.WaitForLoadState()
	// page.Pause()

	page.GetByLabel("Email or Phone").Click()
	page.GetByLabel("Email or Phone").Fill(username)
	page.GetByLabel("Password").Click()
	page.GetByLabel("Password").Fill(password)
	page.GetByLabel("Sign in", playwright.PageGetByLabelOptions{Exact: playwright.Bool(true)}).Click()

	// pause for solving the puzzle
	page.Pause()
}

var businessRoles = []string{
	"founder",
	"cofounder",
	"chief executive officer",
	"CEO",
	"director",
	"SEO",
	"head of marketing",
	"head of ecommerce",
	"marketing director",
	"marketing",
	"digital",
}

type Prospect struct {
	URL     string
	Name    string
	Title   string
	Members string
}

func getProspectName(page playwright.Page, linkedinURL string) (Prospect, error) {
	// Make sure to remove trailing /
	linkedinURL, _ = strings.CutSuffix(linkedinURL, "/")
	peopleURl := linkedinURL + "/people/"

	page.Goto(peopleURl)
	// page.Pause()

	numberOfMembers, err := page.GetByRole("heading", playwright.PageGetByRoleOptions{Name: "associated member"}).TextContent()
	if err != nil {
		return Prospect{}, err
	}
	numberOfMembers = strings.TrimSpace(numberOfMembers)
	if numberOfMembers == "0 associated members" {
		return Prospect{}, nil
	}

	page.Locator(".org-top-card-summary__title").WaitFor(playwright.LocatorWaitForOptions{Timeout: playwright.Float(2000)})
	companyName, err := page.Locator(".org-top-card-summary__title").First().TextContent()
	if err != nil {
		fmt.Printf("Error getting company name: %v\n", err)
		return Prospect{}, err
	}

	for _, role := range businessRoles {
		page.GetByPlaceholder("Search employees by title,").Click()
		page.GetByPlaceholder("Search employees by title,").Fill(fmt.Sprintf("%s %s", role, companyName))
		err := page.GetByPlaceholder("Search employees by title,").Press("Enter")
		if err != nil {
			return Prospect{}, err
		}

		// Wait 2 seconds for JS filter finishes
		page.Locator("main .artdeco-entity-lockup__title").WaitFor(playwright.LocatorWaitForOptions{Timeout: playwright.Float(2000)})
		elementNames := page.Locator("main .artdeco-entity-lockup__title")
		elCount, err := elementNames.Count()
		if err != nil {
			return Prospect{}, err
		}
		if elCount == 0 {
			page.GetByRole("button", playwright.PageGetByRoleOptions{Name: "Clear all"}).Click()
			continue
		}

		elementTitles := page.Locator("main .artdeco-entity-lockup__subtitle")

		// Check all results and filter out e.g. "LinkedIn Member"
		allNames, err := elementNames.All()
		if err != nil {
			return Prospect{}, err
		}
		desiredProspect := 0
		for i, n := range allNames {
			name, err := n.TextContent()
			if err != nil {
				fmt.Printf("Error parsing title: %v\n", err)
				continue
			}
			if strings.Contains(name, "LinkedIn Member") {
				desiredProspect = i
				continue
			}
			desiredProspect = i
			break
		}

		name, err := elementNames.Nth(desiredProspect).TextContent()
		if err != nil {
			return Prospect{}, err
		}
		title, err := elementTitles.Nth(desiredProspect).TextContent()
		if err != nil {
			return Prospect{}, err
		}

		return Prospect{Name: strings.TrimSpace(name), Title: strings.TrimSpace(title), Members: numberOfMembers}, nil
	}

	return Prospect{}, nil
}

// isValidURL validates a URL string
func isValidURL(str string) bool {
	u, err := url.Parse(str)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	return true
}

var (
	inCSV  = "source.csv"
	outCSV = "updated_with_names.csv"
)

func main() {
	b, err := browser.GetBrowser(false)
	if err != nil {
		fmt.Printf("count not open the browser: %v", err)
		return
	}
	defer b.Close()

	page, err := b.NewPage()
	if err != nil {
		fmt.Printf("could not create new page: %v", err)
		return
	}
	// set default timeout to 5sec so that we don't waste time when we cannot finde
	// some elements
	page.SetDefaultTimeout(7000)

	// Start scraping
	linkedinLogin(page)

	// Open the CSV file
	file, err := os.Open(inCSV)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Read the CSV file
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	// Extract member names
	var results []Prospect
	for _, record := range records {
		pageURL := record[1]
		baseLinkedinURL := record[6] // Change the index if required
		var linkedinURL string
		if !strings.HasPrefix(baseLinkedinURL, "https") {
			linkedinURL = "https://www." + baseLinkedinURL
		} else {
			linkedinURL = baseLinkedinURL
		}

		if !isValidURL(linkedinURL) {
			fmt.Printf("URL: %s not valid\n", linkedinURL)
			continue
		}
		fmt.Printf("Processing URL: %s\n", linkedinURL)

		prospect, err := getProspectName(page, linkedinURL)
		if err != nil {
			fmt.Printf("could not get prospect's name and title: %v", err)
			continue
		}
		prospect.URL = pageURL
		results = append(results, prospect)
	}

	// Write the results to a new CSV file
	outputFile, err := os.Create(outCSV)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// Write the updated records
	for _, result := range results {
		line := []string{result.URL, result.Name, result.Title, result.Members}
		if err := writer.Write(line); err != nil {
			fmt.Println("Error writing record to file:", err)
			return
		}
	}

	fmt.Println("Process completed successfully!")
}
