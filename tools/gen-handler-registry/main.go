package main

import (
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const version = "1.0.0"

type Handler struct {
	Name       string
	Type       string // "Command", "Query", "Event", or "Validator"
	ImportPath string
}

type Config struct {
	HandlersDir string
	OutputFile  string
	PackageName string
	ShowVersion bool
	Verbose     bool
	DryRun      bool
}

func main() {
	config := parseFlags()

	if config.ShowVersion {
		fmt.Printf("gen-handler-registry version %s\n", version)
		return
	}

	if config.Verbose {
		log.Printf("Starting handler registry generation...")
	}

	// Get the project root where the app is executed from
	projectRoot, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	if config.Verbose {
		log.Printf("Project root: %s", projectRoot)
	}

	// Find handlers directory
	handlersDir := config.HandlersDir
	if handlersDir == "" {
		handlersDir, err = findHandlersDir(projectRoot)
		if err != nil {
			log.Fatalf("Failed to find handlers directory: %v", err)
		}
	} else {
		// Use specified directory, make it absolute if relative
		if !filepath.IsAbs(handlersDir) {
			handlersDir = filepath.Join(projectRoot, handlersDir)
		}
		if _, err := os.Stat(handlersDir); os.IsNotExist(err) {
			log.Fatalf("Specified handlers directory does not exist: %s", handlersDir)
		}
	}

	if config.Verbose {
		log.Printf("Handlers directory: %s", handlersDir)
	}

	// Get module path
	modulePath, err := getModulePath(projectRoot)
	if err != nil {
		log.Fatalf("Failed to determine module path: %v", err)
	}

	if config.Verbose {
		log.Printf("Module path: %s", modulePath)
	}

	// Determine handlers import path
	relHandlersPath, err := filepath.Rel(projectRoot, handlersDir)
	if err != nil {
		log.Fatalf("Failed to determine relative handlers path: %v", err)
	}

	// Convert Windows path separators to Go import path separator '/'
	relHandlersPath = strings.ReplaceAll(relHandlersPath, "\\", "/")
	handlersImportPath := modulePath
	if relHandlersPath != "." {
		handlersImportPath = modulePath + "/" + relHandlersPath
	}

	if config.Verbose {
		log.Printf("Handlers import path: %s", handlersImportPath)
	}

	// Parse handlers
	handlers, err := parseHandlers(handlersDir, handlersImportPath, config.Verbose)
	if err != nil {
		log.Fatal(err)
	}

	if len(handlers) == 0 {
		log.Fatalf("No handlers found in %s", handlersDir)
	}

	if config.Verbose {
		log.Printf("Found %d handlers:", len(handlers))
		for _, h := range handlers {
			log.Printf("  - %s (%s)", h.Name, h.Type)
		}
	}

	// Determine output file
	outputFile := config.OutputFile
	if outputFile == "" {
		outputFile = filepath.Join(projectRoot, "registry_gen.go")
	} else if !filepath.IsAbs(outputFile) {
		outputFile = filepath.Join(projectRoot, outputFile)
	}

	if config.DryRun {
		log.Printf("DRY RUN: Would generate %d handler registrations in %s", len(handlers), outputFile)
		return
	}

	// Generate the code
	if err := generateCode(handlers, outputFile, handlersImportPath, config.PackageName); err != nil {
		log.Fatal(err)
	}

	log.Printf("Successfully generated handler registry at %s", outputFile)
}

// Discover module path from go.mod file
func getModulePath(projectRoot string) (string, error) {
	modFile, err := os.ReadFile(filepath.Join(projectRoot, "go.mod"))
	if err != nil {
		return "", err
	}

	// Get first line and extract module path
	lines := strings.Split(string(modFile), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}

	return "", errors.New("could not determine module path from go.mod")
}

// findHandlersDir finds a directory named "handlers" (case insensitive) in the project root
func findHandlersDir(projectRoot string) (string, error) {
	var handlersDir string

	err := filepath.WalkDir(projectRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and common build directories
		if d.IsDir() {
			name := d.Name()
			if name[0] == '.' || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
		}

		// Check if it's a directory with a name that matches "handlers" case-insensitively
		if d.IsDir() && strings.ToLower(d.Name()) == "handlers" {
			rel, err := filepath.Rel(projectRoot, path)
			if err == nil && rel != "." && !strings.HasPrefix(rel, "..") {
				handlersDir = path
				// Don't return filepath.SkipAll as it might not be available in all Go versions
				// Continue search in case there are multiple handlers directories
			}
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	if handlersDir == "" {
		return "", errors.New("could not find handlers directory")
	}

	return handlersDir, nil
}

func parseHandlers(dir string, importPath string, verbose bool) ([]Handler, error) {
	var handlers []Handler

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, 0)
	if err != nil {
		return nil, err
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok {
					for _, spec := range genDecl.Specs {
						if typeSpec, ok := spec.(*ast.TypeSpec); ok {
							name := typeSpec.Name.Name
							if strings.HasSuffix(name, "Handler") || strings.HasSuffix(name, "Validator") {
								handlerType := getHandlerType(name)
								if handlerType != "" {
									handlers = append(handlers, Handler{
										Name:       name,
										Type:       handlerType,
										ImportPath: importPath,
									})
								}
							}
						}
					}
				}
			}
		}
	}

	return handlers, nil
}

func getHandlerType(name string) string {
	switch {
	case strings.HasSuffix(name, "CommandHandler"):
		return "Command"
	case strings.HasSuffix(name, "QueryHandler"):
		return "Query"
	case strings.HasSuffix(name, "EventHandler"):
		return "Event"
	case strings.HasSuffix(name, "Validator"):
		return "Validator"
	default:
		return ""
	}
}

const codeTemplate = `// Code generated by gen-handler-registry. DO NOT EDIT.

{{if .PackageName}}package {{.PackageName}}{{else}}package main{{end}}

import (
	"github.com/kmdeveloping/go-cqrs/cqrs"
	"{{ .ImportPath }}"
)

func registerHandlers() {
	// Register handlers
	{{- range .Handlers }}
	{{- if ne .Type "Validator" }}
	cqrs.Register{{ .Type }}Handler(&handlers.{{ .Name }}{})
	{{- else }}
	cqrs.RegisterValidator(&handlers.{{ .Name }}{})
	{{- end }}
	{{- end }}
}
`

func generateCode(handlers []Handler, outputFile string, handlersImportPath string, packageName string) error {
	tmpl, err := template.New("registry").Parse(codeTemplate)
	if err != nil {
		return err
	}

	// Create output file
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	data := struct {
		Handlers    []Handler
		ImportPath  string
		PackageName string
	}{
		Handlers:    handlers,
		ImportPath:  handlersImportPath,
		PackageName: packageName,
	}

	return tmpl.Execute(f, data)
}

func parseFlags() Config {
	config := Config{}

	flag.StringVar(&config.HandlersDir, "dir", "", "Specify the handlers directory (relative or absolute path)")
	flag.StringVar(&config.OutputFile, "output", "", "Specify the output file path (default: registry_gen.go)")
	flag.StringVar(&config.PackageName, "package", "", "Specify the package name for generated code (default: main)")
	flag.BoolVar(&config.ShowVersion, "version", false, "Show version information")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&config.DryRun, "dry-run", false, "Show what would be generated without writing files")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Handler Registry Generator v%s\n", version)
		fmt.Fprintf(os.Stderr, "Automatically generates handler registration code for go-cqrs projects.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s                              # Auto-discover handlers and generate registry_gen.go\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -dir ./internal/handlers     # Use specific handlers directory\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -output handlers_registry.go # Specify output file\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -package handlers            # Generate with 'handlers' package\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -verbose -dry-run            # Preview what would be generated\n", os.Args[0])
	}

	flag.Parse()
	return config
}
