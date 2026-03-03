package engine

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestExtractLinks(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `
			<html>
				<body>
					<a href="/internal">Internal Link</a>
					<a href="https://external.com">External Link</a>
					<button>Not a link</button>
					<div>Just text</div>
				</body>
			</html>
		`)
	}))
	t.Cleanup(ts.Close)

	eng := NewEngine()
	t.Cleanup(eng.Close)

	links, err := eng.ExtractLinks(ts.URL, 10*time.Second, false)
	if err != nil {
		t.Fatalf("Failed to extract links: %v", err)
	}

	foundInternal := false
	foundExternal := false
	for _, link := range links {
		if strings.Contains(link, "/internal") {
			foundInternal = true
		}
		if strings.Contains(link, "external.com") {
			foundExternal = true
		}
	}

	if !foundInternal {
		t.Error("Did not find internal link")
	}
	if !foundExternal {
		t.Error("Did not find external link")
	}
}

func TestScanErrors(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/404" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		fmt.Fprintln(w, `
			<html>
				<body>
					<img src="/404">
					<script>
						// JS Syntax Error
						console.log("Hello"
					</script>
				</body>
			</html>
		`)
	}))
	t.Cleanup(ts.Close)

	eng := NewEngine()
	t.Cleanup(eng.Close)

	errors, err := eng.ScanPage(ts.URL, false, 10*time.Second, false, 0)
	if err != nil {
		t.Fatalf("ScanPage failed: %v", err)
	}

	var found404, foundJS bool
	for _, e := range errors {
		if e.Type == TypeErrorNetwork && strings.Contains(e.Message, "404") {
			found404 = true
		}
		if e.Type == TypeErrorJS {
			foundJS = true
		}
	}

	if !found404 {
		t.Error("Expected 404 network error not found")
	}
	if !foundJS {
		t.Error("Expected JS exception not found")
	}
}

func TestInteract(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `
			<html>
				<body>
					<button id="errBtn" onclick="console.error('Interaction Error Triggered')">Click Me</button>
				</body>
			</html>
		`)
	}))
	t.Cleanup(ts.Close)

	eng := NewEngine()
	t.Cleanup(eng.Close)

	// Enable interaction and limit to 1 click for speed
	errors, err := eng.ScanPage(ts.URL, false, 15*time.Second, true, 1)
	if err != nil {
		t.Fatalf("ScanPage with interaction failed: %v", err)
	}

	foundInteractionErr := false
	for _, e := range errors {
		if e.Type == TypeErrorInteraction && strings.Contains(e.Message, "Interaction Error Triggered") {
			foundInteractionErr = true
			break
		}
	}

	if !foundInteractionErr {
		t.Error("Expected interaction error (console.error) not found")
	}
}
