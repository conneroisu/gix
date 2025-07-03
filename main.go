// Package main implements the gix command-line interface.
//
// gix is a pure Go implementation of the Nix expression language interpreter.
// It provides a complete lexer, parser, and evaluator for Nix expressions,
// supporting all major language features including:
//
//   - Arithmetic and logical expressions
//   - String and number literals
//   - Lists and attribute sets
//   - Function definitions and applications
//   - Let bindings and variable scoping
//   - Conditional expressions (if-then-else)
//   - Derivation creation and handling
//   - Built-in functions for type checking, list operations, etc.
//
// The CLI supports three modes of operation:
//   - Interactive REPL mode (-i flag)
//   - Expression evaluation mode (-e flag)
//   - File evaluation mode (positional argument)
//
// Examples:
//
//	gix -e "1 + 2"                    # Evaluate expression
//	gix -i                            # Start REPL
//	gix file.nix                      # Evaluate file
//	gix -e 'let x = 5; in x * 2'     # Complex expression
//	gix -e 'derivation { name = "hello"; builder = "/bin/sh"; }'  # Create derivation
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/conneroisu/gix/pkg/eval"
	"github.com/conneroisu/gix/pkg/lexer"
	"github.com/conneroisu/gix/pkg/parser"
)

// main is the entry point for the gix CLI.
//
// It parses command-line flags and dispatches to the appropriate mode:
//   - Help mode: shows usage information
//   - Expression mode: evaluates a single expression from the command line
//   - Interactive mode: starts a REPL session
//   - File mode: evaluates a Nix file
//   - Default: shows help if no arguments provided
func main() {
	var (
		interactive = flag.Bool("i", false, "Interactive REPL mode")
		expression  = flag.String("e", "", "Evaluate expression")
		help        = flag.Bool("h", false, "Show help")
	)
	flag.Parse()

	if *help {
		showHelp()

		return
	}

	if *expression != "" {
		// Evaluate expression from command line
		evalExpression(*expression, ".")
	} else if *interactive {
		// Start REPL
		startREPL()
	} else if flag.NArg() > 0 {
		// Evaluate file
		evalFile(flag.Arg(0))
	} else {
		showHelp()
	}
}

// showHelp displays usage information for the gix CLI.
//
// This includes all command-line options, usage patterns, and practical examples
// to help users understand how to use the interpreter effectively.
func showHelp() {
	fmt.Println("gix - A pure Go implementation of Nix")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  gix [options] [file]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -i          Interactive REPL mode")
	fmt.Println("  -e EXPR     Evaluate expression")
	fmt.Println("  -h          Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  gix -e '1 + 2'")
	fmt.Println("  gix -i")
	fmt.Println("  gix file.nix")
}

// evalExpression evaluates a single Nix expression and prints the result.
//
// This function implements the complete evaluation pipeline:
//  1. Lexical analysis to tokenize the input string
//  2. Syntactic analysis to build an Abstract Syntax Tree (AST)
//  3. Semantic evaluation to compute the final value
//  4. Pretty-printing of the result
//
// Parameters:
//   - expr: The Nix expression string to evaluate
//   - baseDir: The base directory for resolving relative paths
//
// If any step fails, the function prints an error message and exits with status 1.
func evalExpression(expr string, baseDir string) {
	// Tokenize the input expression
	l := lexer.New(expr)

	// Parse tokens into an Abstract Syntax Tree
	p := parser.New(l)
	ast, err := p.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	// Evaluate the AST to produce a value
	e := eval.New(baseDir)
	result, err := e.Eval(ast)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Evaluation error: %v\n", err)
		os.Exit(1)
	}

	// Display the result
	fmt.Println(result.String())
}

// evalFile reads and evaluates a Nix file from the filesystem.
//
// This function handles file I/O and delegates expression evaluation to evalExpression.
// The base directory for path resolution is set to the directory containing the file,
// allowing relative imports and paths to work correctly.
//
// Parameters:
//   - filename: Path to the Nix file to evaluate
//
// If the file cannot be read, the function prints an error and exits with status 1.
func evalFile(filename string) {
	// Read the entire file content
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Use the file's directory as the base for path resolution
	baseDir := filepath.Dir(filename)
	evalExpression(string(content), baseDir)
}

// startREPL starts an interactive Read-Eval-Print Loop for the Nix interpreter.
//
// The REPL provides an interactive environment where users can:
//   - Enter Nix expressions line by line
//   - See immediate evaluation results
//   - Use special commands (prefixed with ':')
//   - Maintain state across multiple evaluations
//
// The REPL continues until the user types ":quit", ":q", or sends EOF (Ctrl+D).
// Each expression is evaluated in the same environment, so variable bindings
// persist across lines.
//
// Special commands:
//   - :quit, :q  - Exit the REPL
//   - :help, :h  - Show available commands
func startREPL() {
	fmt.Println("gix repl - Type :quit to exit")
	fmt.Println()

	// Create a scanner for reading user input line by line
	scanner := bufio.NewScanner(os.Stdin)

	// Create a single evaluator instance to maintain state across evaluations
	e := eval.New(".")

	for {
		// Display the prompt and wait for input
		fmt.Print("nix-repl> ")
		if !scanner.Scan() {
			// EOF or error, break the loop
			break
		}

		// Get and clean up the input line
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Handle quit commands
		if line == ":quit" || line == ":q" {
			break
		}

		// Handle other REPL commands (prefixed with ':')
		if strings.HasPrefix(line, ":") {
			handleReplCommand(line)

			continue
		}

		// Parse and evaluate the Nix expression
		l := lexer.New(line)
		p := parser.New(l)
		ast, err := p.Parse()
		if err != nil {
			fmt.Printf("Parse error: %v\n", err)

			continue
		}

		result, err := e.Eval(ast)
		if err != nil {
			fmt.Printf("Evaluation error: %v\n", err)

			continue
		}

		// Display the result
		fmt.Println(result.String())
	}
}

// handleReplCommand processes special REPL commands that start with ':'.
//
// These commands provide meta-functionality for the REPL environment,
// such as displaying help information or executing system-level operations.
//
// Parameters:
//   - cmd: The command string including the ':' prefix
//
// Currently supported commands:
//   - :help, :h - Display available commands and their descriptions
//   - :quit, :q - Exit the REPL (handled in the main loop)
func handleReplCommand(cmd string) {
	switch cmd {
	case ":help", ":h":
		fmt.Println("Available commands:")
		fmt.Println("  :help, :h    Show this help")
		fmt.Println("  :quit, :q    Exit the REPL")
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		fmt.Println("Type :help for available commands")
	}
}
