package generator

import (
	"encoding/json"
	"fmt"
	"go/format"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tredy-ai/prisma2go/internal/config"
	"github.com/tredy-ai/prisma2go/internal/dmmf"
)

// Generator orchestrates the code generation process.
type Generator struct {
	cfg *config.Config
}

// New creates a new Generator with the given configuration.
func New(cfg *config.Config) *Generator {
	return &Generator{cfg: cfg}
}

// Run executes the full generation process.
func (g *Generator) Run() error {
	// Read input
	datamodel, err := g.readInput()
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	// Generate code
	code, err := Render(datamodel, g.cfg)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// Format code
	formatted, err := format.Source(code)
	if err != nil {
		// If formatting fails, output unformatted code for debugging
		fmt.Fprintf(os.Stderr, "warning: generated code failed gofmt: %v\n", err)
		formatted = code
	}

	// Write output
	if err := g.writeOutput(formatted); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	// Run enumer for each enum (if not disabled)
	if !g.cfg.NoEnumer && len(datamodel.Enums) > 0 {
		g.runEnumer(datamodel.Enums)
	}

	return nil
}

// readInput reads and parses the DMMF JSON input.
func (g *Generator) readInput() (*dmmf.Datamodel, error) {
	var reader io.Reader

	if g.cfg.Input == "-" || g.cfg.Input == "" {
		// Check if stdin has data
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			reader = os.Stdin
		} else if g.cfg.Input == "" {
			return nil, fmt.Errorf("no input specified and stdin is empty")
		}
	}

	if reader == nil {
		f, err := os.Open(g.cfg.Input)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		reader = f
	}

	// Try parsing as full DMMF first (with datamodel wrapper)
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// Try full DMMF structure
	var fullDMMF struct {
		Datamodel *dmmf.Datamodel `json:"datamodel"`
	}
	if err := json.Unmarshal(data, &fullDMMF); err == nil && fullDMMF.Datamodel != nil {
		return fullDMMF.Datamodel, nil
	}

	// Try direct datamodel structure
	var datamodel dmmf.Datamodel
	if err := json.Unmarshal(data, &datamodel); err != nil {
		return nil, fmt.Errorf("failed to parse DMMF JSON: %w", err)
	}

	return &datamodel, nil
}

// writeOutput writes the generated code to the configured output.
func (g *Generator) writeOutput(code []byte) error {
	if g.cfg.Output == "" || g.cfg.Output == "-" {
		_, err := os.Stdout.Write(code)
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(g.cfg.Output)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return os.WriteFile(g.cfg.Output, code, 0644)
}

// runEnumer invokes the enumer tool for each enum type.
func (g *Generator) runEnumer(enums []dmmf.Enum) {
	// Check if enumer is available
	if _, err := exec.LookPath("enumer"); err != nil {
		fmt.Fprintf(os.Stderr, "warning: enumer not found in PATH, skipping enum generation\n")
		fmt.Fprintf(os.Stderr, "  install with: go install github.com/dmarkham/enumer@latest\n")
		return
	}

	outputDir := g.cfg.OutputDir()

	for _, enum := range enums {
		outputFile := filepath.Join(outputDir, strings.ToLower(enum.Name)+"_enumer.go")

		args := []string{
			"-type=" + enum.Name,
			"-json",
			"-sql",
			"-text",
			"-trimprefix=" + enum.Name,
			"-output=" + outputFile,
		}

		// If we have an output file, run enumer in that directory
		cmd := exec.Command("enumer", args...)
		if g.cfg.Output != "" && g.cfg.Output != "-" {
			cmd.Dir = outputDir
			// Point to the generated file
			cmd.Args = append(cmd.Args, filepath.Base(g.cfg.Output))
		}

		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: enumer failed for %s: %v\n", enum.Name, err)
			if len(output) > 0 {
				fmt.Fprintf(os.Stderr, "  %s\n", string(output))
			}
		}
	}
}
