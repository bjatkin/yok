package cmd

import (
	"fmt"
	"os"

	"github.com/bjatkin/yok/codegen/gensh"
	"github.com/bjatkin/yok/compiler"
	"github.com/bjatkin/yok/parser"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "transpile yok code into sh code",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		srcFile := args[0]
		destFile := args[1]

		source, err := os.ReadFile(srcFile)
		if err != nil {
			return err
		}

		p := parser.New(source)
		script, err := p.Parse()
		if err != nil {
			for _, e := range p.Errors {
				fmt.Println(e)
			}
			return err
		}

		c := compiler.New(source)
		shAst, err := c.Compile(script)
		if err != nil {
			return err
		}

		code := gensh.Generate(shAst)
		err = os.WriteFile(destFile, []byte(code), 0o0755)
		if err != nil {
			return err
		}

		return nil
	},
}
