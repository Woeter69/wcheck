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
			for _, err := range res.Errors {
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

	pterm.Println()
	pterm.Info.Printfln("Scan completed in %s", duration.Round(time.Millisecond))
}
