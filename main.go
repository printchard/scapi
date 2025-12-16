package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/printchard/scapi/dsl"
	"github.com/printchard/scapi/generators/golang"
	"github.com/printchard/scapi/generators/ts"
)

const usage = `Usage: scapi [command] [options] <input file>

Commands:
	generate [LANGUAGE] [TARGET]   Generate code from the input file
	validate 	                     Validate the input file

Options:
	-help       Show this help message
	-i          Input file path
	-o          Output file path (optional, defaults to stdout)`

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		return
	}

	command := os.Args[1]

	switch command {
	case "generate":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, usage)
			return
		}
		fs := flag.NewFlagSet("generate", flag.ExitOnError)
		inputFile := fs.String("i", "", "Input file path")
		outputPath := fs.String("o", "", "Output file path")
		outputFile := os.Stdout
		if *outputPath != "" {
			f, err := os.Create(*outputPath)
			if err != nil {
				log.Fatalf("Error creating output file: %s", err)
			}
			defer f.Close()
			outputFile = f
		}
		fs.Parse(os.Args[4:])
		if *inputFile == "" {
			fmt.Fprintln(os.Stderr, "Input file is required")
			return
		}
		lang := os.Args[2]
		target := os.Args[3]
		generateCode(lang, target, *inputFile, outputFile)
	case "validate":
		fs := flag.NewFlagSet("validate", flag.ExitOnError)
		inputFile := fs.String("i", "", "Input file path")
		fs.Parse(os.Args[2:])
		err := validateFile(*inputFile)
		if err != nil {
			log.Fatalf("Invalid File: %s", err)
		}
		log.Println("Validation successful")
	default:
		fmt.Fprintln(os.Stderr, usage)
	}
}

func validateFile(inputPath string) error {
	f, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}

	_, err = dsl.NewTranslatorFromString(string(f))
	return err
}

func generateCode(lang, target, inputPath string, outputFile io.Writer) {
	f, err := os.ReadFile(inputPath)
	if err != nil {
		log.Fatalf("Error reading file: %s", err)
	}

	apiSpec, err := dsl.NewTranslatorFromString(string(f))
	if err != nil {
		log.Fatalf("Error parsing file: %s", err)
	}

	err = apiSpec.Validate()
	if err != nil {
		log.Fatalf("Validation error: %s", err)
	}

	switch lang {
	case "go":
		gen := golang.NewGoGenerator(apiSpec)
		types := gen.GenerateTypeDefs()
		switch target {
		case "server":
			serverCode := gen.GenerateEndpoints()
			outputFile.Write([]byte(types + serverCode))
		case "client":
			clientCode := gen.GenerateClientMethods()
			outputFile.Write([]byte(types + clientCode))
		default:
			log.Fatalf("Unknown target: %s", target)
		}
	case "ts":
		gen := ts.NewTsGenerator(apiSpec)
		types := gen.GenerateTypeDefs()
		switch target {
		case "server":
			log.Fatalf("TypeScript server generation not implemented yet")
		case "client":
			clientCode := gen.GenerateClient()
			outputFile.Write([]byte(types + clientCode))
		default:
			log.Fatalf("Unknown target: %s", target)
		}
	}
}
