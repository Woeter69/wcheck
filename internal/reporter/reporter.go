package reporter

import (
	"strings"
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
func PrintScanResults(url string, errors []string) {
	if len(errors) == 0 {
		pterm.Success.Printfln("No errors found on %s", url)
	} else {
		pterm.Warning.Printfln("Found %d errors on %s", len(errors), url)
	}
}

// PrintSummary displays a table of all results and the total time
func PrintSummary(results []engine.ScanResult, duration time.Duration) {
	pterm.Println()
	pterm.DefaultHeader.WithFullWidth().Println("Scan Summary")

	tableData := pterm.TableData{
		{"URL", "Status", "Errors"},
	}

	for _, res := range results {
		status := "OK"
		if len(res.Errors) > 0 {
			status = "FAIL"
		}
		
		errDetail := strings.Join(res.Errors, ", ")
		if len(errDetail) > 50 {
			errDetail = errDetail[:47] + "..."
		}
		
		tableData = append(tableData, []string{res.URL, status, errDetail})
	}

	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()

	pterm.Println()
	pterm.Info.Printfln("Scan completed in %s", duration.Round(time.Millisecond))
}
