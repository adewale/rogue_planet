package main

import (
	"flag"
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
  remove-feed <url> Remove a feed from the planet
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
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	feedsFile := fs.String("f", "", "Import feeds from file")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	opts := InitOptions{
		FeedsFile:  *feedsFile,
		ConfigPath: "config.ini",
		Output:     os.Stdout,
	}

	return cmdInit(opts)
}

func runAddFeed() error {
	fs := flag.NewFlagSet("add-feed", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: rp add-feed <url>")
		return fmt.Errorf("missing feed URL argument")
	}

	opts := AddFeedOptions{
		URL:        fs.Arg(0),
		ConfigPath: *configPath,
		Output:     os.Stdout,
	}

	return cmdAddFeed(opts)
}

func runAddAll() error {
	fs := flag.NewFlagSet("add-all", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	feedsFile := fs.String("f", "", "Path to feeds file")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	if *feedsFile == "" {
		fmt.Fprintln(os.Stderr, "Usage: rp add-all -f <feeds-file>")
		return fmt.Errorf("missing feeds file argument")
	}

	opts := AddAllOptions{
		FeedsFile:  *feedsFile,
		ConfigPath: *configPath,
		Output:     os.Stdout,
	}

	return cmdAddAll(opts)
}

func runRemoveFeed() error {
	fs := flag.NewFlagSet("remove-feed", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: rp remove-feed <url>")
		return fmt.Errorf("missing feed URL argument")
	}

	opts := RemoveFeedOptions{
		URL:        fs.Arg(0),
		ConfigPath: *configPath,
		Output:     os.Stdout,
	}

	return cmdRemoveFeed(opts)
}

func runListFeeds() error {
	fs := flag.NewFlagSet("list-feeds", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	opts := ListFeedsOptions{
		ConfigPath: *configPath,
		Output:     os.Stdout,
	}

	return cmdListFeeds(opts)
}

func runStatus() error {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	opts := StatusOptions{
		ConfigPath: *configPath,
		Output:     os.Stdout,
	}

	return cmdStatus(opts)
}

func runUpdate() error {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	verbose := fs.Bool("verbose", false, "Enable verbose logging")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	opts := UpdateOptions{
		ConfigPath: *configPath,
		Verbose:    *verbose,
		Output:     os.Stdout,
	}

	return cmdUpdate(opts)
}

func runFetch() error {
	fs := flag.NewFlagSet("fetch", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	verbose := fs.Bool("verbose", false, "Enable verbose logging")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	opts := FetchOptions{
		ConfigPath: *configPath,
		Verbose:    *verbose,
		Output:     os.Stdout,
	}

	return cmdFetch(opts)
}

func runGenerate() error {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	days := fs.Int("days", 0, "Number of days to include (overrides config)")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	opts := GenerateOptions{
		ConfigPath: *configPath,
		Days:       *days,
		Output:     os.Stdout,
	}

	return cmdGenerate(opts)
}

func runPrune() error {
	fs := flag.NewFlagSet("prune", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	days := fs.Int("days", 90, "Remove entries older than N days")
	dryRun := fs.Bool("dry-run", false, "Show what would be deleted without deleting")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	opts := PruneOptions{
		ConfigPath: *configPath,
		Days:       *days,
		DryRun:     *dryRun,
		Output:     os.Stdout,
	}

	return cmdPrune(opts)
}

func runVerify() error {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	opts := VerifyOptions{
		ConfigPath: *configPath,
		Output:     os.Stdout,
	}

	return cmdVerify(opts)
}

func runImportOPML() error {
	fs := flag.NewFlagSet("import-opml", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	dryRun := fs.Bool("dry-run", false, "Preview feeds without importing")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Usage: rp import-opml <opml-file> [--dry-run]")
		return fmt.Errorf("missing OPML file argument")
	}

	opts := ImportOPMLOptions{
		OPMLFile:   fs.Arg(0),
		ConfigPath: *configPath,
		DryRun:     *dryRun,
		Output:     os.Stdout,
	}

	return cmdImportOPML(opts)
}

func runExportOPML() error {
	fs := flag.NewFlagSet("export-opml", flag.ExitOnError)
	configPath := fs.String("config", "./config.ini", "Path to config file")
	output := fs.String("output", "", "Output file (default: stdout)")
	if err := fs.Parse(os.Args[2:]); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	opts := ExportOPMLOptions{
		ConfigPath: *configPath,
		OutputFile: *output,
		Output:     os.Stdout,
	}

	return cmdExportOPML(opts)
}
