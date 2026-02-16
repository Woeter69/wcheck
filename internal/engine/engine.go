package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
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
		chromedp.Flag("disable-dev-shm-usage", true),
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
		if verbose {
			fmt.Printf("\n[DEBUG] Listener caught error: %s\n", msg)
		}
	}

	// Set up event listeners BEFORE we run the navigation
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *runtime.EventExceptionThrown:
			msg := ev.ExceptionDetails.Text
			if ev.ExceptionDetails.Exception != nil && ev.ExceptionDetails.Exception.Description != "" {
				msg = ev.ExceptionDetails.Exception.Description
			}
			addErr(fmt.Sprintf("JS Exception: %s", msg))

		case *runtime.EventConsoleAPICalled:
			var msg string
			for _, arg := range ev.Args {
				if arg.Value != nil {
					msg += fmt.Sprintf("%v ", arg.Value)
				} else if arg.Description != "" {
					msg += fmt.Sprintf("%s ", arg.Description)
				}
			}
			msg = strings.TrimSpace(msg)
			
			if verbose {
				fmt.Printf("\n[DEBUG] Console %s: %s\n", ev.Type, msg)
			}

			if ev.Type == runtime.APITypeError {
				addErr(fmt.Sprintf("Console Error: %s", msg))
			}

		case *network.EventResponseReceived:
			if ev.Response.Status >= 400 {
				addErr(fmt.Sprintf("HTTP Error %d: %s", ev.Response.Status, ev.Response.URL))
			}

		case *network.EventLoadingFailed:
			// Optionally ignore canceled requests (e.g., if JS aborts a fetch)
			if ev.Canceled {
				return
			}
			addErr(fmt.Sprintf("Network Failure: %s (%s)", ev.ErrorText, ev.Type))
		}
	})

	// Wait for the page load event
	loadEventChan := make(chan struct{})
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if _, ok := ev.(*page.EventLoadEventFired); ok {
			// Check if channel is already closed to avoid panic
			select {
			case <-loadEventChan:
			default:
				close(loadEventChan)
			}
		}
	})

	// Execute scan: Enable domains, Navigate, Wait, and Settle
	err := chromedp.Run(ctx,
		network.Enable(),
		runtime.Enable(),
		page.Enable(),
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Wait for load event or context timeout
			select {
			case <-loadEventChan:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second), // Settle period for async JS/Network errors
	)

	if err != nil {
		return nil, err
	}

	mu.Lock()
	defer mu.Unlock()
	// Return a copy to avoid race conditions with lingering events
	res := make([]string, len(errors))
	copy(res, errors)
	return res, nil
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
