package main

import (
	"os"
	"wcheck/cmd"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "wcheck",
	Short:   "wcheck is a headless browser scanner for localhost websites",
	Version: "1.0.1",
}

func main() {
	cmd.RegisterCommands(rootCmd)
	rootCmd.Flags().BoolP("version", "V", false, "Print the version number of wcheck")
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
