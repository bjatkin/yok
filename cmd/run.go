package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run [source file] [destination file]",
	Short: "run first complies the source file to the target and then runs it",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		src := args[0]
		dest := args[1]
		err := buildSource(src, dest)
		if err != nil {
			return fmt.Errorf("failed to build source: %w", err)
		}

		run := exec.Command("./" + dest)
		fmt.Println("run: ", run.String())
		out, err := run.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to run complied file: %w", err)
		}

		fmt.Println(string(out))
		return nil
	},
}
