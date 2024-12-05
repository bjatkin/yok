package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "yok <command> [arguments]",
	Short:        "yok is the tool for managing yok source code",
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(1)
	}
}
