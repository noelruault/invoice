package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/noelruault/invoice"
)

type Flags struct {
	Config string
	Input  string
	Output string
}

func parseFlags() *Flags {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)

	config := flag.String("config", basepath+"/../templates/config.yml", "")
	input := flag.String("input", basepath+"/../templates/example.yml", "")
	output := flag.String("output", basepath+"/../tmp/output.pdf", "")

	flag.Parse()

	return &Flags{
		Config: *config,
		Input:  *input,
		Output: *output,
	}
}

func main() {
	flags := parseFlags()

	err := invoice.Generate(flags.Config, flags.Input, flags.Output)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	log.Printf("Config: %s", flags.Config)
	log.Printf("Input template: %s", flags.Input)
	log.Printf("File created at: %s", flags.Output)
}
