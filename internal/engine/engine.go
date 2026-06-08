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
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("max-connections-per-browser", "10"),
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
func (e *Engine) InteractWithPage(ctx context.Context, url string, maxClicks int, verbose bool, addErr func(PageError), setCurrentElement func(string)) error {
	type candidate struct {
		XPath string  `json:"xpath"`
		Text  string  `json:"text"`
		ID    string  `json:"id"`
		X     float64 `json:"x"`
		Y     float64 `json:"y"`
	}
	var candidates []candidate

	// Step 1: Wait for hydration and stability
	err := chromedp.Run(ctx,
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(2*time.Second), // Settle time for React/Next.js
	)
	if err != nil {
		return fmt.Errorf("page body not stable: %w", err)
	}

	// Step 2: Find all potential clickable elements using a broader JS selector
	err = chromedp.Run(ctx,
		chromedp.Evaluate(`(function() {
			const selector = 'a, button, [role="button"], .btn, .button, input[type="submit"]';
			const elements = Array.from(document.querySelectorAll(selector));
			const seenText = new Set();
			const result = [];
			
			function getXPath(element) {
				if (element.id !== '') return '//*[@id="' + element.id + '"]';
				if (element === document.body) return '/html/body';
				var ix = 0;
				var siblings = element.parentNode.childNodes;
				for (var i = 0; i < siblings.length; i++) {
					var sibling = siblings[i];
					if (sibling === element) return getXPath(element.parentNode) + '/' + element.tagName.toLowerCase() + '[' + (ix + 1) + ']';
					if (sibling.nodeType === 1 && sibling.tagName === element.tagName) ix++;
				}
			}

			function isVisible(el) {
				const rect = el.getBoundingClientRect();
				if (rect.width === 0 || rect.height === 0) return false;
				
				const style = window.getComputedStyle(el);
				if (style.display === 'none' || style.visibility === 'hidden' || style.opacity === '0') return false;
				
				// Check if element or its parent is hidden
				if (el.offsetParent === null && style.position !== 'fixed') return false;

				return true;
			}

			for (const el of elements) {
				if (!isVisible(el)) continue;
				const text = (el.innerText || el.value || "").trim();
				if (text && seenText.has(text)) continue;
				if (text) seenText.add(text);

				const rect = el.getBoundingClientRect();
				result.push({
					xpath: getXPath(el),
					text: text.substring(0, 30),
					id: el.id || "",
					x: rect.left + rect.width / 2,
					y: rect.top + rect.height / 2
				});
			}
			return result;
		})()`, &candidates),
	)
	if err != nil {
		return fmt.Errorf("failed to find clickable elements: %w", err)
	}

	if verbose {
		fmt.Printf("\n[DEBUG] Smarter Monkey: Found %d unique visible interactive elements on %s\n", len(candidates), url)
	}

	if len(candidates) == 0 {
		return fmt.Errorf("no visible interactive elements found")
	}

	safetyTerms := []string{"logout", "delete", "remove", "sign out", "signout", "clear", "destroy"}
	count := 0
	var interactionErrors int

	for _, cand := range candidates {
		if count >= maxClicks {
			break
		}

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

		// Reset errors for this specific interaction to avoid cross-contamination
		// Actually, we don't clear all errors, but we mark the start of THIS interaction
		
		// Click with strict sub-context
		clickCtx, clickCancel := context.WithTimeout(ctx, 10*time.Second)
		err = chromedp.Run(clickCtx,
			chromedp.ScrollIntoView(cand.XPath, chromedp.BySearch),
			chromedp.Sleep(200*time.Millisecond),
			chromedp.MouseClickXY(cand.X, cand.Y),
			chromedp.Sleep(1000*time.Millisecond), // Wait for async effects
		)

		if err != nil {
			interactionErrors++
			addErr(PageError{
				Type:        TypeErrorInteraction,
				Message:     fmt.Sprintf("Interaction failed on [%s]: %v", info, err),
				ElementInfo: info,
			})
			if verbose {
				fmt.Printf("[DEBUG] [FAIL] Interaction failed on [%s]: %v\n", info, err)
			}
		} else if verbose {
			fmt.Printf("[DEBUG] [OK] Successfully clicked [%s]\n", info)
		}

		// Reset page state - navigate back and wait for it to settle
		// We don't want "Immediate Console Error" from the reload to be attributed to the PREVIOUS click
		setCurrentElement("")
		_ = chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.WaitReady("body", chromedp.ByQuery),
			chromedp.Sleep(500*time.Millisecond),
		)
		clickCancel()
	}

	if interactionErrors > 0 {
		return fmt.Errorf("%d interactions failed (timeouts or errors)", interactionErrors)
	}
	return nil
}

// ScanPage navigates to a URL and monitors for errors
func (e *Engine) ScanPage(url string, verbose bool, timeout time.Duration, interact bool, maxClicks int) ([]PageError, error) {
	// Create a new tab/context for this specific scan
	ctx, cancelTab := chromedp.NewContext(e.AllocCtx)
	defer cancelTab()

	// Use the provided timeout for the entire page lifecycle
	ctx, cancelTimeout := context.WithTimeout(ctx, timeout)
	defer cancelTimeout()

	var (
		mu                 sync.Mutex
		errors             []PageError
		isInteractionPhase bool
		currentElement     string
	)

	addErr := func(err PageError) {
		mu.Lock()
		defer mu.Unlock()

		// Deduplicate: Don't add exactly the same error twice
		for _, existing := range errors {
			if existing.Type == err.Type && existing.Message == err.Message && existing.URL == err.URL && existing.Line == err.Line {
				return
			}
		}

		if isInteractionPhase && currentElement != "" {
			if err.Type == TypeErrorJS || err.Type == TypeErrorConsole || err.Type == TypeErrorNetwork {
				err.Type = TypeErrorInteraction
				err.ElementInfo = currentElement
				err.Message = fmt.Sprintf("Interaction with [%s] triggered: %s", currentElement, err.Message)
			} else if err.Type == TypeErrorInteraction && err.ElementInfo == "" {
				err.ElementInfo = currentElement
			}
		}

		errors = append(errors, err)
		if verbose && err.Type != TypeErrorInteraction {
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
			if ev.Type == runtime.APITypeError {
				addErr(PageError{
					Type:    TypeErrorConsole,
					Message: msg,
				})
			}

		case *network.EventResponseReceived:
			if ev.Response.Status >= 400 {
				msg := fmt.Sprintf("HTTP %d - %s", ev.Response.Status, ev.Response.URL)
				addErr(PageError{
					Type:    TypeErrorNetwork,
					Message: msg,
					URL:     ev.Response.URL,
				})
			}

		case *network.EventLoadingFailed:
			if !ev.Canceled {
				addErr(PageError{
					Type:    TypeErrorNetwork,
					Message: fmt.Sprintf("Network Error: %s (%s)", ev.ErrorText, ev.Type),
				})
			}
		}
	})

	// Step 1: Execute scan: Enable domains, Navigate with Backoff
	err := chromedp.Run(ctx,
		network.Enable(),
		runtime.Enable(),
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := debugger.Enable().Do(ctx)
			if err != nil {
				return err
			}
			return debugger.SetAsyncCallStackDepth(32).Do(ctx)
		}),
		page.Enable(),
	)
	if err != nil {
		return nil, err
	}

	// Exponential Backoff for initial navigation
	maxRetries := 3
	var navErr error
	for i := 0; i < maxRetries; i++ {
		navErr = chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.WaitReady("body", chromedp.ByQuery),
		)
		if navErr == nil {
			break
		}
		if i < maxRetries-1 {
			time.Sleep(time.Duration(1<<uint(i)) * time.Second)
		}
	}
	if navErr != nil {
		return nil, navErr
	}

	// Step 2: Perform Interaction Testing (Optional)
	if interact {
		mu.Lock()
		isInteractionPhase = true
		mu.Unlock()
		
		err := e.InteractWithPage(ctx, url, maxClicks, verbose, addErr, func(info string) {
			mu.Lock()
			currentElement = info
			mu.Unlock()
		})
		if err != nil {
			addErr(PageError{
				Type:    TypeErrorInteraction,
				Message: err.Error(),
			})
		}
		
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
	ctx, cancelTab := chromedp.NewContext(e.AllocCtx)
	defer cancelTab()

	// Use user-provided timeout for the scraper
	ctx, cancelTimeout := context.WithTimeout(ctx, timeout)
	defer cancelTimeout()

	var hrefs []string
	// Step 1: Navigate with Backoff and wait for DOM ready (much faster than WaitVisible)
	maxRetries := 3
	var navErr error
	for i := 0; i < maxRetries; i++ {
		navErr = chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.WaitReady("body", chromedp.ByQuery),
		)
		if navErr == nil {
			break
		}
		if i < maxRetries-1 {
			// Check if context is already done before sleeping
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("link extraction timed out during navigation: %w", ctx.Err())
			default:
			}
			backoff := time.Duration(1<<uint(i)) * time.Second
			if verbose {
				fmt.Printf("[DEBUG] Scout navigation failed, retrying in %v: %v\n", backoff, navErr)
			}
			time.Sleep(backoff)
		}
	}

	// Step 2: Attempt extraction regardless of Step 1 success (Fallback)
	errEval := chromedp.Run(ctx,
		chromedp.Evaluate(`Array.from(document.querySelectorAll('a')).map(a => a.href)`, &hrefs),
	)

	if errEval != nil {
		if navErr != nil {
			return nil, fmt.Errorf("link extraction failed: target page took too long to respond and fallback failed: %w", navErr)
		}
		return nil, fmt.Errorf("link extraction fallback failed: %w", errEval)
	}

	if navErr != nil && verbose {
		fmt.Printf("\n[DEBUG] Scout: Body wait timed out, but extracted %d links from partial DOM\n", len(hrefs))
	}

	return hrefs, nil
}
