package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bjatkin/yok/codegen/genyok"
	"github.com/bjatkin/yok/parser"
)

func init() {
	rootCmd.AddCommand(fmtCmd)
}

var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "auto-format your yok code",
	RunE: func(cmd *cobra.Command, args []string) error {
		srcFile := args[0]

		yokCode, err := os.ReadFile(srcFile)
		if err != nil {
			return err
		}

		p := parser.New(yokCode)
		script, err := p.Parse()
		if err != nil {
			for _, e := range p.Errors {
				fmt.Println(e)
			}
			return err
		}

		formatedCode := genyok.Generate(script, yokCode)
		err = os.WriteFile(srcFile, []byte(formatedCode), 0o0655)
		if err != nil {
			return err
		}

		return nil
	},
}
