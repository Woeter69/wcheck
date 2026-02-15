package main

import (
	"os"
	"wcheck/cmd"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wcheck",
	Short: "wcheck is a headless browser scanner for localhost websites",
}

func main() {
	cmd.RegisterCommands(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
