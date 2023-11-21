package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bjatkin/yok/v2/parser"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
	Use:   "build [source file] [destination file]",
	Short: "transpile yok source code into bash",
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

	parser := parser.New()
	program, err := parser.Parse(src, code)
	if err != nil {
		return err
	}

	// Just To Debug
	programJson, err := json.MarshalIndent(program, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(programJson))

	return nil
}

/*
func buildSource(src, dest string) error {
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
	yokAST := astClient.Build(parseTree)

	err = astClient.Validate(yokAST)
	if err != nil {
		return fmt.Errorf("validation failure: %w", err)
	}

	bashClient := bash.NewClient(table)
	bashAST := bashClient.Build(yokAST)

	rawBash := bashClient.Bash(bashAST)

	err = os.WriteFile(dest, rawBash, 0o0775)
	if err != nil {
		return fmt.Errorf("failed to write dest file %w", err)
	}

	return nil
}
*/
