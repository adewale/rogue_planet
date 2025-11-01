package main

import (
	"fmt"
	"os"
)

const version = "0.4.0"

func main() {
	if err := run(); err != nil {
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

	switch command {
	case "init":
		return runInit()
	case "add-feed":
		return runAddFeed()
	case "add-all":
		return runAddAll()
	case "remove-feed":
		return runRemoveFeed()
	case "list-feeds":
		return runListFeeds()
	case "status":
		return runStatus()
	case "update":
		return runUpdate()
	case "fetch":
		return runFetch()
	case "generate":
		return runGenerate()
	case "prune":
		return runPrune()
	case "verify":
		return runVerify()
	case "import-opml":
		return runImportOPML()
	case "export-opml":
		return runExportOPML()
	case "version":
		fmt.Printf("rp version %s\n", version)
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

func runInit() error {
	opts, err := parseInitFlags(os.Args[2:])
	if err != nil {
		return err
	}
	opts.Output = os.Stdout
	return cmdInit(opts)
}

func runAddFeed() error {
	opts, err := parseAddFeedFlags(os.Args[2:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Usage: rp add-feed <url>")
		return err
	}
	opts.Output = os.Stdout
	return cmdAddFeed(opts)
}

func runAddAll() error {
	opts, err := parseAddAllFlags(os.Args[2:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Usage: rp add-all -f <feeds-file>")
		return err
	}
	opts.Output = os.Stdout
	return cmdAddAll(opts)
}

func runRemoveFeed() error {
	opts, err := parseRemoveFeedFlags(os.Args[2:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Usage: rp remove-feed <url> [--force]")
		return err
	}
	opts.Output = os.Stdout
	opts.Input = os.Stdin

	err = cmdRemoveFeed(opts)
	// Check if this is a user cancellation
	if _, ok := err.(*ErrUserCancelled); ok {
		// "Cancelled." already printed by cmdRemoveFeed
		// Exit with code 1 without printing error message
		os.Exit(1)
	}
	return err
}

func runListFeeds() error {
	opts, err := parseListFeedsFlags(os.Args[2:])
	if err != nil {
		return err
	}
	opts.Output = os.Stdout
	return cmdListFeeds(opts)
}

func runStatus() error {
	opts, err := parseStatusFlags(os.Args[2:])
	if err != nil {
		return err
	}
	opts.Output = os.Stdout
	return cmdStatus(opts)
}

func runUpdate() error {
	opts, err := parseUpdateFlags(os.Args[2:])
	if err != nil {
		return err
	}
	opts.Output = os.Stdout
	return cmdUpdate(opts)
}

func runFetch() error {
	opts, err := parseFetchFlags(os.Args[2:])
	if err != nil {
		return err
	}
	opts.Output = os.Stdout
	return cmdFetch(opts)
}

func runGenerate() error {
	opts, err := parseGenerateFlags(os.Args[2:])
	if err != nil {
		return err
	}
	opts.Output = os.Stdout
	return cmdGenerate(opts)
}

func runPrune() error {
	opts, err := parsePruneFlags(os.Args[2:])
	if err != nil {
		return err
	}
	opts.Output = os.Stdout
	return cmdPrune(opts)
}

func runVerify() error {
	opts, err := parseVerifyFlags(os.Args[2:])
	if err != nil {
		return err
	}
	opts.Output = os.Stdout
	return cmdVerify(opts)
}

func runImportOPML() error {
	opts, err := parseImportOPMLFlags(os.Args[2:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Usage: rp import-opml <opml-file> [--dry-run]")
		return err
	}
	opts.Output = os.Stdout
	return cmdImportOPML(opts)
}

func runExportOPML() error {
	opts, err := parseExportOPMLFlags(os.Args[2:])
	if err != nil {
		return err
	}
	opts.Output = os.Stdout
	return cmdExportOPML(opts)
}
