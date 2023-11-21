package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "yok <command> [arguments]",
	Short:         "yok is the tool for managing yok source code",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// TODO: can we/ should we tie this into the error itself?
		fmt.Println(err)
		os.Exit(1)
	}
}
