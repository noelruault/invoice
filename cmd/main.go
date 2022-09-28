package main

import (
	"flag"
	"log"
	"os"

	"github.com/noelruault/invoice"
)

type Flags struct {
	Config string
	Input  string
	Output string
}

func parseFlags() *Flags {
	config := flag.String("config", "", "YML file that contains information about the fonts and texts")
	input := flag.String("input", "", "(required) YML file that contains all the data required to model the invoice")
	output := flag.String("output", "", "Full path of the output PDF file")

	flag.Parse()

	return &Flags{
		Config: *config,
		Input:  *input,
		Output: *output,
	}
}

func main() {
	flags := parseFlags()

	if flags.Input == "" {
		flag.Usage()
		os.Exit(1)
	}

	output, err := invoice.Generate(flags.Config, flags.Input, flags.Output)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	log.Printf("Config: %s", flags.Config)
	log.Printf("Input template: %s", flags.Input)
	log.Printf("File created at: %s", *output)
}
