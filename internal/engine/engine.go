package engine

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/debugger"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

// ErrorType defines the category of error found
type ErrorType string

const (
	TypeErrorJS      ErrorType = "JS Exception"
	TypeErrorConsole ErrorType = "Console Error"
	TypeErrorNetwork ErrorType = "Network Error"
)

// PageError holds detailed information about a single error found on a page
type PageError struct {
	Type       ErrorType
	Message    string
	StackTrace string
	URL        string
	Line       int
}

// ScanResult holds the findings of a single page scan
type ScanResult struct {
	URL    string
	Errors []PageError
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

// buildStackTrace converts a runtime.StackTrace into a string recursively to include async boundaries
func buildStackTrace(st *runtime.StackTrace) string {
	if st == nil {
		return ""
	}
	var sb strings.Builder
	for _, frame := range st.CallFrames {
		sb.WriteString(fmt.Sprintf("   at %s (%s:%d:%d)\n", frame.FunctionName, frame.URL, frame.LineNumber, frame.ColumnNumber))
	}
	if st.Parent != nil {
		sb.WriteString("   --- async boundary ---\n")
		sb.WriteString(buildStackTrace(st.Parent))
	}
	return sb.String()
}

// ScanPage navigates to a URL and monitors for errors
func (e *Engine) ScanPage(url string, verbose bool, timeout time.Duration) ([]PageError, error) {
	// Create a new tab/context for this specific scan
	ctx, cancel := chromedp.NewContext(e.AllocCtx)
	defer cancel()

	// Use the provided timeout for the entire page lifecycle
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	var (
		mu     sync.Mutex
		errors []PageError
	)

	addErr := func(err PageError) {
		mu.Lock()
		errors = append(errors, err)
		mu.Unlock()
		if verbose {
			fmt.Printf("\n[DEBUG] Listener caught %s: %s\n", err.Type, err.Message)
		}
	}

	// Set up event listeners BEFORE we run the navigation
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *runtime.EventExceptionThrown:
			msg := ev.ExceptionDetails.Text
			if ev.ExceptionDetails.Exception != nil {
				desc := ev.ExceptionDetails.Exception.Description
				if desc != "" {
					msg = strings.Split(desc, "\n")[0]
				}
			}
			
			stack := buildStackTrace(ev.ExceptionDetails.StackTrace)

			addErr(PageError{
				Type:       TypeErrorJS,
				Message:    msg,
				StackTrace: stack,
				URL:        ev.ExceptionDetails.URL,
				Line:       int(ev.ExceptionDetails.LineNumber),
			})

		case *runtime.EventConsoleAPICalled:
			var msgParts []string
			for _, arg := range ev.Args {
				val := ""
				if arg.Value != nil {
					s := string(arg.Value)
					if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
						val = strings.Trim(s, "\"")
					} else {
						val = fmt.Sprint(arg.Value)
					}
				} else if arg.Description != "" {
					val = arg.Description
				}
				if val != "" {
					msgParts = append(msgParts, val)
				}
			}
			msg := strings.Join(msgParts, " ")

			if verbose {
				fmt.Printf("\n[DEBUG] Console %s: %s\n", ev.Type, msg)
			}

			if ev.Type == runtime.APITypeError {
				addErr(PageError{
					Type:    TypeErrorConsole,
					Message: msg,
				})
			}

		case *network.EventResponseReceived:
			if ev.Response.Status >= 400 {
				addErr(PageError{
					Type:    TypeErrorNetwork,
					Message: fmt.Sprintf("%d - %s", ev.Response.Status, ev.Response.URL),
					URL:     ev.Response.URL,
				})
			}

		case *network.EventLoadingFailed:
			if ev.Canceled {
				return
			}
			addErr(PageError{
				Type:    TypeErrorNetwork,
				Message: fmt.Sprintf("%s (%s)", ev.ErrorText, ev.Type),
			})
		}
	})

	// Execute scan: Enable domains, Navigate, Wait, and Settle
	err := chromedp.Run(ctx,
		network.Enable(),
		runtime.Enable(),
		chromedp.ActionFunc(func(ctx context.Context) error {
			if _, err := debugger.Enable().Do(ctx); err != nil {
				return err
			}
			return debugger.SetAsyncCallStackDepth(32).Do(ctx)
		}),
		page.Enable(),
		chromedp.Navigate(url),
		// Wait for the page to be ready and let the CPU settle briefly to catch post-load JS
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.Sleep(1*time.Second),
	)

	if err != nil {
		return nil, err
	}

	mu.Lock()
	defer mu.Unlock()
	res := make([]PageError, len(errors))
	copy(res, errors)
	return res, nil
}
// ExtractLinks finds all <a> tags and returns their href attributes.
// It uses a fallback strategy: if the body doesn't appear in time, it still tries to scrape what's there.
func (e *Engine) ExtractLinks(url string, timeout time.Duration, verbose bool) ([]string, error) {
	ctx, cancel := chromedp.NewContext(e.AllocCtx)
	defer cancel()

	// Use user-provided timeout with a 45s minimum for the scraper
	scraperTimeout := timeout
	if scraperTimeout < 45*time.Second {
		scraperTimeout = 45 * time.Second
	}

	ctx, cancel = context.WithTimeout(ctx, scraperTimeout)
	defer cancel()

	if verbose {
		fmt.Printf("\n[DEBUG] Scout is waiting for body to appear on %s...\n", url)
	}

	var hrefs []string
	// Step 1: Navigate and wait for body
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQuery),
	)

	// Step 2: Attempt extraction regardless of Step 1 success (Fallback)
	errEval := chromedp.Run(ctx,
		chromedp.Evaluate(`Array.from(document.querySelectorAll('a')).map(a => a.href)`, &hrefs),
	)

	if errEval != nil {
		if err != nil {
			return nil, fmt.Errorf("link extraction failed: target page took too long to respond and fallback failed: %w", err)
		}
		return nil, fmt.Errorf("link extraction fallback failed: %w", errEval)
	}

	if err != nil && verbose {
		fmt.Printf("\n[DEBUG] Scout: Body wait timed out, but extracted %d links from partial DOM\n", len(hrefs))
	}

	return hrefs, nil
}
