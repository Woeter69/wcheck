package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"wcheck/internal/engine"
)

func TestInteractionsFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var sb strings.Builder
		sb.WriteString("<html><body>")
		for i := 1; i <= 10; i++ {
			sb.WriteString(fmt.Sprintf("<button id='btn%d'>Button %d</button>", i, i))
			sb.WriteString(fmt.Sprintf("<a href='#'>Link %d</a>", i))
		}
		sb.WriteString("</body></html>")
		fmt.Fprintln(w, sb.String())
	}))
	t.Cleanup(ts.Close)

	eng := engine.NewEngine()
	t.Cleanup(eng.Close)

	// Custom scanner that just returns the candidates found
	ctx, cancel := chromedp.NewContext(eng.AllocCtx)
	defer cancel()

	var candidatesFound int
	err := chromedp.Run(ctx,
		chromedp.Navigate(ts.URL),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Evaluate(`(function() {
			const selector = 'a, button, [role="button"], .btn, .button, input[type="submit"]';
			return document.querySelectorAll(selector).length;
		})()`, &candidatesFound),
	)

	if err != nil {
		t.Fatalf("Failed to run discovery JS: %v", err)
	}

	if candidatesFound < 20 {
		t.Errorf("Expected at least 20 elements, found %d", candidatesFound)
	}
}

func TestTimeoutHandling(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `
			<html>
				<body>
					<button id="errBtn" onclick="setTimeout(() => console.error('Delayed Error'), 100)">Trigger Error</button>
				</body>
			</html>
		`)
	}))
	t.Cleanup(ts.Close)

	eng := engine.NewEngine()
	t.Cleanup(eng.Close)

	// We expect the error to be caught as a Broken Interaction
	errors, err := eng.ScanPage(ts.URL, false, 20*time.Second, true, 1)
	if err != nil {
		t.Fatalf("ScanPage failed unexpectedly: %v", err)
	}

	foundErr := false
	for _, e := range errors {
		t.Logf("Found error: Type=%s, Message=%s", e.Type, e.Message)
		if e.Type == engine.TypeErrorInteraction && strings.Contains(e.Message, "Delayed Error") {
			foundErr = true
			break
		}
	}

	if !foundErr {
		t.Errorf("Expected interaction error was not reported. Total errors: %d", len(errors))
	}
}
