// Package cli provides command-line interface handling for h3repl.
package cli

import (
	"flag"
	"fmt"
	"os"
)

const version = "0.1.0"

// Options holds the parsed command-line options
type Options struct {
	ConfigPath string
	Verbose    bool
	DryRun     bool
	Version    bool
}

// Parse parses command-line arguments and returns options
func Parse() (*Options, error) {
	opts := &Options{}

	flag.StringVar(&opts.ConfigPath, "config", "", "path to TOML config file (required)")
	flag.BoolVar(&opts.Verbose, "v", false, "enable debug logging")
	flag.BoolVar(&opts.DryRun, "dry-run", false, "simulate actions, write nothing")
	flag.BoolVar(&opts.Version, "version", false, "print tool version and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "h3repl - S3 replication tool for homelabs\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if opts.Version {
		fmt.Printf("h3repl version %s\n", version)
		os.Exit(0)
	}

	if opts.ConfigPath == "" {
		return nil, fmt.Errorf("--config flag is required")
	}

	return opts, nil
}
