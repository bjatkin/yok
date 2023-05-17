package cmd

import (
	"fmt"
	"os"

	"github.com/bjatkin/yok/ast"
	"github.com/bjatkin/yok/parse"
	"github.com/bjatkin/yok/sym"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(fmtCmd)
}

var fmtCmd = &cobra.Command{
	Use:   "fmt [source file] [destination file]",
	Short: "format and re-write the source file",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		src := args[0]
		dest := args[1]
		return fmtSource(src, dest)
	},
}

func fmtSource(src, dest string) error {
	code, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to open src file %w", err)
	}

	// code must always be newline terminated in order to be lexed correctly
	code = append(code, '\n')

	table := sym.NewTable()
	client := parse.NewClient(table)
	tokens, err := client.Lex(src, code)
	if err != nil {
		return fmt.Errorf("failed to tokenize src %w", err)
	}

	parseTree, err := client.Parse(tokens)
	if err != nil {
		return fmt.Errorf("failed to parse src %w", err)
	}

	astClient := ast.NewClient(table)
	ast := astClient.Build(parseTree)

	rawYok := astClient.Yok(ast)

	err = os.WriteFile(dest, rawYok, 0o0665)
	if err != nil {
		return fmt.Errorf("failed to write dest file %w", err)
	}

	return nil
}
