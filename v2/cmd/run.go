package cmd

import (
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "compilte your yok code to sh and run the sh script",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcFile := args[0]

		yokCode, err := os.ReadFile(srcFile)
		if err != nil {
			return err
		}

		shCode, err := complieYok(yokCode)
		if err != nil {
			return err
		}

		shFileName, err := writeTempScript(srcFile, shCode)
		if err != nil {
			return err
		}

		shCmd := exec.CommandContext(cmd.Context(), shFileName, args[1:]...)
		shCmd.Stdout = os.Stdout
		shCmd.Stderr = os.Stderr
		shCmd.Stdin = os.Stdin
		err = shCmd.Run()
		if err != nil {
			return err
		}

		// try to remove the temp file
		_ = os.Remove(shFileName)

		return nil
	},
}

func writeTempScript(yokFileName string, shCode []byte) (string, error) {
	yokBase := path.Base(yokFileName)
	yokPrefix := strings.TrimSuffix(yokBase, ".yok")
	tmpFileName := "*_" + yokPrefix + ".sh"
	systemTempDir := os.TempDir()
	file, err := os.CreateTemp(systemTempDir, tmpFileName)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = file.Write(shCode)
	if err != nil {
		return "", err
	}

	// make sure we can execute the script
	err = file.Chmod(0o0755)
	if err != nil {
		return "", err
	}

	return file.Name(), nil
}
