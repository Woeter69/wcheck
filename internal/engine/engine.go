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
	TypeErrorJS          ErrorType = "JS Exception"
	TypeErrorConsole     ErrorType = "Console Error"
	TypeErrorNetwork     ErrorType = "Network Error"
	TypeErrorInteraction ErrorType = "Broken Interaction"
)

// PageError holds detailed information about a single error found on a page
type PageError struct {
	Type        ErrorType
	Message     string
	StackTrace  string
	URL         string
	Line        int
	ElementInfo string
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
		chromedp.Flag("max-connection-per-browser", "2"),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
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

// InteractWithPage finds and clicks all clickable elements to catch interaction errors.
// It filters out sensitive terms (Delete, Logout, etc.) and refreshes the page after each click.
func (e *Engine) InteractWithPage(ctx context.Context, url string, maxClicks int, verbose bool, setCurrentElement func(string)) error {
	type candidate struct {
		XPath string `json:"xpath"`
		Text  string `json:"text"`
		ID    string `json:"id"`
	}
	var candidates []candidate

	// Step 1: Find all potential clickable elements using a tighter JS selector
	err := chromedp.Run(ctx,
		chromedp.Evaluate(`(function() {
			const elements = Array.from(document.querySelectorAll('button, a, .btn, input[type="submit"], input[type="button"], [role="button"]'));
			const seenText = new Set();
			const result = [];
			
			function getXPath(element) {
				if (element.id !== '') return 'id("' + element.id + '")';
				if (element === document.body) return 'BODY';
				var ix = 0;
				var siblings = element.parentNode.childNodes;
				for (var i = 0; i < siblings.length; i++) {
					var sibling = siblings[i];
					if (sibling === element) return getXPath(element.parentNode) + '/' + element.tagName + '[' + (ix + 1) + ']';
					if (sibling.nodeType === 1 && sibling.tagName === element.tagName) ix++;
				}
			}

			for (const el of elements) {
				const style = window.getComputedStyle(el);
				if (style.display === 'none' || style.visibility === 'hidden' || style.opacity === '0') continue;
				if (style.cursor !== 'pointer') continue;

				const text = (el.innerText || el.value || "").trim();
				// Deduplicate by text to avoid clicking mobile/desktop duplicates
				if (text && seenText.has(text)) continue;
				if (text) seenText.add(text);

				result.push({
					xpath: getXPath(el),
					text: text.substring(0, 30),
					id: el.id || ""
				});
			}
			return result;
		})()`, &candidates),
	)
	if err != nil {
		return fmt.Errorf("failed to find clickable elements: %w", err)
	}

	if verbose {
		fmt.Printf("\n[DEBUG] Smarter Monkey: Found %d unique interactive elements on %s\n", len(candidates), url)
	}

	safetyTerms := []string{"logout", "delete", "remove", "sign out", "signout", "clear", "destroy"}
	
	count := 0
	for _, cand := range candidates {
		if count >= maxClicks {
			break
		}

		// Safety Filter
		skip := false
		lowerText := strings.ToLower(cand.Text)
		for _, term := range safetyTerms {
			if strings.Contains(lowerText, term) {
				skip = true
				break
			}
		}

		if skip {
			continue
		}

		info := cand.Text
		if info == "" {
			info = cand.ID
		}
		if info == "" {
			info = cand.XPath
		}
		
		setCurrentElement(info)
		count++

		if verbose {
			fmt.Printf("[DEBUG] Monkey: Testing [%s] (%d/%d)\n", info, count, len(candidates))
		}

		// Use a tighter 3s timeout for each click
		clickCtx, clickCancel := context.WithTimeout(ctx, 3*time.Second)
		
		// Perform Scroll and Click
		_ = chromedp.Run(clickCtx,
			chromedp.ScrollIntoView(cand.XPath, chromedp.BySearch),
			chromedp.Click(cand.XPath, chromedp.BySearch),
			chromedp.Sleep(300*time.Millisecond),
			// Reset quickly
			chromedp.Navigate(url),
			chromedp.WaitReady("body", chromedp.ByQuery),
		)
		
		clickCancel()
		setCurrentElement("")
	}

	return nil
}

// ScanPage navigates to a URL and monitors for errors
func (e *Engine) ScanPage(url string, verbose bool, timeout time.Duration, interact bool, maxClicks int) ([]PageError, error) {
	// Create a new tab/context for this specific scan
	ctx, cancel := chromedp.NewContext(e.AllocCtx)
	defer cancel()

	// Use the provided timeout for the entire page lifecycle
	ctx, cancel = context.WithTimeout(ctx, timeout)
	defer cancel()

	var (
		mu                 sync.Mutex
		errors             []PageError
		isInteractionPhase bool
		currentElement     string
	)

	addErr := func(err PageError) {
		mu.Lock()
		if isInteractionPhase && (err.Type == TypeErrorJS || err.Type == TypeErrorConsole) {
			err.Type = TypeErrorInteraction
			err.ElementInfo = currentElement
			err.Message = fmt.Sprintf("Interaction with [%s] triggered: %s", currentElement, err.Message)
		}
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
				msg := fmt.Sprintf("%d - %s", ev.Response.Status, ev.Response.URL)
				if ev.Response.Status == 429 {
					msg = fmt.Sprintf("⚠️ Rate Limited (429): %s", ev.Response.URL)
				}
				addErr(PageError{
					Type:    TypeErrorNetwork,
					Message: msg,
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

	// Step 1: Execute scan: Enable domains, Navigate, Wait, and Settle
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

	// Step 2: Perform Interaction Testing (Optional)
	if interact {
		if verbose {
			fmt.Printf("\n[DEBUG] Starting interaction testing for %s...\n", url)
		}
		
		mu.Lock()
		isInteractionPhase = true
		mu.Unlock()
		
		_ = e.InteractWithPage(ctx, url, maxClicks, verbose, func(info string) {
			mu.Lock()
			currentElement = info
			mu.Unlock()
		})
		
		mu.Lock()
		isInteractionPhase = false
		mu.Unlock()
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
	if verbose {
		fmt.Printf("\n[DEBUG] Scout starting (Interaction disabled for this phase) on %s\n", url)
	}
	ctx, cancel := chromedp.NewContext(e.AllocCtx)
	defer cancel()

	// Use user-provided timeout with a 45s minimum for the scraper
	scraperTimeout := timeout
	if scraperTimeout < 45*time.Second {
		scraperTimeout = 45 * time.Second
	}

	ctx, cancel = context.WithTimeout(ctx, scraperTimeout)
	defer cancel()

	var hrefs []string
	// Step 1: Navigate and wait for DOM ready (much faster than WaitVisible)
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),
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
