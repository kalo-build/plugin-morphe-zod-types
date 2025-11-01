package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kalo-build/plugin-morphe-zod-types/pkg/compile"
)

// CompileConfig represents the configuration passed to the plugin
type CompileConfig struct {
	InputPath  string                 `json:"inputPath"`
	OutputPath string                 `json:"outputPath"`
	Config     map[string]interface{} `json:"config,omitempty"`
	Verbose    bool                   `json:"verbose,omitempty"`
}

// Exit codes
const (
	ExitSuccess         = 0
	ExitCompileFailed   = 1
	ExitMissingConfig   = 3
	ExitInvalidConfig   = 4
	ExitInputPathError  = 12
	ExitOutputPathError = 13
)

// logInfo prints info messages only when verbose mode is enabled
func logInfo(verbose bool, format string, args ...interface{}) {
	if verbose {
		fmt.Fprintf(os.Stdout, format+"\n", args...)
	}
}

func main() {
	// Check command line arguments
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: plugin-morphe-zod-types <config>")
		fmt.Fprintln(os.Stderr, "  config: JSON string with inputPath, outputPath, and optional config parameters")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Example:")
		fmt.Fprintln(os.Stderr, `  plugin-morphe-zod-types '{"inputPath":"./morphe","outputPath":"./output","verbose":true}'`)
		os.Exit(ExitMissingConfig)
	}

	// Parse configuration
	rawConfig := os.Args[1]
	var compileConfig CompileConfig
	if err := json.Unmarshal([]byte(rawConfig), &compileConfig); err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing config JSON:", err)
		fmt.Fprintln(os.Stderr, "Expected format: {\"inputPath\":\"...\",\"outputPath\":\"...\",\"config\":{...},\"verbose\":false}")
		os.Exit(ExitInvalidConfig)
	}

	// Validate required fields
	if compileConfig.InputPath == "" {
		fmt.Fprintln(os.Stderr, "Error: inputPath is required")
		os.Exit(ExitInputPathError)
	}

	if compileConfig.OutputPath == "" {
		fmt.Fprintln(os.Stderr, "Error: outputPath is required")
		os.Exit(ExitOutputPathError)
	}

	// Convert to absolute paths
	inputAbs, err := filepath.Abs(compileConfig.InputPath)
	if err == nil {
		compileConfig.InputPath = inputAbs
	}

	outputAbs, err := filepath.Abs(compileConfig.OutputPath)
	if err == nil {
		compileConfig.OutputPath = outputAbs
	}

	logInfo(compileConfig.Verbose, "Processing Morphe registry from: '%s'", compileConfig.InputPath)
	logInfo(compileConfig.Verbose, "Output Zod schemas to: '%s'", compileConfig.OutputPath)

	// Initialize the compile configuration
	logInfo(compileConfig.Verbose, "Initializing compile configuration...")
	morpheConfig := compile.DefaultMorpheCompileConfig(
		compileConfig.InputPath,
		compileConfig.OutputPath,
	)

	// TODO: Parse format-specific configuration from compileConfig.Config
	// Example:
	// if indent, ok := compileConfig.Config["indentSize"].(float64); ok {
	//     morpheConfig.FormatConfig.IndentSize = int(indent)
	// }

	// Validate configuration
	if err := morpheConfig.Validate(); err != nil {
		fmt.Fprintln(os.Stderr, "Invalid configuration:", err)
		os.Exit(ExitInvalidConfig)
	}

	// Run compilation
	logInfo(compileConfig.Verbose, "Starting compilation process...")
	if err := compile.MorpheToZod(morpheConfig); err != nil {
		fmt.Fprintln(os.Stderr, "Compilation failed:", err)
		os.Exit(ExitCompileFailed)
	}

	logInfo(compileConfig.Verbose, "Compilation completed successfully")
	os.Exit(ExitSuccess)
}
