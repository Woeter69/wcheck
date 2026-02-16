package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

// ScanResult holds the findings of a single page scan
type ScanResult struct {
	URL    string
	Errors []string
}

// Engine handles the browser lifecycle and scanning
type Engine struct {
	AllocCtx context.Context
	Cancel   context.CancelFunc
}

// NewEngine initializes a new headless browser engine
func NewEngine() *Engine {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoSandbox,
		chromedp.DisableGPU,
	)
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	
	return &Engine{
		AllocCtx: allocCtx,
		Cancel:   allocCancel,
	}
}

// Close shuts down the browser engine
func (e *Engine) Close() {
	if e.Cancel != nil {
		e.Cancel()
	}
}

// ScanPage navigates to a URL and monitors for errors
func (e *Engine) ScanPage(url string, verbose bool) ([]string, error) {
	// Create a new tab/context for this specific scan to ensure thread-safety
	ctx, cancel := chromedp.NewContext(e.AllocCtx)
	defer cancel()

	// Set a timeout for the entire page load
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var (
		mu     sync.Mutex
		errors []string
	)

	addErr := func(msg string) {
		mu.Lock()
		errors = append(errors, msg)
		mu.Unlock()
	}

	// Set up event listeners BEFORE we run the navigation
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *runtime.EventExceptionThrown:
			addErr(fmt.Sprintf("JS Exception: %s", ev.ExceptionDetails.Text))
		case *runtime.EventConsoleAPICalled:
			if ev.Type == runtime.APITypeError {
				var msg string
				for _, arg := range ev.Args {
					if arg.Value != nil {
						msg += fmt.Sprintf("%v ", arg.Value)
					}
				}
				addErr(fmt.Sprintf("Console Error: %s", strings.TrimSpace(msg)))
			}
		case *network.EventLoadingFailed:
			// Optionally ignore canceled requests (e.g., if JS aborts a fetch)
			if ev.Canceled {
				return
			}
			addErr(fmt.Sprintf("Network Failure: %s (%s)", ev.ErrorText, ev.Type))
		}
	})

	// Execute scan: Enable domains, Navigate, Wait, and Settle
	err := chromedp.Run(ctx,
		network.Enable(),
		runtime.Enable(),
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second), // Settle period for async JS/Network errors
	)

	if err != nil {
		return nil, err
	}

	return errors, nil
}

// ExtractLinks finds all <a> tags and returns their href attributes
func (e *Engine) ExtractLinks(url string) ([]string, error) {
	ctx, cancel := chromedp.NewContext(e.AllocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var hrefs []string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.Evaluate(`Array.from(document.querySelectorAll('a')).map(a => a.href)`, &hrefs),
	)

	return hrefs, err
}
