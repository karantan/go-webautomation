package browser

import (
	"fmt"

	"github.com/playwright-community/playwright-go"
)

func GetBrowser(headless bool) (playwright.Browser, error) {
	// Use local chrome
	runOption := &playwright.RunOptions{
		SkipInstallBrowsers: true,
	}
	err := playwright.Install(runOption)
	if err != nil {
		return nil, fmt.Errorf("could not install playwright dependencies: %v", err)
	}

	// Initialize Playwright
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not start Playwright: %w", err)
	}
	// defer pw.Stop()

	// Launch a new browser instance
	option := playwright.BrowserTypeLaunchOptions{
		Channel:  playwright.String("chrome"),
		Headless: playwright.Bool(headless),
	}

	return pw.Chromium.Launch(option)
}
