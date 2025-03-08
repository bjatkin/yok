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

		yokCode, err := os.ReadFile(srcFile)
		if err != nil {
			return err
		}

		shCode, err := complieYok(yokCode)
		if err != nil {
			return err
		}

		err = os.WriteFile(destFile, shCode, 0o0755)
		if err != nil {
			return err
		}

		return nil
	},
}

func complieYok(yokCode []byte) ([]byte, error) {
	p := parser.New(yokCode)
	script, err := p.Parse()
	if err != nil {
		for _, e := range p.Errors {
			fmt.Println(e)
		}
		return nil, err
	}

	c := compiler.New(yokCode)
	shAst, err := c.Compile(script)
	if err != nil {
		return nil, err
	}

	shCode := gensh.Generate(shAst)
	return []byte(shCode), nil
}
