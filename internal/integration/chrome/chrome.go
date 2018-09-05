// Package chrome implements some helpers for WebRTC integration testing with Chrome
package chrome

import (
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/wirepair/gcd"
	"github.com/wirepair/gcd/gcdapi"
)

type Chrome struct {
	dbg        *gcd.Gcd
	serverPort string
}

func New() *Chrome {
	res := &Chrome{
		serverPort: "9090",
	}

	return res
}

// Spawn an instance of chrome for testing
func (c *Chrome) Spawn() error {
	c.dbg = gcd.NewChromeDebugger()

	c.dbg.AddFlags([]string{"--headless", "--disable-gpu"})
	err := c.dbg.StartProcess(c.getProcessDetails())
	if err != nil {
		return fmt.Errorf("failed to start Chrome: %v", err)
	}

	return nil
}

// Page opens a page in a new browser tab
func (c *Chrome) Page(page string) error {
	if !strings.HasPrefix(page, "http://") {
		page = "http://localhost:" + c.serverPort + "/" + page
	}

	target, err := c.dbg.NewTab()
	if err != nil {
		return fmt.Errorf("failed to open new tab: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	target.Subscribe("Page.loadEventFired", func(targ *gcd.ChromeTarget, v []byte) {
		wg.Done()
	})

	// get the Page API and enable it
	if _, err := target.Page.Enable(); err != nil {
		return fmt.Errorf("failed to enable page notifications: %v", err)
	}

	navigateParams := &gcdapi.PageNavigateParams{Url: page}
	_, _, _, err = target.Page.NavigateWithParams(navigateParams)
	if err != nil {
		return fmt.Errorf("failed to navigate: %v", err)
	}

	// ensure the page is loaded
	wg.Wait()

	return nil
}

// getProcessDetails gets chrome process details based on the runtime.GOOS
func (c *Chrome) getProcessDetails() (exePath, userDir, port string) {
	var path string
	var dir string

	switch runtime.GOOS {
	case "windows":
		path = "C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe"
		dir = "C:\\temp\\"

	case "darwin":
		path = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
		dir = "/tmp/"

	case "linux":
		path = "/usr/bin/chromium-browser"
		dir = "/tmp/"
	}

	return path, dir, "9222"
}

// Close closes the
func (c *Chrome) Close() error {
	if c.dbg == nil {
		return nil
	}
	err := c.dbg.ExitProcess()
	if err != nil {
		return err
	}
	c.dbg = nil
	return nil
}
