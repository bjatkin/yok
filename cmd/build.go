package cmd

import (
	"fmt"
	"os"

	"github.com/bjatkin/yok/ast"
	"github.com/bjatkin/yok/bash"
	"github.com/bjatkin/yok/parse"
	"github.com/bjatkin/yok/sym"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
	Use:   "build [source file] [destination file]",
	Short: "transpile blowK source code into bash",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		src := args[0]
		dest := args[1]
		return buildSource(src, dest)
	},
}

func buildSource(src, dest string) error {
	code, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to open src file %w", err)
	}

	// code must always be newline terminated in order to be lexed correctly
	if code[len(code)-1] != '\n' {
		code = append(code, '\n')
	}

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
	yokAST := astClient.Build(parseTree)

	// TODO: how to resolve this double wrap
	yokAST = yokAST.Stmts[0].(ast.Root)
	// TODO: how to resolve this douple wrap

	bashClient := bash.NewClient(table)
	bashAST := bashClient.Build(yokAST)

	rawBash, err := bashClient.Bash(bashAST)
	if err != nil {
		return fmt.Errorf("failed to create raw bash %w", err)
	}

	err = os.WriteFile(dest, rawBash, 0o0775)
	if err != nil {
		return fmt.Errorf("failed to write dest file %w", err)
	}

	return nil
}
