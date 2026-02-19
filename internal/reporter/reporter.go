package reporter

import (
	"fmt"
	"time"
	"wcheck/internal/engine"

	"github.com/pterm/pterm"
)

// PrintDebug prints a debug message if verbose mode is enabled
func PrintDebug(verbose bool, format string, a ...interface{}) {
	if verbose {
		pterm.Debug.Printfln(format, a...)
	}
}

// PrintError prints an error message
func PrintError(format string, a ...interface{}) {
	pterm.Error.Printfln(format, a...)
}

// PrintScanResults prints the results of a page scan immediately
func PrintScanResults(url string, errors []engine.PageError) {
	if len(errors) == 0 {
		pterm.Success.Printfln("No errors found on %s", url)
	} else {
		pterm.Warning.Printfln("Found %d errors on %s", len(errors), url)
	}
}

// PrintSummary displays a table of all results and the total time
func PrintSummary(results []engine.ScanResult, duration time.Duration, verbose bool) {
	pterm.Println()
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgBlue)).Println("SCAN SUMMARY")

	tableData := pterm.TableData{
		{"URL", "Status", "Errors"},
	}

	failedResults := []engine.ScanResult{}

	for _, res := range results {
		status := pterm.Green("OK")
		errDetail := "0 Errors"
		if len(res.Errors) > 0 {
			status = pterm.Red("FAIL")
			errDetail = fmt.Sprintf("%d Errors Found (See below)", len(res.Errors))
			failedResults = append(failedResults, res)
		}

		tableData = append(tableData, []string{res.URL, status, errDetail})
	}

	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()

	// Deep Dive Section
	if len(failedResults) > 0 {
		pterm.Println()
		pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgRed)).Println("DEEP DIVE: ERROR DETAILS")

		for _, res := range failedResults {
			pterm.DefaultSection.Printf("Errors for %s", res.URL)
			
			// Group interactions if any
			var interactionErrors []engine.PageError
			var otherErrors []engine.PageError
			
			for _, err := range res.Errors {
				if err.Type == engine.TypeErrorInteraction {
					interactionErrors = append(interactionErrors, err)
				} else {
					otherErrors = append(otherErrors, err)
				}
			}

			if len(interactionErrors) > 0 {
				pterm.DefaultHeader.WithFullWidth(false).Println("FAILED INTERACTIONS")
				for _, err := range interactionErrors {
					pterm.Printf("💥 %s [%s]\n", pterm.Magenta("Broken Interaction:"), pterm.Bold.Sprint(err.ElementInfo))
					pterm.Printf("   Reason: %s\n", pterm.Red(err.Message))
					if err.StackTrace != "" {
						pterm.Println(pterm.Gray(err.StackTrace))
					}
					pterm.Println()
				}
			}

			if len(otherErrors) > 0 {
				if len(interactionErrors) > 0 {
					pterm.DefaultHeader.WithFullWidth(false).Println("OTHER ERRORS")
				}
				for _, err := range otherErrors {
					var prefix string
					var msgStyle *pterm.Style

					switch err.Type {
					case engine.TypeErrorJS:
						prefix = "❌ JS Exception:"
						msgStyle = pterm.NewStyle(pterm.FgRed)
					case engine.TypeErrorConsole:
						prefix = "⚠️ Console Error:"
						msgStyle = pterm.NewStyle(pterm.FgYellow)
					case engine.TypeErrorNetwork:
						prefix = "🚫 Network Error:"
						msgStyle = pterm.NewStyle(pterm.FgCyan)
					default:
						prefix = "❓ Unknown Error:"
						msgStyle = pterm.NewStyle(pterm.FgGray)
					}

					pterm.Printf("%s %s\n", msgStyle.Sprint(prefix), err.Message)

					if err.StackTrace != "" {
						pterm.Println(pterm.Gray(err.StackTrace))
					} else if verbose && err.URL != "" {
						pterm.Printf("   at %s:%d\n", pterm.Gray(err.URL), err.Line)
					}
					pterm.Println()
				}
			}
		}
	}

	pterm.Println()
	pterm.Info.Printfln("Scan completed in %s", duration.Round(time.Millisecond))
}
