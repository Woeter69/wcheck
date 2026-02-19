package cmd

import (
	"context"
	"fmt"
	"os"
	"time"
	"wcheck/internal/engine"
	"wcheck/internal/reporter"

	"github.com/spf13/cobra"
)

var (
	interactTimeout   int
	interactMaxClicks int
	interactVerbose   bool
	interactDelay     int
)

var interactCmd = &cobra.Command{
	Use:   "interact [url]",
	Short: "Interactively test a single page by clicking elements",
	Long: `Interactively test a single page by finding and clicking on all interactive elements
(buttons, links, etc.) to detect JS errors triggered by user interaction.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()
		targetURL := args[0]

		reporter.PrintDebug(interactVerbose, "Initializing Interaction Monkey for %s", targetURL)

		eng := engine.NewEngine()
		defer eng.Close()

		if interactDelay > 0 {
			time.Sleep(time.Duration(interactDelay) * time.Second)
		}

		reporter.PrintDebug(interactVerbose, "Scanning %s with interaction enabled (max %d clicks)...", targetURL, interactMaxClicks)
		
		errors, err := eng.ScanPage(targetURL, interactVerbose, time.Duration(interactTimeout)*time.Second, true, interactMaxClicks)
		
		// Retry logic for interact
		if err != nil || hasRateLimit(errors) {
			if err == context.DeadlineExceeded || hasRateLimit(errors) {
				retryDelay := 5
				reporter.PrintDebug(interactVerbose, "Retrying %s after %ds sleep...", targetURL, retryDelay)
				time.Sleep(time.Duration(retryDelay) * time.Second)
				errors, err = eng.ScanPage(targetURL, interactVerbose, time.Duration(interactTimeout)*time.Second, true, interactMaxClicks)
			}
		}

		if err != nil {
			msg := err.Error()
			if err == context.DeadlineExceeded {
				msg = fmt.Sprintf("⌛ Page load timed out after %ds", interactTimeout)
			}
			reporter.PrintError("Interaction test failed for %s: %s", targetURL, msg)
			os.Exit(1)
		}

		res := engine.ScanResult{
			URL:    targetURL,
			Errors: errors,
		}

		reporter.PrintScanResults(res.URL, res.Errors)
		reporter.PrintSummary([]engine.ScanResult{res}, time.Since(start), interactVerbose)

		if len(errors) > 0 {
			os.Exit(1)
		}
	},
}

// RegisterInteract adds interactCmd to the root command
func RegisterInteract(rootCmd *cobra.Command) {
	rootCmd.AddCommand(interactCmd)
	interactCmd.Flags().IntVarP(&interactTimeout, "timeout", "t", 60, "Timeout in seconds for the entire interaction session")
	interactCmd.Flags().IntVarP(&interactMaxClicks, "max-clicks", "m", 20, "Maximum number of elements to click")
	interactCmd.Flags().BoolVarP(&interactVerbose, "verbose", "v", false, "Enable verbose logging")
	interactCmd.Flags().IntVarP(&interactDelay, "delay", "d", 0, "Delay in seconds before scan")
}
