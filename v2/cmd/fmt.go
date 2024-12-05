package cmd

import (
	"github.com/spf13/cobra"

	"github.com/bjatkin/yok/errors"
)

func init() {
	rootCmd.AddCommand(fmtCmd)
}

var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "auto-format your yok code",
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("TODO")
	},
}
