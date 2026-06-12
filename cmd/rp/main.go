package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/adewale/rogue_planet/pkg/version"
)

func main() {
	if err := run(); err != nil {
		// ErrUserCancelled is a silent exit (message already printed)
		if _, ok := err.(*ErrUserCancelled); ok {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		printUsage()
		return fmt.Errorf("no command specified")
	}

	command := os.Args[1]

	// Create context with signal handling for long-running commands
	// This enables graceful cancellation with Ctrl+C (SIGINT) or kill (SIGTERM)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	switch command {
	case "init":
		return runInit(ctx)
	case "add-feed":
		return runAddFeed(ctx)
	case "add-all":
		return runAddAll(ctx)
	case "remove-feed":
		return runRemoveFeed(ctx)
	case "list-feeds":
		return runListFeeds(ctx)
	case "status":
		return runStatus(ctx)
	case "update":
		return runUpdate(ctx)
	case "fetch":
		return runFetch(ctx)
	case "generate":
		return runGenerate(ctx)
	case "prune":
		return runPrune(ctx)
	case "verify":
		return runVerify(ctx)
	case "import-opml":
		return runImportOPML(ctx)
	case "export-opml":
		return runExportOPML(ctx)
	case "version":
		fmt.Printf("rp version %s\n", version.Version)
		return nil
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		return fmt.Errorf("unknown command: %s", command)
	}
}

func printUsage() {
	fmt.Print(`Rogue Planet - Modern feed aggregator
Supports RSS, Atom, and JSON Feed formats

Usage:
  rp <command> [flags]

Commands:
  init [-f FILE]    Initialize a new planet in the current directory
  add-feed <url>    Add a feed to the planet
  add-all -f FILE   Add multiple feeds from a file
  remove-feed <url> Remove a feed from the planet (interactive confirmation)
  list-feeds        List all configured feeds
  status            Show planet status (feed and entry counts)
  update            Fetch all feeds and regenerate site
  fetch             Fetch all feeds without generating
  generate          Generate site without fetching
  prune             Remove old entries from database
  verify            Validate configuration and environment
  import-opml FILE  Import feeds from OPML file
  export-opml       Export feeds to OPML format
  version           Show version information
  help              Show this help message

Init Flags:
  -f FILE           Import feeds from file (one URL per line)

Add-All Flags:
  -f FILE           Path to feeds file (one URL per line)

Remove-Feed Flags:
  --force           Skip confirmation prompt (for scripting)

Import-OPML Flags:
  --dry-run         Preview feeds without importing

Export-OPML Flags:
  --output FILE     Output file (default: stdout)

Global Flags:
  --config <path>   Path to config file (default: ./config.ini)
  --verbose         Enable verbose logging
  --quiet           Only show errors

Examples:
  rp init
  rp init -f feeds.txt
  rp add-feed https://blog.golang.org/feed.atom
  rp add-feed https://username.micro.blog/feed.json
  rp add-all -f feeds.txt
  rp remove-feed https://example.com/feed.xml
  rp remove-feed https://example.com/feed.xml --force
  rp list-feeds
  rp status
  rp update
  rp generate --days 14
  rp prune --days 90
  rp import-opml feeds.opml
  rp import-opml feeds.opml --dry-run
  rp export-opml --output feeds.opml

`)
}

func runInit(ctx context.Context) error {
	opts, err := parseInitFlags(os.Args[2:])
	if err != nil {
		return err
	}
	opts.Output = os.Stdout
	return cmdInit(ctx, opts)
}

func runAddFeed(ctx context.Context) error {
	opts, err := parseAddFeedFlags(os.Args[2:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Usage: rp add-feed <url>")
		return err
	}
	opts.Output = os.Stdout
	return cmdAddFeed(ctx, opts)
}

func runAddAll(ctx context.Context) error {
	opts, err := parseAddAllFlags(os.Args[2:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Usage: rp add-all -f <feeds-file>")
		return err
	}
	opts.Output = os.Stdout
	return cmdAddAll(ctx, opts)
}

func runRemoveFeed(ctx context.Context) error {
	opts, err := parseRemoveFeedFlags(os.Args[2:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Usage: rp remove-feed <url> [--force]")
		return err
	}
	opts.Output = os.Stdout
	opts.Input = os.Stdin
	return cmdRemoveFeed(ctx, opts)
}

func runListFeeds(ctx context.Context) error {
	opts, err := parseListFeedsFlags(os.Args[2:])
	if err != nil {
		return err
	}
	opts.Output = os.Stdout
	return cmdListFeeds(ctx, opts)
}

func runStatus(ctx context.Context) error {
	opts, err := parseStatusFlags(os.Args[2:])
	if err != nil {
		return err
	}
	opts.Output = os.Stdout
	return cmdStatus(ctx, opts)
}

func runUpdate(ctx context.Context) error {
	opts, err := parseUpdateFlags(os.Args[2:])
	if err != nil {
		return err
	}
	opts.Output = os.Stdout
	return cmdUpdate(ctx, opts)
}

func runFetch(ctx context.Context) error {
	opts, err := parseFetchFlags(os.Args[2:])
	if err != nil {
		return err
	}
	opts.Output = os.Stdout
	return cmdFetch(ctx, opts)
}

func runGenerate(ctx context.Context) error {
	opts, err := parseGenerateFlags(os.Args[2:])
	if err != nil {
		return err
	}
	opts.Output = os.Stdout
	return cmdGenerate(ctx, opts)
}

func runPrune(ctx context.Context) error {
	opts, err := parsePruneFlags(os.Args[2:])
	if err != nil {
		return err
	}
	opts.Output = os.Stdout
	return cmdPrune(ctx, opts)
}

func runVerify(ctx context.Context) error {
	opts, err := parseVerifyFlags(os.Args[2:])
	if err != nil {
		return err
	}
	opts.Output = os.Stdout
	return cmdVerify(ctx, opts)
}

func runImportOPML(ctx context.Context) error {
	opts, err := parseImportOPMLFlags(os.Args[2:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Usage: rp import-opml <opml-file> [--dry-run]")
		return err
	}
	opts.Output = os.Stdout
	return cmdImportOPML(ctx, opts)
}

func runExportOPML(ctx context.Context) error {
	opts, err := parseExportOPMLFlags(os.Args[2:])
	if err != nil {
		return err
	}
	opts.Output = os.Stdout
	return cmdExportOPML(ctx, opts)
}
