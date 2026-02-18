package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
	"wcheck/internal/crawler"
	"wcheck/internal/engine"
	"wcheck/internal/reporter"

	"github.com/spf13/cobra"
)

var (
	workers int
	verbose bool
	timeout int
)

var scanCmd = &cobra.Command{
	Use:   "scan [url]",
	Short: "Scan a URL with multiple workers",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()
		targetURL := args[0]

		reporter.PrintDebug(verbose, "Initializing engine for %s with %d workers", targetURL, workers)

		eng := engine.NewEngine()
		defer eng.Close()

		c := crawler.NewCrawler(targetURL, eng)
		links, err := c.GetLinks(time.Duration(timeout)*time.Second, verbose)
		if err != nil {
			reporter.PrintError("Failed to get links: %v", err)
			os.Exit(1)
		}

		reporter.PrintDebug(verbose, "Discovered %d internal links...", len(links))

		jobs := make(chan string, len(links))
		resultsChan := make(chan engine.ScanResult, len(links))
		var wg sync.WaitGroup

		// Spawn workers
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for link := range jobs {
					reporter.PrintDebug(verbose, "[Worker %d] Scanning %s...", workerID, link)
					errors, err := eng.ScanPage(link, verbose, time.Duration(timeout)*time.Second)
					if err != nil {
						msg := err.Error()
						if err == context.DeadlineExceeded {
							msg = fmt.Sprintf("⌛ Page load timed out after %ds", timeout)
						}
						reporter.PrintError("[Worker %d] Scan failed for %s: %s", workerID, link, msg)
						resultsChan <- engine.ScanResult{URL: link, Errors: []engine.PageError{{
							Type:    engine.TypeErrorNetwork,
							Message: msg,
						}}}
						continue
					}
					resultsChan <- engine.ScanResult{URL: link, Errors: errors}
				}
			}(i)
		}

		// Feed jobs
		for _, link := range links {
			jobs <- link
		}
		close(jobs)

		// Wait for workers in a separate goroutine to close results
		go func() {
			wg.Wait()
			close(resultsChan)
		}()

		var allResults []engine.ScanResult
		hasErrors := false

		// Process results as they come in
		for res := range resultsChan {
			reporter.PrintScanResults(res.URL, res.Errors)
			allResults = append(allResults, res)
			if len(res.Errors) > 0 {
				hasErrors = true
			}
		}

		reporter.PrintSummary(allResults, time.Since(start), verbose)

		if hasErrors {
			os.Exit(1)
		}
	},
}

// RegisterCommands adds scanCmd to the root command
func RegisterCommands(rootCmd *cobra.Command) {
	rootCmd.AddCommand(scanCmd)
	scanCmd.Flags().IntVarP(&workers, "workers", "w", 5, "Number of parallel workers")
	scanCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	scanCmd.Flags().IntVarP(&timeout, "timeout", "t", 30, "Timeout in seconds for each page scan")
}
