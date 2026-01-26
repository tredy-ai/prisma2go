package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tredy-ai/prisma2go/internal/config"
	"github.com/tredy-ai/prisma2go/internal/generator"
)

const version = "0.1.0"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Define flags
	var (
		inputFlag      string
		outputFlag     string
		packageFlag    string
		configFlag     string
		noJSONTags     bool
		noDBTags       bool
		noEnumer       bool
		showVersion    bool
		showHelp       bool
	)

	flag.StringVar(&inputFlag, "i", "", "Input DMMF JSON file (or - for stdin)")
	flag.StringVar(&inputFlag, "input", "", "Input DMMF JSON file (or - for stdin)")
	flag.StringVar(&outputFlag, "o", "", "Output Go file (or - for stdout)")
	flag.StringVar(&outputFlag, "output", "", "Output Go file (or - for stdout)")
	flag.StringVar(&packageFlag, "p", "", "Go package name")
	flag.StringVar(&packageFlag, "package", "", "Go package name")
	flag.StringVar(&configFlag, "c", "", "Config file path")
	flag.StringVar(&configFlag, "config", "", "Config file path")
	flag.BoolVar(&noJSONTags, "no-json-tags", false, "Disable json struct tags")
	flag.BoolVar(&noDBTags, "no-db-tags", false, "Disable db struct tags")
	flag.BoolVar(&noEnumer, "no-enumer", false, "Skip automatic enumer invocation")
	flag.BoolVar(&showVersion, "v", false, "Show version")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.BoolVar(&showHelp, "h", false, "Show help")
	flag.BoolVar(&showHelp, "help", false, "Show help")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `prisma2go - Generate Go structs from Prisma DMMF

Usage:
  prisma2go generate [flags]

Examples:
  # With config file
  prisma2go generate

  # With explicit input/output
  prisma2go generate -i dmmf.json -o models.go -p mypackage

  # From stdin to stdout
  cat dmmf.json | prisma2go generate -p models

Flags:
`)
		flag.PrintDefaults()
	}

	// Handle subcommand
	if len(os.Args) < 2 {
		flag.Usage()
		return nil
	}

	switch os.Args[1] {
	case "generate":
		// Parse flags after "generate"
		flag.CommandLine.Parse(os.Args[2:])
	case "version", "-v", "--version":
		fmt.Printf("prisma2go %s\n", version)
		return nil
	case "help", "-h", "--help":
		flag.Usage()
		return nil
	default:
		// Try parsing as flags directly for backwards compatibility
		flag.CommandLine.Parse(os.Args[1:])
	}

	if showVersion {
		fmt.Printf("prisma2go %s\n", version)
		return nil
	}

	if showHelp {
		flag.Usage()
		return nil
	}

	// Load config
	cfg, err := config.LoadConfig(configFlag)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Apply flag overrides
	if inputFlag != "" {
		cfg.Input = inputFlag
	}
	if outputFlag != "" {
		cfg.Output = outputFlag
	}
	if packageFlag != "" {
		cfg.Package = packageFlag
	}
	if noJSONTags {
		cfg.JSONTags = false
	}
	if noDBTags {
		cfg.DBTags = false
	}
	if noEnumer {
		cfg.NoEnumer = true
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return err
	}

	// Run generator
	gen := generator.New(cfg)
	return gen.Run()
}
